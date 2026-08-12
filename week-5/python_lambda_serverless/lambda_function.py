import json
import logging
import os
import uuid
from datetime import datetime, timezone
from urllib.parse import parse_qs

LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
APP_NAME = os.getenv("APP_NAME", "python-serverless-demo")

logger = logging.getLogger()
logger.setLevel(LOG_LEVEL)


def _json_response(status_code, body, headers=None):
    default_headers = {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers": "content-type,authorization",
        "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
    }
    if headers:
        default_headers.update(headers)

    return {
        "statusCode": status_code,
        "headers": default_headers,
        "body": json.dumps(body),
    }


def _get_method(event):
    # API Gateway HTTP API v2 / Lambda Function URL
    request_context = event.get("requestContext", {})
    http = request_context.get("http", {})
    if http.get("method"):
        return http["method"].upper()

    # API Gateway REST API v1
    if event.get("httpMethod"):
        return event["httpMethod"].upper()

    return "GET"


def _get_path(event):
    # HTTP API v2 / Function URL
    if event.get("rawPath"):
        return event["rawPath"]

    # REST API v1
    if event.get("path"):
        return event["path"]

    return "/"


def _get_query_params(event):
    params = event.get("queryStringParameters")
    if params:
        return params

    raw_query = event.get("rawQueryString", "")
    if raw_query:
        parsed = parse_qs(raw_query)
        return {k: v[-1] for k, v in parsed.items()}

    return {}


def _get_json_body(event):
    body = event.get("body")
    if not body:
        return {}

    if isinstance(body, dict):
        return body

    try:
        return json.loads(body)
    except json.JSONDecodeError as exc:
        raise ValueError("Request body must contain valid JSON") from exc


def lambda_handler(event, context):
    request_id = getattr(context, "aws_request_id", None) or str(uuid.uuid4())
    method = _get_method(event)
    path = _get_path(event)

    logger.info(
        json.dumps({
            "message": "request_received",
            "request_id": request_id,
            "method": method,
            "path": path,
        })
    )

    try:
        if method == "OPTIONS":
            return _json_response(204, {})

        if method == "GET" and path in ("/", "/health"):
            return _json_response(
                200,
                {
                    "status": "ok",
                    "application": APP_NAME,
                    "requestId": request_id,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                },
            )

        if method == "GET" and path == "/hello":
            query = _get_query_params(event)
            name = query.get("name", "Guest")

            return _json_response(
                200,
                {
                    "message": f"Hello {name}!",
                    "application": APP_NAME,
                    "requestId": request_id,
                },
            )

        if method == "POST" and path == "/echo":
            body = _get_json_body(event)

            return _json_response(
                200,
                {
                    "received": body,
                    "requestId": request_id,
                },
            )

        return _json_response(
            404,
            {
                "error": "Not Found",
                "method": method,
                "path": path,
                "requestId": request_id,
            },
        )

    except ValueError as exc:
        logger.warning(
            json.dumps({
                "message": "bad_request",
                "request_id": request_id,
                "error": str(exc),
            })
        )
        return _json_response(
            400,
            {
                "error": str(exc),
                "requestId": request_id,
            },
        )

    except Exception:
        logger.exception("Unhandled exception")
        return _json_response(
            500,
            {
                "error": "Internal Server Error",
                "requestId": request_id,
            },
        )
