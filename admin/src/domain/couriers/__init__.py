"""Courier management domain module.

This module provides admin interface for managing couriers through
the Delivery Service gRPC API. Unlike other domain modules, it does
not use Django models - data is fetched from the external service.
"""

default_app_config = "domain.couriers.apps.CouriersConfig"
