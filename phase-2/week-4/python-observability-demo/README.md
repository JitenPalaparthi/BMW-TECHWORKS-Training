# Python Observability Demo

Two FastAPI applications demonstrating the three core observability signals:

- **Metrics**: Prometheus
- **Traces**: OpenTelemetry -> Jaeger
- **Logs**: structured JSON logs with `trace_id` and `span_id`

## Architecture

```text
Client
  |
  v
Service A :8000  ----HTTP---->  Service B :8001
  |                              |
  +---------- metrics -----------+----> Prometheus :9090
  |                              |
  +---------- OTLP traces -------+----> Jaeger :16686

Both services write JSON logs to stdout.
```

## Start everything

```bash
docker compose up --build
```

## Generate a distributed trace

```bash
curl http://localhost:8000/api/demo
```

Repeat it a few times:

```bash
for i in {1..10}; do curl -s http://localhost:8000/api/demo; echo; done
```

## URLs

| Component | URL |
|---|---|
| Service A | http://localhost:8000 |
| Service B | http://localhost:8001 |
| Service A metrics | http://localhost:8000/metrics |
| Service B metrics | http://localhost:8001/metrics |
| Prometheus | http://localhost:9090 |
| Jaeger UI | http://localhost:16686 |

## Prometheus queries

Total requests by service:

```promql
sum by (service) (http_requests_total)
```

Request rate over five minutes:

```promql
sum by (service) (rate(http_requests_total[5m]))
```

95th percentile HTTP latency:

```promql
histogram_quantile(
  0.95,
  sum by (le, service) (rate(http_request_duration_seconds_bucket[5m]))
)
```

Downstream calls from Service A:

```promql
sum by (result) (downstream_calls_total)
```

## Jaeger tracing

1. Call `http://localhost:8000/api/demo`.
2. Open Jaeger at `http://localhost:16686`.
3. Select `service-a` from the Service dropdown.
4. Click **Find Traces**.
5. Open a trace. You should see Service A calling Service B, including the custom `business-work` span.

OpenTelemetry propagates the W3C trace context across the HTTP call automatically through the HTTPX instrumentation.

## Logs

Show logs from both applications:

```bash
docker compose logs -f service-a service-b
```

Example application log:

```json
{"timestamp":"2026-09-10T18:00:00","level":"INFO","service":"service-a","message":"calling service-b","trace_id":"...","span_id":"..."}
```

The `trace_id` lets you correlate an application log with the same request in Jaeger.

## Stop

```bash
docker compose down
```

Remove containers and volumes:

```bash
docker compose down -v
```
