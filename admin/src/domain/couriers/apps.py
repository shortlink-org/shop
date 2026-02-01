"""Couriers app configuration."""

from django.apps import AppConfig


class CouriersConfig(AppConfig):
    """Configuration for the couriers app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "domain.couriers"
    verbose_name = "Courier Management"
