"""Delivery Service gRPC Client.

This module provides a client for communicating with the Delivery Service
via gRPC. It wraps the generated protobuf stubs with a Pythonic interface.
"""

import logging
from dataclasses import dataclass
from datetime import datetime
from typing import Optional

import grpc
from django.conf import settings

logger = logging.getLogger(__name__)


@dataclass
class WorkHours:
    """Courier work hours."""

    start_time: str
    end_time: str
    work_days: list[int]


@dataclass
class Location:
    """GPS location."""

    latitude: float
    longitude: float


@dataclass
class Courier:
    """Courier data from Delivery Service."""

    courier_id: str
    name: str
    phone: str
    email: str
    transport_type: str
    max_distance_km: float
    status: str
    current_load: int
    max_load: int
    rating: float
    work_hours: Optional[WorkHours]
    work_zone: str
    current_location: Optional[Location]
    successful_deliveries: int
    failed_deliveries: int
    created_at: Optional[datetime]
    last_active_at: Optional[datetime]


@dataclass
class CourierPoolResult:
    """Result of GetCourierPool query."""

    couriers: list[Courier]
    total_count: int
    current_page: int
    page_size: int
    total_pages: int


@dataclass
class Address:
    """Delivery address."""

    street: str
    city: str
    postal_code: str
    country: str
    latitude: float
    longitude: float


@dataclass
class DeliveryRecord:
    """A delivery record for a courier."""

    package_id: str
    order_id: str
    status: str
    pickup_address: Optional[Address]
    delivery_address: Optional[Address]
    assigned_at: Optional[datetime]
    delivered_at: Optional[datetime]
    priority: str


@dataclass
class CourierDeliveriesResult:
    """Result of GetCourierDeliveries query."""

    deliveries: list[DeliveryRecord]
    total_count: int


class DeliveryServiceError(Exception):
    """Base exception for Delivery Service errors."""

    pass


class CourierNotFoundError(DeliveryServiceError):
    """Courier not found."""

    pass


