"""Admin views for courier management."""

import logging

from django.contrib import messages
from django.contrib.admin.views.decorators import staff_member_required
from django.shortcuts import redirect, render
from django.views.decorators.http import require_GET, require_http_methods

from infrastructure.grpc import DeliveryServiceError, WorkHours, get_delivery_client

from .forms import (
    ArchiveCourierForm,
    ChangeTransportTypeForm,
    CourierFilterForm,
    RegisterCourierForm,
    UpdateContactInfoForm,
    UpdateWorkScheduleForm,
)

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
def courier_list(request):
    """Display list of couriers with filtering and pagination."""
    form = CourierFilterForm(request.GET)
    page = _parse_positive_int(request.GET.get("page"), default=1, min_value=1)
    page_size = _parse_positive_int(request.GET.get("page_size"), default=20, min_value=1, max_value=100)

    # Build filters from form
    status_filter = None
    transport_filter = None
    zone_filter = None
    available_only = False

    if form.is_valid():
        if form.cleaned_data.get("status"):
            status_filter = [form.cleaned_data["status"]]
        if form.cleaned_data.get("transport_type"):
            transport_filter = [form.cleaned_data["transport_type"]]
        zone_filter = form.cleaned_data.get("work_zone")
        available_only = form.cleaned_data.get("available_only", False)

    try:
        client = get_delivery_client()
        result = client.get_courier_pool(
            status_filter=status_filter,
            transport_type_filter=transport_filter,
            zone_filter=zone_filter,
            available_only=available_only,
            include_location=True,
            page=page,
            page_size=page_size,
        )
        couriers = result.couriers
        total_count = result.total_count
        total_pages = result.total_pages
    except DeliveryServiceError as e:
        logger.error(f"Error fetching couriers: {e}")
        messages.error(request, f"Error connecting to Delivery Service: {e}")
        couriers = []
        total_count = 0
        total_pages = 1

    # Calculate pagination
    has_previous = page > 1
    has_next = page < total_pages
    page_range = range(max(1, page - 2), min(total_pages + 1, page + 3))

    context = {
        "title": "Courier Management",
        "couriers": couriers,
        "filter_form": form,
        "total_count": total_count,
        "page": page,
        "page_size": page_size,
        "total_pages": total_pages,
        "has_previous": has_previous,
        "has_next": has_next,
        "page_range": page_range,
    }

    return render(request, "admin/couriers/courier_list.html", context)


@staff_member_required
@require_GET
def courier_detail(request, courier_id):
    """Display courier details."""
    try:
        client = get_delivery_client()
        courier = client.get_courier(courier_id, include_location=True)

        if not courier:
            messages.error(request, f"Courier not found: {courier_id}")
            return redirect("couriers:list")

        # Fetch recent deliveries for this courier
        deliveries_result = client.get_courier_deliveries(courier_id, limit=5)

    except DeliveryServiceError as e:
        logger.error(f"Error fetching courier: {e}")
        messages.error(request, f"Error connecting to Delivery Service: {e}")
        return redirect("couriers:list")

    context = {
        "title": f"Courier: {courier.name}",
        "courier": courier,
        "recent_deliveries": deliveries_result.deliveries,
        "total_deliveries": deliveries_result.total_count,
        "update_contact_form": UpdateContactInfoForm(
            initial={"phone": courier.phone, "email": courier.email}
        ),
        "update_schedule_form": UpdateWorkScheduleForm(
            initial={
                "work_zone": courier.work_zone,
                "max_distance_km": courier.max_distance_km,
                "work_start": courier.work_hours.start_time if courier.work_hours else None,
                "work_end": courier.work_hours.end_time if courier.work_hours else None,
                "work_days": courier.work_hours.work_days if courier.work_hours else [],
            }
        ),
        "change_transport_form": ChangeTransportTypeForm(
            initial={"transport_type": courier.transport_type}
        ),
        "archive_form": ArchiveCourierForm(),
    }

    return render(request, "admin/couriers/courier_detail.html", context)


@staff_member_required
@require_http_methods(["GET", "POST"])
def courier_register(request):
    """Register a new courier."""
    if request.method == "POST":
        form = RegisterCourierForm(request.POST)
        if form.is_valid():
            try:
                client = get_delivery_client()
                work_hours = WorkHours(
                    start_time=form.cleaned_data["work_start"].strftime("%H:%M"),
                    end_time=form.cleaned_data["work_end"].strftime("%H:%M"),
                    work_days=[int(d) for d in form.cleaned_data["work_days"]],
                )

                courier_id = client.register_courier(
                    name=form.cleaned_data["name"],
                    phone=form.cleaned_data["phone"],
                    email=form.cleaned_data["email"],
                    transport_type=form.cleaned_data["transport_type"],
                    max_distance_km=form.cleaned_data["max_distance_km"],
                    work_zone=form.cleaned_data["work_zone"],
                    work_hours=work_hours,
                )

                messages.success(request, f"Courier registered successfully: {courier_id}")
                return redirect("couriers:detail", courier_id=courier_id)

            except DeliveryServiceError as e:
                logger.error(f"Error registering courier: {e}")
                messages.error(request, f"Error registering courier: {e}")
    else:
        form = RegisterCourierForm()

    context = {
        "title": "Register New Courier",
        "form": form,
    }

    return render(request, "admin/couriers/courier_register.html", context)


