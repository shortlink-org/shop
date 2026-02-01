"""OMS (Order Management System) gRPC Client.

This module provides a client for communicating with the OMS Service
via gRPC using generated protobuf stubs.
"""

import logging
from dataclasses import dataclass
from datetime import datetime
from decimal import Decimal
from enum import IntEnum
from typing import Optional

import grpc
from django.conf import settings

logger = logging.getLogger(__name__)


class OrderStatus(IntEnum):
    """Order status enum matching OMS proto definitions."""

    UNSPECIFIED = 0
    PENDING = 1
    PROCESSING = 2
    COMPLETED = 3
    CANCELLED = 4


ORDER_STATUS_NAMES = {
    OrderStatus.UNSPECIFIED: "Unspecified",
    OrderStatus.PENDING: "Pending",
    OrderStatus.PROCESSING: "Processing",
    OrderStatus.COMPLETED: "Completed",
    OrderStatus.CANCELLED: "Cancelled",
}


@dataclass
class OrderItem:
    """Order item from OMS."""

    good_id: str
    quantity: int
    price: Decimal


@dataclass
class DeliveryAddress:
    """Delivery address."""

    street: str
    city: str
    postal_code: str
    country: str
    latitude: float
    longitude: float


@dataclass
class DeliveryPeriod:
    """Delivery time window."""

    start_time: Optional[datetime]
    end_time: Optional[datetime]


@dataclass
class DeliveryInfo:
    """Delivery information."""

    pickup_address: Optional[DeliveryAddress]
    delivery_address: Optional[DeliveryAddress]
    delivery_period: Optional[DeliveryPeriod]
    priority: int  # 0=unspecified, 1=normal, 2=urgent


@dataclass
class Order:
    """Order data from OMS."""

    order_id: str
    customer_id: str
    items: list[OrderItem]
    status: OrderStatus
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
    delivery_info: Optional[DeliveryInfo]

    @property
    def status_name(self) -> str:
        """Get human-readable status name."""
        return ORDER_STATUS_NAMES.get(self.status, "Unknown")

    @property
    def total_amount(self) -> Decimal:
        """Calculate total order amount."""
        return sum(item.price * item.quantity for item in self.items)

    @property
    def item_count(self) -> int:
        """Get total item count."""
        return sum(item.quantity for item in self.items)


@dataclass
class OrderListResult:
    """Result of list orders query."""

    orders: list[Order]
    total_count: int
    current_page: int
    page_size: int
    total_pages: int


class OmsServiceError(Exception):
    """Base exception for OMS Service errors."""

    pass


class OrderNotFoundError(OmsServiceError):
    """Order not found."""

    pass


