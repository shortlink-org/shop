"""gRPC client infrastructure.

This package contains gRPC clients for external services.
"""

from .delivery_client import (
    Address,
    Courier,
    CourierDeliveriesResult,
    CourierPoolResult,
    DeliveryClient,
    DeliveryRecord,
    DeliveryServiceError,
    Location,
    WorkHours,
    get_delivery_client,
)
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
    # Delivery
    "Address",
    "Courier",
    "CourierDeliveriesResult",
    "CourierPoolResult",
    "DeliveryClient",
    "DeliveryRecord",
    "DeliveryServiceError",
    "Location",
    "WorkHours",
    "get_delivery_client",
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
