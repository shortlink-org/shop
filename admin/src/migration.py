#!/usr/bin/env python
"""Lightweight Django migration runner.

This script is optimized for running migrations in init containers.
It uses settings_migration.py by default for faster startup.

To use full settings, set DJANGO_SETTINGS_MODULE=admin.settings
"""

import os
import sys


def main():
    """Run migration tasks with minimal overhead."""
    # Use lightweight settings by default for migrations
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "admin.settings_migration")

    from django.core.management import execute_from_command_line

    execute_from_command_line(sys.argv)


if __name__ == "__main__":
    main()
