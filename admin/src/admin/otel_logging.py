"""Custom LogRecord with OpenTelemetry trace context.

This module provides a custom LogRecord class that automatically adds
OpenTelemetry trace_id and span_id to all log records for distributed tracing.
"""

import logging

from opentelemetry import trace


class CustomLogRecord(logging.LogRecord):
    """Custom LogRecord that includes OpenTelemetry trace context.

    Adds the following attributes to log records:
    - trace_id: 32-char hex string of the current trace ID
    - span_id: 16-char hex string of the current span ID

    These can be used in log formatters to correlate logs with traces.
    """

    def __init__(self, *args, **kwargs):
        """Initialize LogRecord with OpenTelemetry context."""
        super().__init__(*args, **kwargs)

        current_span = trace.get_current_span()
        if current_span is not None:
            span_context = current_span.get_span_context()
            if span_context.is_valid:
                # Use standard hex format for trace IDs
                self.trace_id = format(span_context.trace_id, "032x")
                self.span_id = format(span_context.span_id, "016x")
            else:
                self.trace_id = None
                self.span_id = None
        else:
            self.trace_id = None
            self.span_id = None

        # Keep legacy attribute names for backwards compatibility
        self.otelTraceID = self.trace_id
        self.otelSpanID = self.span_id