@staff_member_required
@require_http_methods(["POST"])
def courier_activate(request, courier_id):
    """Activate a courier."""
    try:
        client = get_delivery_client()
        client.activate_courier(courier_id)
        messages.success(request, "Courier activated successfully.")
    except DeliveryServiceError as e:
        logger.error(f"Error activating courier: {e}")
        messages.error(request, f"Error activating courier: {e}")

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_http_methods(["POST"])
def courier_deactivate(request, courier_id):
    """Deactivate a courier."""
    reason = request.POST.get("reason")

    try:
        client = get_delivery_client()
        client.deactivate_courier(courier_id, reason=reason)
        messages.success(request, "Courier deactivated successfully.")
    except DeliveryServiceError as e:
        logger.error(f"Error deactivating courier: {e}")
        messages.error(request, f"Error deactivating courier: {e}")

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_http_methods(["POST"])
def courier_archive(request, courier_id):
    """Archive a courier."""
    form = ArchiveCourierForm(request.POST)

    if form.is_valid():
        try:
            client = get_delivery_client()
            client.archive_courier(courier_id, reason=form.cleaned_data.get("reason"))
            messages.success(request, "Courier archived successfully.")
            return redirect("couriers:list")
        except DeliveryServiceError as e:
            logger.error(f"Error archiving courier: {e}")
            messages.error(request, f"Error archiving courier: {e}")
    else:
        messages.error(request, "Please confirm the archival.")

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_http_methods(["POST"])
def courier_update_contact(request, courier_id):
    """Update courier contact information."""
    form = UpdateContactInfoForm(request.POST)

    if form.is_valid():
        try:
            client = get_delivery_client()
            client.update_contact_info(
                courier_id=courier_id,
                phone=form.cleaned_data.get("phone"),
                email=form.cleaned_data.get("email"),
            )
            messages.success(request, "Contact information updated successfully.")
        except DeliveryServiceError as e:
            logger.error(f"Error updating contact info: {e}")
            messages.error(request, f"Error updating contact info: {e}")
    else:
        for error in form.non_field_errors():
            messages.error(request, error)

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_http_methods(["POST"])
def courier_update_schedule(request, courier_id):
    """Update courier work schedule."""
    form = UpdateWorkScheduleForm(request.POST)

    if form.is_valid():
        try:
            client = get_delivery_client()

            work_hours = None
            if form.cleaned_data.get("work_start") and form.cleaned_data.get("work_end"):
                work_hours = WorkHours(
                    start_time=form.cleaned_data["work_start"].strftime("%H:%M"),
                    end_time=form.cleaned_data["work_end"].strftime("%H:%M"),
                    work_days=[int(d) for d in form.cleaned_data.get("work_days", [])],
                )

            client.update_work_schedule(
                courier_id=courier_id,
                work_hours=work_hours,
                work_zone=form.cleaned_data.get("work_zone"),
                max_distance_km=form.cleaned_data.get("max_distance_km"),
            )
            messages.success(request, "Work schedule updated successfully.")
        except DeliveryServiceError as e:
            logger.error(f"Error updating work schedule: {e}")
            messages.error(request, f"Error updating work schedule: {e}")

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_http_methods(["POST"])
def courier_change_transport(request, courier_id):
    """Change courier transport type."""
    form = ChangeTransportTypeForm(request.POST)

    if form.is_valid():
        try:
            client = get_delivery_client()
            new_max_load = client.change_transport_type(
                courier_id=courier_id,
                transport_type=form.cleaned_data["transport_type"],
            )
            messages.success(
                request,
                f"Transport type changed successfully. New max load: {new_max_load} packages.",
            )
        except DeliveryServiceError as e:
            logger.error(f"Error changing transport type: {e}")
            messages.error(request, f"Error changing transport type: {e}")

    return redirect("couriers:detail", courier_id=courier_id)


@staff_member_required
@require_GET
def courier_map(request):
    """Display map with courier locations."""
    try:
        client = get_delivery_client()
        # Get all active couriers with locations
        result = client.get_courier_pool(
            status_filter=["FREE", "BUSY"],
            include_location=True,
            page=1,
            page_size=100,  # Get more for map view
        )
        couriers = [c for c in result.couriers if c.current_location]
    except DeliveryServiceError as e:
        logger.error(f"Error fetching couriers for map: {e}")
        messages.error(request, f"Error connecting to Delivery Service: {e}")
        couriers = []

    context = {
        "title": "Courier Map",
        "couriers": couriers,
    }

    return render(request, "admin/couriers/courier_map.html", context)
