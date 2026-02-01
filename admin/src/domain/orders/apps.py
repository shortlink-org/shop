"""Django app configuration for orders domain."""

from django.apps import AppConfig


class OrdersConfig(AppConfig):
    """Configuration for the orders app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "domain.orders"
    verbose_name = "Orders"
