"""Shared fixtures for admin service tests."""

from datetime import datetime
from unittest.mock import MagicMock, patch

import pytest

from infrastructure.grpc.delivery_client import (
    Courier,
    CourierPoolResult,
    DeliveryClient,
    DeliveryServiceError,
    Location,
    WorkHours,
)


@pytest.fixture
def sample_work_hours():
    """Sample work hours fixture."""
    return WorkHours(
        start_time="09:00",
        end_time="18:00",
        work_days=[1, 2, 3, 4, 5],
    )


@pytest.fixture
def sample_location():
    """Sample location fixture."""
    return Location(
        latitude=52.5200,
        longitude=13.4050,
    )


@pytest.fixture
def sample_courier(sample_work_hours):
    """Sample courier fixture."""
    return Courier(
        courier_id="123e4567-e89b-12d3-a456-426614174000",
        name="John Doe",
        phone="+49123456789",
        email="john.doe@example.com",
        transport_type="BICYCLE",
        max_distance_km=10.0,
        status="FREE",
        current_load=1,
        max_load=2,
        rating=4.5,
        work_hours=sample_work_hours,
        work_zone="Berlin-Mitte",
        current_location=None,
        successful_deliveries=50,
        failed_deliveries=2,
        created_at=datetime(2024, 1, 15, 10, 30, 0),
        last_active_at=datetime(2024, 6, 1, 14, 0, 0),
    )


@pytest.fixture
def sample_courier_unavailable(sample_work_hours):
    """Sample unavailable courier fixture."""
    return Courier(
        courier_id="223e4567-e89b-12d3-a456-426614174001",
        name="Jane Smith",
        phone="+49987654321",
        email="jane.smith@example.com",
        transport_type="CAR",
        max_distance_km=25.0,
        status="UNAVAILABLE",
        current_load=0,
        max_load=5,
        rating=4.8,
        work_hours=sample_work_hours,
        work_zone="Berlin-Kreuzberg",
        current_location=None,
        successful_deliveries=120,
        failed_deliveries=5,
        created_at=datetime(2023, 6, 1, 9, 0, 0),
        last_active_at=None,
    )


@pytest.fixture
def sample_courier_archived(sample_work_hours):
    """Sample archived courier fixture."""
    return Courier(
        courier_id="323e4567-e89b-12d3-a456-426614174002",
        name="Archived User",
        phone="+49111222333",
        email="archived@example.com",
        transport_type="WALKING",
        max_distance_km=5.0,
        status="ARCHIVED",
        current_load=0,
        max_load=1,
        rating=3.5,
        work_hours=sample_work_hours,
        work_zone="Berlin-Mitte",
        current_location=None,
        successful_deliveries=10,
        failed_deliveries=8,
        created_at=datetime(2023, 1, 1, 8, 0, 0),
        last_active_at=None,
    )


@pytest.fixture
def sample_courier_pool_result(sample_courier, sample_courier_unavailable):
    """Sample courier pool result fixture."""
    return CourierPoolResult(
        couriers=[sample_courier, sample_courier_unavailable],
        total_count=2,
        current_page=1,
        page_size=20,
        total_pages=1,
    )


@pytest.fixture
def empty_courier_pool_result():
    """Empty courier pool result fixture."""
    return CourierPoolResult(
        couriers=[],
        total_count=0,
        current_page=1,
        page_size=20,
        total_pages=0,
    )


@pytest.fixture
def mock_delivery_client(sample_courier, sample_courier_pool_result):
    """Mock DeliveryClient fixture.

    Returns a MagicMock that can be configured for specific tests.
    """
    client = MagicMock(spec=DeliveryClient)

    # Default return values
    client.get_courier.return_value = sample_courier
    client.get_courier_pool.return_value = sample_courier_pool_result
    client.register_courier.return_value = "new-courier-id-123"
    client.activate_courier.return_value = True
    client.deactivate_courier.return_value = True
    client.archive_courier.return_value = True
    client.update_contact_info.return_value = True
    client.update_work_schedule.return_value = True
    client.change_transport_type.return_value = 3

    return client


@pytest.fixture
def mock_delivery_client_error():
    """Mock DeliveryClient that raises errors."""
    client = MagicMock(spec=DeliveryClient)

    error = DeliveryServiceError("Connection failed")
    client.get_courier.side_effect = error
    client.get_courier_pool.side_effect = error
    client.register_courier.side_effect = error
    client.activate_courier.side_effect = error
    client.deactivate_courier.side_effect = error
    client.archive_courier.side_effect = error
    client.update_contact_info.side_effect = error
    client.update_work_schedule.side_effect = error
    client.change_transport_type.side_effect = error

    return client


@pytest.fixture
def mock_delivery_client_not_found():
    """Mock DeliveryClient that returns None for courier lookups."""
    client = MagicMock(spec=DeliveryClient)
    client.get_courier.return_value = None
    return client


@pytest.fixture
def patch_get_delivery_client(mock_delivery_client):
    """Patch get_delivery_client to return mock client."""
    with patch(
        "domain.couriers.views.get_delivery_client",
        return_value=mock_delivery_client,
    ) as patched:
        yield patched


@pytest.fixture
def django_rf():
    """Django RequestFactory fixture."""
    from django.test import RequestFactory

    return RequestFactory()
