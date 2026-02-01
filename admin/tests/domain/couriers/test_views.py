"""Tests for courier views."""

from unittest.mock import MagicMock, patch

import pytest
from django.contrib.auth.models import User
from django.contrib.messages import get_messages
from django.test import Client, RequestFactory, override_settings
from django.urls import reverse

from infrastructure.grpc.delivery_client import (
    Courier,
    CourierPoolResult,
    DeliveryServiceError,
    WorkHours,
)


@pytest.fixture
def staff_user(db):
    """Create a staff user for testing."""
    return User.objects.create_user(
        username="teststaff",
        email="staff@test.com",
        password="testpass123",
        is_staff=True,
    )


@pytest.fixture
def authenticated_client(staff_user):
    """Create an authenticated Django test client."""
    client = Client()
    client.login(username="teststaff", password="testpass123")
    return client


@pytest.fixture
def sample_work_hours():
    """Sample work hours."""
    return WorkHours(
        start_time="09:00",
        end_time="18:00",
        work_days=[1, 2, 3, 4, 5],
    )


@pytest.fixture
def sample_courier(sample_work_hours):
    """Sample courier for testing."""
    return Courier(
        courier_id="test-courier-id-123",
        name="Test Courier",
        phone="+49123456789",
        email="test@example.com",
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
        created_at=None,
        last_active_at=None,
    )


@pytest.fixture
def sample_pool_result(sample_courier):
    """Sample courier pool result."""
    return CourierPoolResult(
        couriers=[sample_courier],
        total_count=1,
        current_page=1,
        page_size=20,
        total_pages=1,
    )


@pytest.fixture
def empty_pool_result():
    """Empty courier pool result."""
    return CourierPoolResult(
        couriers=[],
        total_count=0,
        current_page=1,
        page_size=20,
        total_pages=0,
    )


@pytest.mark.django_db
class TestCourierListView:
    """Tests for courier_list view."""

    def test_list_requires_staff(self, client):
        """Anonymous users should be redirected."""
        response = client.get("/admin/couriers/")
        assert response.status_code == 302
        assert "login" in response.url or "admin" in response.url

    def test_list_renders_template(self, authenticated_client, sample_pool_result):
        """Staff users should see courier list."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.return_value = sample_pool_result
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/")

        assert response.status_code == 200
        assert "couriers" in response.context

    def test_list_with_filters(self, authenticated_client, sample_pool_result):
        """List should apply filters from query params."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.return_value = sample_pool_result
            mock_get_client.return_value = mock_client

            response = authenticated_client.get(
                "/admin/couriers/?status=FREE&transport_type=BICYCLE"
            )

        assert response.status_code == 200
        # Verify filter was passed to client
        mock_client.get_courier_pool.assert_called_once()
        call_kwargs = mock_client.get_courier_pool.call_args.kwargs
        assert call_kwargs["status_filter"] == ["FREE"]
        assert call_kwargs["transport_type_filter"] == ["BICYCLE"]

    def test_list_with_pagination(self, authenticated_client, sample_pool_result):
        """List should handle pagination params."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.return_value = sample_pool_result
            mock_get_client.return_value = mock_client

            response = authenticated_client.get(
                "/admin/couriers/?page=2&page_size=50"
            )

        assert response.status_code == 200
        call_kwargs = mock_client.get_courier_pool.call_args.kwargs
        assert call_kwargs["page"] == 2
        assert call_kwargs["page_size"] == 50

    def test_list_handles_grpc_error(self, authenticated_client):
        """List should show error message on gRPC failure."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.side_effect = DeliveryServiceError(
                "Connection failed"
            )
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/")

        assert response.status_code == 200
        assert response.context["couriers"] == []
        messages = list(get_messages(response.wsgi_request))
        assert len(messages) == 1
        assert "Error" in str(messages[0])

    def test_list_empty_result(self, authenticated_client, empty_pool_result):
        """List should handle empty results."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.return_value = empty_pool_result
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/")

        assert response.status_code == 200
        assert response.context["total_count"] == 0


@pytest.mark.django_db
class TestCourierDetailView:
    """Tests for courier_detail view."""

    def test_detail_requires_staff(self, client):
        """Anonymous users should be redirected."""
        response = client.get("/admin/couriers/test-id/")
        assert response.status_code == 302

    def test_detail_renders_courier(self, authenticated_client, sample_courier):
        """Detail view should render courier data."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier.return_value = sample_courier
            mock_get_client.return_value = mock_client

            response = authenticated_client.get(
                "/admin/couriers/test-courier-id-123/"
            )

        assert response.status_code == 200
        assert response.context["courier"] == sample_courier

    def test_detail_not_found_redirects(self, authenticated_client):
        """Detail view should redirect when courier not found."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier.return_value = None
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/nonexistent/")

        assert response.status_code == 302

    def test_detail_handles_grpc_error(self, authenticated_client):
        """Detail view should redirect on gRPC error."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier.side_effect = DeliveryServiceError(
                "Connection failed"
            )
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/test-id/")

        assert response.status_code == 302


