"""Structured request logging middleware.

Trace context (trace_id/span_id) is injected by the custom LogRecord factory.
Do not duplicate it in message payload.
"""

import logging
import time
from http import HTTPStatus

logger = logging.getLogger(__name__)

# HTTP status code thresholds for log levels
HTTP_SERVER_ERROR = HTTPStatus.INTERNAL_SERVER_ERROR  # 500
HTTP_CLIENT_ERROR = HTTPStatus.BAD_REQUEST  # 400
SKIP_LOG_PATH_PREFIXES = ("/healthz", "/health", "/ping")


class JsonLogMiddleware:
    """Structured logging middleware with trace context.

    Logs each request/response with:
    - HTTP method, path, status code
    - Response time in milliseconds
    - optional user_id, query_string
    """

    def __init__(self, get_response):
        """Initialize middleware."""
        self.get_response = get_response

    def __call__(self, request):
        """Log request and response data with trace context."""
        start_time = time.perf_counter()

        response = self.get_response(request)

        # Skip noisy probe endpoints.
        if request.path.startswith(SKIP_LOG_PATH_PREFIXES):
            return response

        # Calculate response time
        duration_ms = (time.perf_counter() - start_time) * 1000

        # Build structured log data.
        # trace_id/span_id are injected by the custom LogRecord factory,
        # so we don't duplicate them inside message payload.
        log_data = {
            "method": request.method,
            "path": request.path,
            "status_code": response.status_code,
            "duration_ms": round(duration_ms, 2),
        }

        # Add user info if authenticated
        if hasattr(request, "user") and request.user.is_authenticated:
            log_data["user_id"] = str(request.user.id) if hasattr(request.user, "id") else None

        # Add query string if present
        if request.META.get("QUERY_STRING"):
            log_data["query_string"] = request.META["QUERY_STRING"]

        # Log at appropriate level based on status code.
        # `extra` keeps fields structured in JSON formatter output.
        if response.status_code >= HTTP_SERVER_ERROR:
            logger.error("http_request", extra=log_data)
        elif response.status_code >= HTTP_CLIENT_ERROR:
            logger.warning("http_request", extra=log_data)
        else:
            logger.info("http_request", extra=log_data)

        return response
