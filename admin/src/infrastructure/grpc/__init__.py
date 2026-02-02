"""gRPC client infrastructure.

This package contains gRPC clients for external services.

Note: Delivery/Courier management has been moved to admin-ui.
DeliveryClient was removed. See: admin-ui/ROADMAP.md
"""

from .oms_client import (
    DeliveryAddress,
    DeliveryInfo,
    DeliveryPeriod,
    OmsClient,
    OmsServiceError,
    Order,
    OrderItem,
    OrderListResult,
    OrderNotFoundError,
    OrderStatus,
    get_oms_client,
)

__all__ = [
    # OMS
    "DeliveryAddress",
    "DeliveryInfo",
    "DeliveryPeriod",
    "OmsClient",
    "OmsServiceError",
    "Order",
    "OrderItem",
    "OrderListResult",
    "OrderNotFoundError",
    "OrderStatus",
    "get_oms_client",
]
