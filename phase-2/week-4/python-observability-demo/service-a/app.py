import json
import logging
import os
import time
from contextlib import asynccontextmanager

import httpx
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import Response
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Histogram, generate_latest

SERVICE_NAME = os.getenv("SERVICE_NAME", "service-a")
SERVICE_B_URL = os.getenv("SERVICE_B_URL", "http://localhost:8001")
OTLP_ENDPOINT = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")


class JsonFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        span = trace.get_current_span()
        ctx = span.get_span_context()
        payload = {
            "timestamp": self.formatTime(record, "%Y-%m-%dT%H:%M:%S"),
            "level": record.levelname,
            "service": SERVICE_NAME,
            "message": record.getMessage(),
            "trace_id": format(ctx.trace_id, "032x") if ctx.is_valid else None,
            "span_id": format(ctx.span_id, "016x") if ctx.is_valid else None,
        }
        return json.dumps(payload)


handler = logging.StreamHandler()
handler.setFormatter(JsonFormatter())
logger = logging.getLogger(SERVICE_NAME)
logger.handlers.clear()
logger.addHandler(handler)
logger.setLevel(logging.INFO)
logger.propagate = False

resource = Resource.create({"service.name": SERVICE_NAME})
provider = TracerProvider(resource=resource)
provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=OTLP_ENDPOINT, insecure=True)))
trace.set_tracer_provider(provider)

REQUESTS = Counter(
    "http_requests_total",
    "Total HTTP requests",
    ["service", "method", "path", "status"],
)
REQUEST_LATENCY = Histogram(
    "http_request_duration_seconds",
    "HTTP request latency in seconds",
    ["service", "method", "path"],
)
DOWNSTREAM_CALLS = Counter(
    "downstream_calls_total",
    "Calls from service-a to service-b",
    ["result"],
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("service starting")
    yield
    logger.info("service stopping")
    provider.shutdown()


app = FastAPI(title="Service A", lifespan=lifespan)


@app.middleware("http")
async def metrics_and_logs(request: Request, call_next):
    if request.url.path == "/metrics":
        return await call_next(request)

    start = time.perf_counter()
    status = 500
    try:
        response = await call_next(request)
        status = response.status_code
        return response
    finally:
        duration = time.perf_counter() - start
        REQUESTS.labels(SERVICE_NAME, request.method, request.url.path, str(status)).inc()
        REQUEST_LATENCY.labels(SERVICE_NAME, request.method, request.url.path).observe(duration)
        logger.info(f"{request.method} {request.url.path} status={status} duration={duration:.4f}s")


@app.get("/")
async def root():
    return {"service": SERVICE_NAME, "message": "Service A is running"}


@app.get("/api/demo")
async def demo():
    logger.info("calling service-b")
    try:
        async with httpx.AsyncClient(timeout=3.0) as client:
            response = await client.get(f"{SERVICE_B_URL}/api/work")
            response.raise_for_status()
        DOWNSTREAM_CALLS.labels("success").inc()
        logger.info("service-b call succeeded")
        return {
            "service": SERVICE_NAME,
            "message": "Distributed request completed",
            "downstream": response.json(),
        }
    except httpx.HTTPError as exc:
        DOWNSTREAM_CALLS.labels("error").inc()
        logger.exception("service-b call failed")
        raise HTTPException(status_code=502, detail=str(exc)) from exc


@app.get("/metrics")
async def metrics():
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)


FastAPIInstrumentor.instrument_app(app)
HTTPXClientInstrumentor().instrument()
