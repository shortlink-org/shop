"""OpenTelemetry bootstrap for admin service.

Used by both CLI (manage.py) and WSGI runtime (gunicorn).
"""

from __future__ import annotations

import logging
import os

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider as SDKTracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from admin.otel_logging import CustomLogRecord


def _env_bool(name: str, default: bool) -> bool:
    raw = os.environ.get(name)
    if raw is None:
        return default
    return raw.strip().lower() in {"1", "true", "yes", "on"}


def configure_opentelemetry() -> None:
    """Configure tracing/export and trace-aware log records."""
    logging.setLogRecordFactory(CustomLogRecord)

    # If SDK provider already exists (e.g. initialized earlier), keep it.
    if isinstance(trace.get_tracer_provider(), SDKTracerProvider):
        return

    service_name = os.environ.get("OTEL_SERVICE_NAME", "shop-admin")
    endpoint = os.environ.get(
        "OTEL_EXPORTER_OTLP_ENDPOINT",
        "http://grafana-tempo.grafana:4317",
    )
    insecure = _env_bool("OTEL_EXPORTER_OTLP_INSECURE", True)

    provider = SDKTracerProvider(
        resource=Resource.create({"service.name": service_name})
    )
    provider.add_span_processor(
        BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=insecure))
    )
    trace.set_tracer_provider(provider)

