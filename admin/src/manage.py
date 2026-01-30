#!/usr/bin/env python
"""Django's command-line utility for administrative tasks."""

import logging
import os
import sys

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from admin.otel_logging import CustomLogRecord

# Initialize OpenTelemetry with OTLP exporter
provider = TracerProvider()
otlp_exporter = OTLPSpanExporter(
    endpoint=os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://grafana-tempo.grafana:4317"),
    insecure=True,
)
processor = BatchSpanProcessor(otlp_exporter)
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

logging.setLogRecordFactory(CustomLogRecord)


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
