"""Minimal Django settings for running migrations faster.

This settings module excludes unnecessary apps, middleware, and services
to speed up migration execution in init containers.

Key optimizations:
- Only essential INSTALLED_APPS (no debug_toolbar, health_check, prometheus, etc.)
- DummyCache instead of Redis (no network calls)
- No middleware stack
- Minimal logging
"""

import os
from pathlib import Path

import environ

env = environ.Env()

# GDAL/GEOS library paths for GeoDjango
# These are required for django.contrib.gis to find the libraries
GDAL_LIBRARY_PATH = env("GDAL_LIBRARY_PATH", default=None)
GEOS_LIBRARY_PATH = env("GEOS_LIBRARY_PATH", default=None)

# Read .env file if it exists
environ.Env.read_env(os.path.join(Path(__file__).resolve().parent.parent, ".env"))

# Build paths inside the project like this: BASE_DIR / 'subdir'.
BASE_DIR = Path(__file__).resolve().parent.parent

# Minimal security settings for migrations
SECRET_KEY = "migration-only-key-not-for-production"
DEBUG = False
ALLOWED_HOSTS = ["*"]

# Minimal INSTALLED_APPS - only what's needed for migrations
INSTALLED_APPS = [
    "django.contrib.contenttypes",
    "django.contrib.auth",
    "django.contrib.gis",  # Required for PostGIS models
    # Domain apps with migrations
    "domain.goods",
    "domain.offices",
]

# No middleware needed for migrations
MIDDLEWARE = []

# Database - same as main settings
DATABASES = {
    "default": {
        "ENGINE": "django.contrib.gis.db.backends.postgis",
        "NAME": env("POSTGRES_DB", default="shortlink"),
        "USER": env("POSTGRES_USER", default="postgres"),
        "PASSWORD": env("POSTGRES_PASSWORD", default="shortlink"),
        "HOST": env("POSTGRES_HOST", default="localhost"),
        "PORT": env("POSTGRES_PORT", default="5432"),
        "OPTIONS": {
            "sslmode": env("POSTGRES_SSLMODE", default="prefer"),
        },
    }
}

# Use DummyCache to avoid Redis connection during migrations
CACHES = {
    "default": {
        "BACKEND": "django.core.cache.backends.dummy.DummyCache",
    }
}

# Internationalization
LANGUAGE_CODE = "en-us"
TIME_ZONE = "UTC"
USE_I18N = False  # Disable i18n for faster startup
USE_TZ = True

# Default primary key field type
DEFAULT_AUTO_FIELD = "django.db.models.BigAutoField"

# Minimal logging - only errors
LOGGING = {
    "version": 1,
    "disable_existing_loggers": True,
    "handlers": {
        "console": {
            "class": "logging.StreamHandler",
        },
    },
    "root": {
        "handlers": ["console"],
        "level": "WARNING",
    },
    "loggers": {
        "django.db.backends": {
            "level": "WARNING",
            "handlers": ["console"],
            "propagate": False,
        },
    },
}
