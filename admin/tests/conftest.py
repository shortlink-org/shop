"""Shared fixtures for admin service tests.

Note: Courier/Delivery fixtures removed - management moved to admin-ui.
"""

import pytest


@pytest.fixture
def django_rf():
    """Django RequestFactory fixture."""
    from django.test import RequestFactory

    return RequestFactory()
