"""Structured logging middleware with OpenTelemetry trace context."""

import logging
import time
from http import HTTPStatus

from opentelemetry import trace

logger = logging.getLogger(__name__)

# HTTP status code thresholds for log levels
HTTP_SERVER_ERROR = HTTPStatus.INTERNAL_SERVER_ERROR  # 500
HTTP_CLIENT_ERROR = HTTPStatus.BAD_REQUEST  # 400


def get_trace_id() -> str | None:
    """Get current OpenTelemetry trace ID as hex string."""
    current_span = trace.get_current_span()
    if current_span is None:
        return None

    span_context = current_span.get_span_context()
    if not span_context.is_valid:
        return None

    # Convert trace_id to 32-char hex string (standard format)
    return format(span_context.trace_id, "032x")


def get_span_id() -> str | None:
    """Get current OpenTelemetry span ID as hex string."""
    current_span = trace.get_current_span()
    if current_span is None:
        return None

    span_context = current_span.get_span_context()
    if not span_context.is_valid:
        return None

    # Convert span_id to 16-char hex string (standard format)
    return format(span_context.span_id, "016x")


class JsonLogMiddleware:
    """Structured logging middleware with trace context.

    Logs each request/response with:
    - HTTP method, path, status code
    - Response time in milliseconds
    - OpenTelemetry trace_id and span_id for distributed tracing
    """

    def __init__(self, get_response):
        """Initialize middleware."""
        self.get_response = get_response

    def __call__(self, request):
        """Log request and response data with trace context."""
        start_time = time.perf_counter()

        response = self.get_response(request)

        # Calculate response time
        duration_ms = (time.perf_counter() - start_time) * 1000

        # Build structured log data
        log_data = {
            "method": request.method,
            "path": request.path,
            "status_code": response.status_code,
            "duration_ms": round(duration_ms, 2),
            "trace_id": get_trace_id(),
            "span_id": get_span_id(),
        }

        # Add user info if authenticated
        if hasattr(request, "user") and request.user.is_authenticated:
            log_data["user_id"] = str(request.user.id) if hasattr(request.user, "id") else None

        # Add query string if present
        if request.META.get("QUERY_STRING"):
            log_data["query_string"] = request.META["QUERY_STRING"]

        # Log at appropriate level based on status code
        if response.status_code >= HTTP_SERVER_ERROR:
            logger.error(log_data)
        elif response.status_code >= HTTP_CLIENT_ERROR:
            logger.warning(log_data)
        else:
            logger.info(log_data)

        return response