@pytest.mark.django_db
class TestCourierRegisterView:
    """Tests for courier_register view."""

    def test_register_get_renders_form(self, authenticated_client):
        """GET should render registration form."""
        response = authenticated_client.get("/admin/couriers/register/")

        assert response.status_code == 200
        assert "form" in response.context

    def test_register_post_valid_form(self, authenticated_client):
        """Valid POST should register courier."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.register_courier.return_value = "new-courier-id"
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/register/",
                {
                    "name": "New Courier",
                    "phone": "+49123456789",
                    "email": "new@example.com",
                    "transport_type": "BICYCLE",
                    "max_distance_km": "10.0",
                    "work_zone": "Berlin-Mitte",
                    "work_start": "09:00",
                    "work_end": "18:00",
                    "work_days": ["1", "2", "3", "4", "5"],
                },
            )

        assert response.status_code == 302
        assert "new-courier-id" in response.url

    def test_register_post_invalid_form(self, authenticated_client):
        """Invalid POST should re-render form with errors."""
        response = authenticated_client.post(
            "/admin/couriers/register/",
            {
                "name": "",  # Required field missing
                "phone": "+49123456789",
                "email": "new@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": "10.0",
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": ["1", "2", "3"],
            },
        )

        assert response.status_code == 200
        assert "form" in response.context
        assert response.context["form"].errors

    def test_register_post_grpc_error(self, authenticated_client):
        """gRPC error should show error message."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.register_courier.side_effect = DeliveryServiceError(
                "Email already exists"
            )
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/register/",
                {
                    "name": "New Courier",
                    "phone": "+49123456789",
                    "email": "existing@example.com",
                    "transport_type": "BICYCLE",
                    "max_distance_km": "10.0",
                    "work_zone": "Berlin-Mitte",
                    "work_start": "09:00",
                    "work_end": "18:00",
                    "work_days": ["1", "2", "3"],
                },
            )

        assert response.status_code == 200
        messages = list(get_messages(response.wsgi_request))
        assert any("Error" in str(m) for m in messages)


@pytest.mark.django_db
class TestCourierActivateView:
    """Tests for courier_activate view."""

    def test_activate_success(self, authenticated_client):
        """Successful activation should redirect with message."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.activate_courier.return_value = True
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/activate/"
            )

        assert response.status_code == 302
        assert "test-id" in response.url

    def test_activate_grpc_error(self, authenticated_client):
        """gRPC error should show error message."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.activate_courier.side_effect = DeliveryServiceError(
                "Not found"
            )
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/activate/"
            )

        assert response.status_code == 302
        messages = list(get_messages(response.wsgi_request))
        assert any("Error" in str(m) for m in messages)


@pytest.mark.django_db
class TestCourierDeactivateView:
    """Tests for courier_deactivate view."""

    def test_deactivate_success(self, authenticated_client):
        """Successful deactivation should redirect."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.deactivate_courier.return_value = True
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/deactivate/",
                {"reason": "End of shift"},
            )

        assert response.status_code == 302


@pytest.mark.django_db
class TestCourierArchiveView:
    """Tests for courier_archive view."""

    def test_archive_success(self, authenticated_client):
        """Successful archive should redirect to list."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.archive_courier.return_value = True
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/archive/",
                {"confirm": "on", "reason": "Resigned"},
            )

        assert response.status_code == 302

    def test_archive_without_confirm(self, authenticated_client):
        """Archive without confirm should show error."""
        response = authenticated_client.post(
            "/admin/couriers/test-id/archive/",
            {"reason": "Resigned"},
        )

        assert response.status_code == 302
        messages = list(get_messages(response.wsgi_request))
        assert any("confirm" in str(m).lower() for m in messages)


@pytest.mark.django_db
class TestCourierUpdateContactView:
    """Tests for courier_update_contact view."""

    def test_update_contact_success(self, authenticated_client):
        """Successful update should redirect."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.update_contact_info.return_value = True
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/update-contact/",
                {"phone": "+49999999999"},
            )

        assert response.status_code == 302

    def test_update_contact_empty_form(self, authenticated_client):
        """Empty form should show error."""
        response = authenticated_client.post(
            "/admin/couriers/test-id/update-contact/",
            {},
        )

        assert response.status_code == 302
        messages = list(get_messages(response.wsgi_request))
        assert len(messages) >= 1


@pytest.mark.django_db
class TestCourierUpdateScheduleView:
    """Tests for courier_update_schedule view."""

    def test_update_schedule_success(self, authenticated_client):
        """Successful update should redirect."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.update_work_schedule.return_value = True
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/update-schedule/",
                {"work_zone": "Berlin-Kreuzberg"},
            )

        assert response.status_code == 302


@pytest.mark.django_db
class TestCourierChangeTransportView:
    """Tests for courier_change_transport view."""

    def test_change_transport_success(self, authenticated_client):
        """Successful change should redirect with message."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.change_transport_type.return_value = 5
            mock_get_client.return_value = mock_client

            response = authenticated_client.post(
                "/admin/couriers/test-id/change-transport/",
                {"transport_type": "CAR"},
            )

        assert response.status_code == 302
        messages = list(get_messages(response.wsgi_request))
        assert any("5 packages" in str(m) for m in messages)


@pytest.mark.django_db
class TestCourierMapView:
    """Tests for courier_map view."""

    def test_map_renders(self, authenticated_client, sample_pool_result):
        """Map view should render."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.return_value = sample_pool_result
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/map/")

        assert response.status_code == 200
        assert "couriers" in response.context

    def test_map_handles_error(self, authenticated_client):
        """Map should handle gRPC errors gracefully."""
        with patch(
            "domain.couriers.views.get_delivery_client"
        ) as mock_get_client:
            mock_client = MagicMock()
            mock_client.get_courier_pool.side_effect = DeliveryServiceError(
                "Connection failed"
            )
            mock_get_client.return_value = mock_client

            response = authenticated_client.get("/admin/couriers/map/")

        assert response.status_code == 200
        assert response.context["couriers"] == []
