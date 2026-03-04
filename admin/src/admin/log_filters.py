"""Logging filters for admin service."""

from __future__ import annotations

import logging


class DropEmptyJsonLineFilter(logging.Filter):
    """Drop empty/noise log lines like '{}' emitted by upstream loggers."""

    NOISE_MESSAGES = {"{}", ""}

    def filter(self, record: logging.LogRecord) -> bool:
        message = record.getMessage()
        return message.strip() not in self.NOISE_MESSAGES
