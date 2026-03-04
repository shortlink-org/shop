"""Define the views for the admin app."""

import logging

from django.conf import settings
from django.contrib.auth import logout
from django.core.cache import cache
from django.db import connections
from django.db.migrations.executor import MigrationExecutor
from django.http import HttpResponse, JsonResponse
from django.shortcuts import redirect
from opentelemetry import trace

tracer = trace.get_tracer(__name__)
logger = logging.getLogger(__name__)


def hello(request):
    """Return a simple hello world response."""
    # Create a custom span
    with tracer.start_as_current_span("hello"):
        return HttpResponse("Hello, World!")


def logout_view(request):
    """Logout view that supports GET requests."""
    logout(request)
    return redirect(settings.LOGIN_URL)


def healthz(_request):
    """Synchronous health endpoint for WSGI/gunicorn runtime."""
    checks: dict[str, str] = {}
    status_code = 200

    try:
        with connections["default"].cursor() as cursor:
            cursor.execute("SELECT 1")
        checks["database"] = "ok"
    except Exception as exc:
        checks["database"] = f"error: {exc}"
        status_code = 500
        logger.error(
            "health_dependency_unavailable",
            extra={"dependency": "database", "error": str(exc), "health_path": "/healthz/"},
        )

    # Cache is non-critical for API availability; keep it observable but
    # do not fail readiness/liveness when Redis is temporarily unavailable.
    try:
        cache_key = "__healthz__"
        cache_value = "ok"
        cache.set(cache_key, cache_value, timeout=5)
        if cache.get(cache_key) == cache_value:
            checks["cache"] = "ok"
        else:
            checks["cache"] = "degraded: mismatch"
            logger.warning(
                "health_dependency_degraded",
                extra={
                    "dependency": "cache",
                    "error": "cache readback mismatch",
                    "health_path": "/healthz/",
                },
            )
    except Exception as exc:
        checks["cache"] = f"degraded: {exc}"
        logger.warning(
            "health_dependency_degraded",
            extra={"dependency": "cache", "error": str(exc), "health_path": "/healthz/"},
        )

    # Migration state is informational for probes.
    # Startup/readiness should not flap because of migration introspection.
    try:
        executor = MigrationExecutor(connections["default"])
        pending = executor.migration_plan(executor.loader.graph.leaf_nodes())
        if not pending:
            checks["migrations"] = "ok"
        else:
            checks["migrations"] = "degraded: pending migrations"
            logger.warning(
                "health_dependency_degraded",
                extra={
                    "dependency": "migrations",
                    "error": "pending migrations",
                    "pending_count": len(pending),
                    "health_path": "/healthz/",
                },
            )
    except Exception as exc:
        checks["migrations"] = f"degraded: {exc}"
        logger.warning(
            "health_dependency_degraded",
            extra={"dependency": "migrations", "error": str(exc), "health_path": "/healthz/"},
        )

    if status_code == 200:
        return HttpResponse("ok", content_type="text/plain; charset=utf-8")

    logger.error("healthz_failed", extra={"checks": checks, "health_path": "/healthz/"})
    return JsonResponse({"status": "error", "checks": checks}, status=status_code)
