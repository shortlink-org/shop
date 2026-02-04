"""Admin views for order management."""

import logging

from django.contrib import messages
from django.contrib.admin.views.decorators import staff_member_required
from django.shortcuts import redirect, render
from django.views.decorators.http import require_GET, require_http_methods

from infrastructure.grpc import OmsServiceError, OrderStatus, get_oms_client
from .forms import CancelOrderForm, OrderFilterForm

logger = logging.getLogger(__name__)


def _parse_positive_int(value, default, min_value=1, max_value=None):
    """Parse a positive integer from a string with optional bounds."""
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        return default

    if parsed < min_value:
        return min_value
    if max_value is not None and parsed > max_value:
        return max_value
    return parsed


@staff_member_required
@require_GET
def order_list(request):
    """Display list of orders with filtering and pagination."""
    form = OrderFilterForm(request.GET)
    page = _parse_positive_int(request.GET.get("page"), default=1, min_value=1)
    page_size = _parse_positive_int(request.GET.get("page_size"), default=20, min_value=1, max_value=100)

    # Build filters from form
    customer_id = None
    status_filter = None

    if form.is_valid():
        if form.cleaned_data.get("customer_id"):
            customer_id = form.cleaned_data["customer_id"]
        if form.cleaned_data.get("status"):
            try:
                status_value = int(form.cleaned_data["status"])
                status_filter = [OrderStatus(status_value)]
            except (ValueError, KeyError):
                pass

    try:
        client = get_oms_client()
        result = client.list_orders(
            customer_id=customer_id,
            status_filter=status_filter,
            page=page,
            page_size=page_size,
        )
        orders = result.orders
        total_count = result.total_count
        total_pages = result.total_pages
    except OmsServiceError as e:
        logger.error(f"Error fetching orders: {e}")
        messages.error(request, f"Error connecting to OMS Service: {e}")
        orders = []
        total_count = 0
        total_pages = 1

    # Calculate pagination
    has_previous = page > 1
    has_next = page < total_pages
    page_range = range(max(1, page - 2), min(total_pages + 1, page + 3))

    context = {
        "title": "Order Management",
        "orders": orders,
        "filter_form": form,
        "total_count": total_count,
        "page": page,
        "page_size": page_size,
        "total_pages": total_pages,
        "has_previous": has_previous,
        "has_next": has_next,
        "page_range": page_range,
    }

    return render(request, "admin/orders/order_list.html", context)


@staff_member_required
@require_GET
def order_detail(request, order_id):
    """Display order details."""
    try:
        client = get_oms_client()
        order = client.get_order(order_id)

        if not order:
            messages.error(request, f"Order not found: {order_id}")
            return redirect("orders:list")

    except OmsServiceError as e:
        logger.error(f"Error fetching order: {e}")
        messages.error(request, f"Error connecting to OMS Service: {e}")
        return redirect("orders:list")

    context = {
        "title": f"Order: {order.order_id[:8]}...",
        "order": order,
        "cancel_form": CancelOrderForm(),
    }

    return render(request, "admin/orders/order_detail.html", context)


@staff_member_required
@require_http_methods(["POST"])
def order_cancel(request, order_id):
    """Cancel an order."""
    form = CancelOrderForm(request.POST)

    if form.is_valid():
        try:
            client = get_oms_client()
            client.cancel_order(order_id)
            messages.success(request, "Order cancelled successfully.")
            return redirect("orders:detail", order_id=order_id)
        except OmsServiceError as e:
            logger.error(f"Error cancelling order: {e}")
            messages.error(request, f"Error cancelling order: {e}")
    else:
        messages.error(request, "Please confirm the cancellation.")

    return redirect("orders:detail", order_id=order_id)
