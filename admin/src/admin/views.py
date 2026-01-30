"""Define the views for the admin app."""

from django.conf import settings
from django.contrib.auth import logout
from django.http import HttpResponse
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