class DeliveryClient:
    """Client for Delivery Service gRPC API.

    This client provides methods to interact with the Delivery Service
    for courier management operations.
    """

    # Transport type mapping
    TRANSPORT_TYPES = {
        0: "UNSPECIFIED",
        1: "WALKING",
        2: "BICYCLE",
        3: "MOTORCYCLE",
        4: "CAR",
    }

    TRANSPORT_TYPE_VALUES = {
        "UNSPECIFIED": 0,
        "WALKING": 1,
        "BICYCLE": 2,
        "MOTORCYCLE": 3,
        "CAR": 4,
    }

    # Status mapping
    COURIER_STATUSES = {
        0: "UNSPECIFIED",
        1: "UNAVAILABLE",
        2: "FREE",
        3: "BUSY",
        4: "ARCHIVED",
    }

    COURIER_STATUS_VALUES = {
        "UNSPECIFIED": 0,
        "UNAVAILABLE": 1,
        "FREE": 2,
        "BUSY": 3,
        "ARCHIVED": 4,
    }

    PACKAGE_STATUSES = {
        0: "UNSPECIFIED",
        1: "ACCEPTED",
        2: "IN_POOL",
        3: "ASSIGNED",
        4: "IN_TRANSIT",
        5: "DELIVERED",
        6: "NOT_DELIVERED",
        7: "REQUIRES_HANDLING",
    }

    PRIORITIES = {
        0: "UNSPECIFIED",
        1: "NORMAL",
        2: "URGENT",
    }

    def __init__(self, host: Optional[str] = None, auth_token: Optional[str] = None):
        """Initialize the Delivery Service client.

        Args:
            host: gRPC host address. Defaults to settings.DELIVERY_GRPC_HOST.
            auth_token: JWT token for authentication. Will be passed as
                        Authorization header to downstream gRPC service.
        """
        self.host = host or getattr(settings, "DELIVERY_GRPC_HOST", "localhost:50051")
        self._channel: Optional[grpc.Channel] = None
        self._stub = None
        self._auth_token = auth_token

    def _get_metadata(self) -> list[tuple[str, str]]:
        """Get gRPC call metadata including auth token if available."""
        if self._auth_token:
            return [("authorization", f"Bearer {self._auth_token}")]
        return []

    def _ensure_connected(self):
        """Ensure gRPC channel is connected."""
        if self._channel is None:
            self._channel = grpc.insecure_channel(self.host)
            # Import generated stubs
            try:
                import delivery_pb2_grpc

                self._stub = delivery_pb2_grpc.DeliveryServiceStub(self._channel)
            except ImportError:
                logger.warning(
                    "Generated gRPC stubs not found. Run 'buf generate' to generate them."
                )
                raise DeliveryServiceError(
                    "gRPC stubs not generated. Run 'buf generate' in admin directory."
                )

    def _proto_to_courier(self, proto_courier) -> Courier:
        """Convert protobuf Courier to dataclass."""
        work_hours = None
        if proto_courier.work_hours:
            work_hours = WorkHours(
                start_time=proto_courier.work_hours.start_time,
                end_time=proto_courier.work_hours.end_time,
                work_days=list(proto_courier.work_hours.work_days),
            )

        location = None
        if proto_courier.HasField("current_location"):
            location = Location(
                latitude=proto_courier.current_location.latitude,
                longitude=proto_courier.current_location.longitude,
            )

        created_at = None
        if proto_courier.created_at:
            created_at = proto_courier.created_at.ToDatetime()

        last_active_at = None
        if proto_courier.HasField("last_active_at"):
            last_active_at = proto_courier.last_active_at.ToDatetime()

        return Courier(
            courier_id=proto_courier.courier_id,
            name=proto_courier.name,
            phone=proto_courier.phone,
            email=proto_courier.email,
            transport_type=self.TRANSPORT_TYPES.get(
                proto_courier.transport_type, "UNSPECIFIED"
            ),
            max_distance_km=proto_courier.max_distance_km,
            status=self.COURIER_STATUSES.get(proto_courier.status, "UNSPECIFIED"),
            current_load=proto_courier.current_load,
            max_load=proto_courier.max_load,
            rating=proto_courier.rating,
            work_hours=work_hours,
            work_zone=proto_courier.work_zone,
            current_location=location,
            successful_deliveries=proto_courier.successful_deliveries,
            failed_deliveries=proto_courier.failed_deliveries,
            created_at=created_at,
            last_active_at=last_active_at,
        )

    def get_courier(
        self, courier_id: str, include_location: bool = False
    ) -> Optional[Courier]:
        """Get a single courier by ID.

        Args:
            courier_id: The courier's unique identifier.
            include_location: Whether to include current location.

        Returns:
            Courier data or None if not found.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.GetCourierRequest(
            courier_id=courier_id,
            include_location=include_location,
        )

        try:
            response = self._stub.GetCourier(request, metadata=self._get_metadata())
            return self._proto_to_courier(response.courier)
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.NOT_FOUND:
                return None
            logger.error(f"gRPC error getting courier: {e}")
            raise DeliveryServiceError(f"Failed to get courier: {e}")

    def get_courier_deliveries(
        self, courier_id: str, limit: int = 5
    ) -> CourierDeliveriesResult:
        """Get recent deliveries for a courier.

        Args:
            courier_id: The courier's unique identifier.
            limit: Maximum number of deliveries to return (default 5).

        Returns:
            CourierDeliveriesResult with deliveries and total count.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.GetCourierDeliveriesRequest(
            courier_id=courier_id,
            limit=limit,
        )

        try:
            response = self._stub.GetCourierDeliveries(request, metadata=self._get_metadata())
            deliveries = []
            for proto_delivery in response.deliveries:
                pickup = None
                if proto_delivery.HasField("pickup_address"):
                    pa = proto_delivery.pickup_address
                    pickup = Address(
                        street=pa.street,
                        city=pa.city,
                        postal_code=pa.postal_code,
                        country=pa.country,
                        latitude=pa.latitude,
                        longitude=pa.longitude,
                    )

                delivery_addr = None
                if proto_delivery.HasField("delivery_address"):
                    da = proto_delivery.delivery_address
                    delivery_addr = Address(
                        street=da.street,
                        city=da.city,
                        postal_code=da.postal_code,
                        country=da.country,
                        latitude=da.latitude,
                        longitude=da.longitude,
                    )

                assigned_at = None
                if proto_delivery.HasField("assigned_at"):
                    assigned_at = proto_delivery.assigned_at.ToDatetime()

                delivered_at = None
                if proto_delivery.HasField("delivered_at"):
                    delivered_at = proto_delivery.delivered_at.ToDatetime()

                deliveries.append(
                    DeliveryRecord(
                        package_id=proto_delivery.package_id,
                        order_id=proto_delivery.order_id,
                        status=self.PACKAGE_STATUSES.get(proto_delivery.status, "UNSPECIFIED"),
                        pickup_address=pickup,
                        delivery_address=delivery_addr,
                        assigned_at=assigned_at,
                        delivered_at=delivered_at,
                        priority=self.PRIORITIES.get(proto_delivery.priority, "UNSPECIFIED"),
                    )
                )

            return CourierDeliveriesResult(
                deliveries=deliveries,
                total_count=response.total_count,
            )
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.NOT_FOUND:
                return CourierDeliveriesResult(deliveries=[], total_count=0)
            logger.error(f"gRPC error getting courier deliveries: {e}")
            raise DeliveryServiceError(f"Failed to get courier deliveries: {e}")

    def get_courier_pool(
        self,
        status_filter: Optional[list[str]] = None,
        transport_type_filter: Optional[list[str]] = None,
        zone_filter: Optional[str] = None,
        available_only: bool = False,
        include_location: bool = False,
        page: int = 1,
        page_size: int = 20,
    ) -> CourierPoolResult:
        """Get couriers with filtering and pagination.

        Args:
            status_filter: Filter by courier status (e.g., ["FREE", "BUSY"]).
            transport_type_filter: Filter by transport type.
            zone_filter: Filter by work zone.
            available_only: Only return couriers with available capacity.
            include_location: Include current location data.
            page: Page number (1-indexed).
            page_size: Number of items per page (max 100).

        Returns:
            CourierPoolResult with couriers and pagination info.
        """
        self._ensure_connected()

        import delivery_pb2

        # Convert status strings to enum values
        status_values = []
        if status_filter:
            for status in status_filter:
                if status in self.COURIER_STATUS_VALUES:
                    status_values.append(self.COURIER_STATUS_VALUES[status])

        # Convert transport type strings to enum values
        transport_values = []
        if transport_type_filter:
            for transport in transport_type_filter:
                if transport in self.TRANSPORT_TYPE_VALUES:
                    transport_values.append(self.TRANSPORT_TYPE_VALUES[transport])

        request = delivery_pb2.GetCourierPoolRequest(
            status_filter=status_values,
            transport_type_filter=transport_values,
            zone_filter=zone_filter or "",
            available_only=available_only,
            include_location=include_location,
            pagination=delivery_pb2.Pagination(page=page, page_size=page_size),
        )

        try:
            response = self._stub.GetCourierPool(request, metadata=self._get_metadata())
            couriers = [self._proto_to_courier(c) for c in response.couriers]

            return CourierPoolResult(
                couriers=couriers,
                total_count=response.total_count,
                current_page=response.pagination.current_page if response.pagination else page,
                page_size=response.pagination.page_size if response.pagination else page_size,
                total_pages=response.pagination.total_pages if response.pagination else 1,
            )
        except grpc.RpcError as e:
            logger.error(f"gRPC error getting courier pool: {e}")
            raise DeliveryServiceError(f"Failed to get courier pool: {e}")

    def register_courier(
        self,
        name: str,
        phone: str,
        email: str,
        transport_type: str,
        max_distance_km: float,
        work_zone: str,
        work_hours: WorkHours,
        push_token: Optional[str] = None,
    ) -> str:
        """Register a new courier.

        Args:
            name: Courier's full name.
            phone: Phone number in international format.
            email: Email address.
            transport_type: Transport type (WALKING, BICYCLE, MOTORCYCLE, CAR).
            max_distance_km: Maximum delivery distance.
            work_zone: Work zone identifier.
            work_hours: Work hours configuration.
            push_token: Optional push notification token.

        Returns:
            The new courier's ID.
        """
        self._ensure_connected()

        import delivery_pb2

        transport_value = self.TRANSPORT_TYPE_VALUES.get(transport_type, 0)

        request = delivery_pb2.RegisterCourierRequest(
            name=name,
            phone=phone,
            email=email,
            transport_type=transport_value,
            max_distance_km=max_distance_km,
            work_zone=work_zone,
            work_hours=delivery_pb2.WorkHours(
                start_time=work_hours.start_time,
                end_time=work_hours.end_time,
                work_days=work_hours.work_days,
            ),
        )

        if push_token:
            request.push_token = push_token

        try:
            response = self._stub.RegisterCourier(request, metadata=self._get_metadata())
            return response.courier_id
        except grpc.RpcError as e:
            logger.error(f"gRPC error registering courier: {e}")
            raise DeliveryServiceError(f"Failed to register courier: {e}")

    def activate_courier(self, courier_id: str) -> bool:
        """Activate a courier (set status to FREE).

        Args:
            courier_id: The courier's unique identifier.

        Returns:
            True if activation was successful.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.ActivateCourierRequest(courier_id=courier_id)

        try:
            self._stub.ActivateCourier(request, metadata=self._get_metadata())
            return True
        except grpc.RpcError as e:
            logger.error(f"gRPC error activating courier: {e}")
            raise DeliveryServiceError(f"Failed to activate courier: {e}")

    def deactivate_courier(
        self, courier_id: str, reason: Optional[str] = None
    ) -> bool:
        """Deactivate a courier (set status to UNAVAILABLE).

        Args:
            courier_id: The courier's unique identifier.
            reason: Optional reason for deactivation.

        Returns:
            True if deactivation was successful.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.DeactivateCourierRequest(courier_id=courier_id)
        if reason:
            request.reason = reason

        try:
            self._stub.DeactivateCourier(request, metadata=self._get_metadata())
            return True
        except grpc.RpcError as e:
            logger.error(f"gRPC error deactivating courier: {e}")
            raise DeliveryServiceError(f"Failed to deactivate courier: {e}")

    def archive_courier(self, courier_id: str, reason: Optional[str] = None) -> bool:
        """Archive a courier (soft delete).

        Args:
            courier_id: The courier's unique identifier.
            reason: Optional reason for archival.

        Returns:
            True if archival was successful.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.ArchiveCourierRequest(courier_id=courier_id)
        if reason:
            request.reason = reason

        try:
            self._stub.ArchiveCourier(request, metadata=self._get_metadata())
            return True
        except grpc.RpcError as e:
            logger.error(f"gRPC error archiving courier: {e}")
            raise DeliveryServiceError(f"Failed to archive courier: {e}")

    def update_contact_info(
        self,
        courier_id: str,
        phone: Optional[str] = None,
        email: Optional[str] = None,
        push_token: Optional[str] = None,
    ) -> bool:
        """Update courier contact information.

        Args:
            courier_id: The courier's unique identifier.
            phone: New phone number (optional).
            email: New email address (optional).
            push_token: New push token (optional).

        Returns:
            True if update was successful.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.UpdateContactInfoRequest(courier_id=courier_id)
        if phone:
            request.phone = phone
        if email:
            request.email = email
        if push_token:
            request.push_token = push_token

        try:
            self._stub.UpdateContactInfo(request, metadata=self._get_metadata())
            return True
        except grpc.RpcError as e:
            logger.error(f"gRPC error updating contact info: {e}")
            raise DeliveryServiceError(f"Failed to update contact info: {e}")

    def update_work_schedule(
        self,
        courier_id: str,
        work_hours: Optional[WorkHours] = None,
        work_zone: Optional[str] = None,
        max_distance_km: Optional[float] = None,
    ) -> bool:
        """Update courier work schedule.

        Args:
            courier_id: The courier's unique identifier.
            work_hours: New work hours (optional).
            work_zone: New work zone (optional).
            max_distance_km: New max distance (optional).

        Returns:
            True if update was successful.
        """
        self._ensure_connected()

        import delivery_pb2

        request = delivery_pb2.UpdateWorkScheduleRequest(courier_id=courier_id)

        if work_hours:
            request.work_hours.CopyFrom(
                delivery_pb2.WorkHours(
                    start_time=work_hours.start_time,
                    end_time=work_hours.end_time,
                    work_days=work_hours.work_days,
                )
            )
        if work_zone:
            request.work_zone = work_zone
        if max_distance_km is not None:
            request.max_distance_km = max_distance_km

        try:
            self._stub.UpdateWorkSchedule(request, metadata=self._get_metadata())
            return True
        except grpc.RpcError as e:
            logger.error(f"gRPC error updating work schedule: {e}")
            raise DeliveryServiceError(f"Failed to update work schedule: {e}")

    def change_transport_type(self, courier_id: str, transport_type: str) -> int:
        """Change courier transport type.

        Args:
            courier_id: The courier's unique identifier.
            transport_type: New transport type (WALKING, BICYCLE, MOTORCYCLE, CAR).

        Returns:
            New max_load value after recalculation.
        """
        self._ensure_connected()

        import delivery_pb2

        transport_value = self.TRANSPORT_TYPE_VALUES.get(transport_type, 0)

        request = delivery_pb2.ChangeTransportTypeRequest(
            courier_id=courier_id,
            transport_type=transport_value,
        )

        try:
            response = self._stub.ChangeTransportType(request, metadata=self._get_metadata())
            return response.max_load
        except grpc.RpcError as e:
            logger.error(f"gRPC error changing transport type: {e}")
            raise DeliveryServiceError(f"Failed to change transport type: {e}")

    def close(self):
        """Close the gRPC channel."""
        if self._channel:
            self._channel.close()
            self._channel = None
            self._stub = None


# Singleton instance for use across the application
_client: Optional[DeliveryClient] = None


def get_delivery_client(auth_token: Optional[str] = None) -> DeliveryClient:
    """Get a Delivery Service client.

    Args:
        auth_token: JWT token for authentication. If provided, creates a new
                    client with the token. If not, returns the singleton client.

    Returns:
        DeliveryClient instance.
    """
    if auth_token:
        # Create a new client with the auth token for this request
        return DeliveryClient(auth_token=auth_token)

    # Fallback to singleton for backward compatibility (no auth)
    global _client
    if _client is None:
        _client = DeliveryClient()
    return _client


def get_auth_token_from_request(request) -> Optional[str]:
    """Extract JWT token from request Authorization header.

    Oathkeeper injects the JWT via id_token mutator as Authorization header.

    Args:
        request: Django HttpRequest object.

    Returns:
        JWT token string or None if not present.
    """
    auth_header = request.META.get("HTTP_AUTHORIZATION", "")
    if auth_header.startswith("Bearer "):
        return auth_header[7:]  # Remove "Bearer " prefix
    return None
