#!/usr/bin/env python
"""Django's command-line utility for administrative tasks."""

import os
import sys

from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor

from admin.otel_setup import configure_opentelemetry

configure_opentelemetry()


def main():
    """Run administrative tasks."""
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "admin.settings")

    try:
        from django.core.management import execute_from_command_line
    except ImportError as exc:
        raise ImportError(
            "Couldn't import Django. Are you sure it's installed and "
            "available on your PYTHONPATH environment variable? Did you "
            "forget to activate a virtual environment?"
        ) from exc
    # Add --noreload only for runserver command
    args = sys.argv
    if len(args) > 1 and args[1] == "runserver" and "--noreload" not in args:
        args = [*args, "--noreload"]

    DjangoInstrumentor().instrument()
    RequestsInstrumentor().instrument()

    execute_from_command_line(args)


if __name__ == "__main__":
    main()
