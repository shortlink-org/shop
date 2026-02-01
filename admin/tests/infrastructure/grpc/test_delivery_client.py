"""Tests for DeliveryClient."""

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


class TestDeliveryClientInit:
    """Tests for DeliveryClient initialization."""

    def test_default_host(self):
        """Client should use default host when none provided."""
        with patch.object(
            DeliveryClient, "_ensure_connected", return_value=None
        ):
            client = DeliveryClient()
            # Default from settings or fallback
            assert client.host is not None

    def test_custom_host(self):
        """Client should use custom host when provided."""
        client = DeliveryClient(host="custom-host:50052")
        assert client.host == "custom-host:50052"

    def test_initial_state(self):
        """Client should have no connection initially."""
        client = DeliveryClient(host="localhost:50051")
        assert client._channel is None
        assert client._stub is None


class TestDeliveryClientMappings:
    """Tests for transport type and status mappings."""

    def test_transport_types_mapping(self):
        """Transport types should map correctly."""
        assert DeliveryClient.TRANSPORT_TYPES[0] == "UNSPECIFIED"
        assert DeliveryClient.TRANSPORT_TYPES[1] == "WALKING"
        assert DeliveryClient.TRANSPORT_TYPES[2] == "BICYCLE"
        assert DeliveryClient.TRANSPORT_TYPES[3] == "MOTORCYCLE"
        assert DeliveryClient.TRANSPORT_TYPES[4] == "CAR"

    def test_transport_type_values_mapping(self):
        """Transport type values should map correctly."""
        assert DeliveryClient.TRANSPORT_TYPE_VALUES["UNSPECIFIED"] == 0
        assert DeliveryClient.TRANSPORT_TYPE_VALUES["WALKING"] == 1
        assert DeliveryClient.TRANSPORT_TYPE_VALUES["BICYCLE"] == 2
        assert DeliveryClient.TRANSPORT_TYPE_VALUES["MOTORCYCLE"] == 3
        assert DeliveryClient.TRANSPORT_TYPE_VALUES["CAR"] == 4

    def test_courier_statuses_mapping(self):
        """Courier statuses should map correctly."""
        assert DeliveryClient.COURIER_STATUSES[0] == "UNSPECIFIED"
        assert DeliveryClient.COURIER_STATUSES[1] == "UNAVAILABLE"
        assert DeliveryClient.COURIER_STATUSES[2] == "FREE"
        assert DeliveryClient.COURIER_STATUSES[3] == "BUSY"
        assert DeliveryClient.COURIER_STATUSES[4] == "ARCHIVED"

    def test_courier_status_values_mapping(self):
        """Courier status values should map correctly."""
        assert DeliveryClient.COURIER_STATUS_VALUES["UNSPECIFIED"] == 0
        assert DeliveryClient.COURIER_STATUS_VALUES["UNAVAILABLE"] == 1
        assert DeliveryClient.COURIER_STATUS_VALUES["FREE"] == 2
        assert DeliveryClient.COURIER_STATUS_VALUES["BUSY"] == 3
        assert DeliveryClient.COURIER_STATUS_VALUES["ARCHIVED"] == 4


class TestProtoToCourier:
    """Tests for _proto_to_courier conversion."""

    @pytest.fixture
    def mock_proto_courier(self):
        """Create a mock protobuf Courier."""
        proto = MagicMock()
        proto.courier_id = "test-id-123"
        proto.name = "Test Courier"
        proto.phone = "+49123456789"
        proto.email = "test@example.com"
        proto.transport_type = 2  # BICYCLE
        proto.max_distance_km = 10.0
        proto.status = 2  # FREE
        proto.current_load = 1
        proto.max_load = 2
        proto.rating = 4.5
        proto.work_zone = "Berlin-Mitte"
        proto.successful_deliveries = 50
        proto.failed_deliveries = 2

        # Work hours
        proto.work_hours = MagicMock()
        proto.work_hours.start_time = "09:00"
        proto.work_hours.end_time = "18:00"
        proto.work_hours.work_days = [1, 2, 3, 4, 5]

        # No location
        proto.HasField = MagicMock(return_value=False)

        # Timestamps
        proto.created_at = MagicMock()
        proto.created_at.ToDatetime = MagicMock(
            return_value=datetime(2024, 1, 15, 10, 0)
        )

        return proto

    def test_converts_basic_fields(self, mock_proto_courier):
        """Should convert basic courier fields."""
        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.courier_id == "test-id-123"
        assert courier.name == "Test Courier"
        assert courier.phone == "+49123456789"
        assert courier.email == "test@example.com"

    def test_converts_transport_type(self, mock_proto_courier):
        """Should convert transport type from enum to string."""
        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.transport_type == "BICYCLE"

    def test_converts_status(self, mock_proto_courier):
        """Should convert status from enum to string."""
        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.status == "FREE"

    def test_converts_work_hours(self, mock_proto_courier):
        """Should convert work hours to WorkHours dataclass."""
        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.work_hours is not None
        assert courier.work_hours.start_time == "09:00"
        assert courier.work_hours.end_time == "18:00"
        assert courier.work_hours.work_days == [1, 2, 3, 4, 5]

    def test_handles_missing_location(self, mock_proto_courier):
        """Should handle missing location gracefully."""
        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.current_location is None

    def test_converts_location_when_present(self, mock_proto_courier):
        """Should convert location when present."""
        mock_proto_courier.HasField = MagicMock(
            side_effect=lambda x: x == "current_location"
        )
        mock_proto_courier.current_location = MagicMock()
        mock_proto_courier.current_location.latitude = 52.52
        mock_proto_courier.current_location.longitude = 13.405

        client = DeliveryClient(host="localhost:50051")
        courier = client._proto_to_courier(mock_proto_courier)

        assert courier.current_location is not None
        assert courier.current_location.latitude == 52.52
        assert courier.current_location.longitude == 13.405


