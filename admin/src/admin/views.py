"""Define the views for the admin app."""

from django.conf import settings
from django.contrib.auth import logout
from django.core.cache import cache
from django.db import connections
from django.db.migrations.executor import MigrationExecutor
from django.http import HttpResponse, JsonResponse
from django.shortcuts import redirect
from opentelemetry import trace

tracer = trace.get_tracer(__name__)


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

    try:
        cache_key = "__healthz__"
        cache_value = "ok"
        cache.set(cache_key, cache_value, timeout=5)
        checks["cache"] = "ok" if cache.get(cache_key) == cache_value else "error: mismatch"
        if checks["cache"] != "ok":
            status_code = 500
    except Exception as exc:
        checks["cache"] = f"error: {exc}"
        status_code = 500

    try:
        executor = MigrationExecutor(connections["default"])
        pending = executor.migration_plan(executor.loader.graph.leaf_nodes())
        checks["migrations"] = "ok" if not pending else "error: pending migrations"
        if pending:
            status_code = 500
    except Exception as exc:
        checks["migrations"] = f"error: {exc}"
        status_code = 500

    if status_code == 200:
        return HttpResponse("ok", content_type="text/plain; charset=utf-8")

    return JsonResponse({"status": "error", "checks": checks}, status=status_code)