class OmsClient:
    """Client for OMS gRPC API."""

    def __init__(self, host: Optional[str] = None):
        """Initialize the OMS client.

        Args:
            host: gRPC host address. Defaults to settings.OMS_GRPC_HOST.
        """
        self.host = host or getattr(settings, "OMS_GRPC_HOST", "localhost:50052")
        self._channel: Optional[grpc.Channel] = None
        self._stub = None

    def _ensure_connected(self):
        """Ensure gRPC channel is connected."""
        if self._channel is None:
            self._channel = grpc.insecure_channel(self.host)
            # Import generated stubs
            from .generated.infrastructure.rpc.order.v1 import order_rpc_pb2_grpc

            self._stub = order_rpc_pb2_grpc.OrderServiceStub(self._channel)

    def _proto_to_order(self, proto_order) -> Order:
        """Convert protobuf Order to dataclass."""
        items = []
        for proto_item in proto_order.items:
            items.append(
                OrderItem(
                    good_id=proto_item.id,
                    quantity=proto_item.quantity,
                    price=Decimal(str(proto_item.price)),
                )
            )

        delivery_info = None
        if proto_order.HasField("delivery_info"):
            di = proto_order.delivery_info
            pickup = None
            if di.HasField("pickup_address"):
                pa = di.pickup_address
                pickup = DeliveryAddress(
                    street=pa.street,
                    city=pa.city,
                    postal_code=pa.postal_code,
                    country=pa.country,
                    latitude=pa.latitude,
                    longitude=pa.longitude,
                )
            delivery = None
            if di.HasField("delivery_address"):
                da = di.delivery_address
                delivery = DeliveryAddress(
                    street=da.street,
                    city=da.city,
                    postal_code=da.postal_code,
                    country=da.country,
                    latitude=da.latitude,
                    longitude=da.longitude,
                )
            period = None
            if di.HasField("delivery_period"):
                dp = di.delivery_period
                period = DeliveryPeriod(
                    start_time=dp.start_time.ToDatetime() if dp.HasField("start_time") else None,
                    end_time=dp.end_time.ToDatetime() if dp.HasField("end_time") else None,
                )
            delivery_info = DeliveryInfo(
                pickup_address=pickup,
                delivery_address=delivery,
                delivery_period=period,
                priority=di.priority,
            )

        created_at = None
        if proto_order.HasField("created_at"):
            created_at = proto_order.created_at.ToDatetime()

        updated_at = None
        if proto_order.HasField("updated_at"):
            updated_at = proto_order.updated_at.ToDatetime()

        return Order(
            order_id=proto_order.id,
            customer_id=proto_order.customer_id,
            items=items,
            status=OrderStatus(proto_order.status),
            created_at=created_at,
            updated_at=updated_at,
            delivery_info=delivery_info,
        )

    def get_order(self, order_id: str) -> Optional[Order]:
        """Get a single order by ID.

        Args:
            order_id: The order's unique identifier.

        Returns:
            Order data or None if not found.
        """
        self._ensure_connected()

        from .generated.infrastructure.rpc.order.v1.model.v1 import model_pb2

        request = model_pb2.GetRequest(id=order_id)

        try:
            response = self._stub.Get(request)
            return self._proto_to_order(response.order)
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.NOT_FOUND:
                return None
            logger.error(f"gRPC error getting order: {e}")
            raise OmsServiceError(f"Failed to get order: {e}")

    def list_orders(
        self,
        customer_id: Optional[str] = None,
        status_filter: Optional[list[OrderStatus]] = None,
        page: int = 1,
        page_size: int = 20,
    ) -> OrderListResult:
        """List orders with filtering and pagination.

        Args:
            customer_id: Filter by customer ID (optional).
            status_filter: Filter by order status (optional).
            page: Page number (1-indexed).
            page_size: Number of items per page (max 100).

        Returns:
            OrderListResult with orders and pagination info.
        """
        self._ensure_connected()

        from .generated.infrastructure.rpc.order.v1.model.v1 import model_pb2

        # Build request
        pagination = model_pb2.Pagination(page=page, page_size=page_size)
        request = model_pb2.ListRequest(
            customer_id=customer_id or "",
            status_filter=[int(s) for s in (status_filter or [])],
            pagination=pagination,
        )

        try:
            response = self._stub.List(request)
            orders = [self._proto_to_order(o) for o in response.orders]

            return OrderListResult(
                orders=orders,
                total_count=response.total_count,
                current_page=response.pagination.current_page if response.pagination else page,
                page_size=response.pagination.page_size if response.pagination else page_size,
                total_pages=response.pagination.total_pages if response.pagination else 1,
            )
        except grpc.RpcError as e:
            logger.error(f"gRPC error listing orders: {e}")
            raise OmsServiceError(f"Failed to list orders: {e}")

    def cancel_order(self, order_id: str) -> bool:
        """Cancel an order.

        Args:
            order_id: The order's unique identifier.

        Returns:
            True if cancellation was successful.
        """
        self._ensure_connected()

        from .generated.infrastructure.rpc.order.v1.model.v1 import model_pb2

        request = model_pb2.CancelRequest(id=order_id)

        try:
            self._stub.Cancel(request)
            return True
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.NOT_FOUND:
                raise OrderNotFoundError(f"Order not found: {order_id}")
            logger.error(f"gRPC error cancelling order: {e}")
            raise OmsServiceError(f"Failed to cancel order: {e}")

    def close(self):
        """Close the gRPC channel."""
        if self._channel:
            self._channel.close()
            self._channel = None
            self._stub = None


# Singleton instance for use across the application
_client: Optional[OmsClient] = None


def get_oms_client() -> OmsClient:
    """Get the singleton OMS client.

    Returns:
        OmsClient instance.
    """
    global _client
    if _client is None:
        _client = OmsClient()
    return _client