class TestGetCourier:
    """Tests for get_courier method."""

    @pytest.fixture
    def mock_stub_and_client(self):
        """Create client with mocked stub."""
        client = DeliveryClient(host="localhost:50051")
        client._channel = MagicMock()
        client._stub = MagicMock()
        return client

    def test_get_courier_success(self, mock_stub_and_client):
        """Should return courier on success."""
        client = mock_stub_and_client

        # Mock response
        mock_response = MagicMock()
        mock_courier = MagicMock()
        mock_courier.courier_id = "test-id"
        mock_courier.name = "Test"
        mock_courier.phone = "+49123"
        mock_courier.email = "test@test.com"
        mock_courier.transport_type = 1
        mock_courier.max_distance_km = 10.0
        mock_courier.status = 2
        mock_courier.current_load = 0
        mock_courier.max_load = 1
        mock_courier.rating = 5.0
        mock_courier.work_zone = "zone"
        mock_courier.successful_deliveries = 10
        mock_courier.failed_deliveries = 0
        mock_courier.work_hours = None
        mock_courier.HasField = MagicMock(return_value=False)
        mock_courier.created_at = None

        mock_response.courier = mock_courier
        client._stub.GetCourier.return_value = mock_response

        with patch.object(client, "_ensure_connected"):
            with patch(
                "infrastructure.grpc.delivery_client.delivery_pb2"
            ) as mock_pb2:
                mock_pb2.GetCourierRequest.return_value = MagicMock()
                result = client.get_courier("test-id")

        assert result is not None
        assert result.courier_id == "test-id"

    def test_get_courier_not_found(self, mock_stub_and_client):
        """Should return None when courier not found."""
        import grpc

        client = mock_stub_and_client

        # Mock NOT_FOUND error
        error = MagicMock(spec=grpc.RpcError)
        error.code.return_value = grpc.StatusCode.NOT_FOUND
        client._stub.GetCourier.side_effect = error

        with patch.object(client, "_ensure_connected"):
            with patch(
                "infrastructure.grpc.delivery_client.delivery_pb2"
            ) as mock_pb2:
                mock_pb2.GetCourierRequest.return_value = MagicMock()
                result = client.get_courier("non-existent")

        assert result is None

    def test_get_courier_error(self, mock_stub_and_client):
        """Should raise DeliveryServiceError on other gRPC errors."""
        import grpc

        client = mock_stub_and_client

        # Mock INTERNAL error
        error = MagicMock(spec=grpc.RpcError)
        error.code.return_value = grpc.StatusCode.INTERNAL
        client._stub.GetCourier.side_effect = error

        with patch.object(client, "_ensure_connected"):
            with patch(
                "infrastructure.grpc.delivery_client.delivery_pb2"
            ) as mock_pb2:
                mock_pb2.GetCourierRequest.return_value = MagicMock()
                with pytest.raises(DeliveryServiceError):
                    client.get_courier("test-id")


class TestClose:
    """Tests for close method."""

    def test_close_closes_channel(self):
        """Close should close the channel."""
        client = DeliveryClient(host="localhost:50051")
        client._channel = MagicMock()
        client._stub = MagicMock()

        client.close()

        client._channel.close.assert_called_once()
        assert client._channel is None
        assert client._stub is None

    def test_close_when_not_connected(self):
        """Close should handle not connected state."""
        client = DeliveryClient(host="localhost:50051")
        # Should not raise
        client.close()


class TestWorkHoursDataclass:
    """Tests for WorkHours dataclass."""

    def test_work_hours_creation(self):
        """WorkHours should be created correctly."""
        wh = WorkHours(
            start_time="09:00",
            end_time="18:00",
            work_days=[1, 2, 3, 4, 5],
        )
        assert wh.start_time == "09:00"
        assert wh.end_time == "18:00"
        assert wh.work_days == [1, 2, 3, 4, 5]


class TestLocationDataclass:
    """Tests for Location dataclass."""

    def test_location_creation(self):
        """Location should be created correctly."""
        loc = Location(latitude=52.52, longitude=13.405)
        assert loc.latitude == 52.52
        assert loc.longitude == 13.405


class TestCourierDataclass:
    """Tests for Courier dataclass."""

    def test_courier_creation(self, sample_work_hours):
        """Courier should be created correctly."""
        courier = Courier(
            courier_id="test-id",
            name="Test",
            phone="+49123",
            email="test@test.com",
            transport_type="BICYCLE",
            max_distance_km=10.0,
            status="FREE",
            current_load=1,
            max_load=2,
            rating=4.5,
            work_hours=sample_work_hours,
            work_zone="Berlin",
            current_location=None,
            successful_deliveries=50,
            failed_deliveries=2,
            created_at=None,
            last_active_at=None,
        )
        assert courier.courier_id == "test-id"
        assert courier.status == "FREE"
