"""Define the admin view for the Office model."""

from typing import ClassVar

from django.contrib import admin, messages
from django.http import HttpRequest
from django.shortcuts import redirect
from django.urls import reverse
from django.utils.translation import gettext_lazy as _
from import_export import resources
from import_export.admin import ImportExportModelAdmin
from mapwidgets import LeafletPointFieldWidget
from simple_history.admin import SimpleHistoryAdmin
from unfold.admin import ModelAdmin
from unfold.contrib.filters.admin import BooleanRadioFilter, RangeDateTimeFilter
from unfold.contrib.import_export.forms import ExportForm, ImportForm
from unfold.decorators import action, display
from unfold.paginator import InfinitePaginator

from .models import Office


class OfficeResource(resources.ModelResource):
    """Resource class for import/export of Office model."""

    class Meta:
        model = Office
        fields = (
            "id",
            "name",
            "address",
            "opening_time",
            "closing_time",
            "working_days",
            "phone",
            "email",
            "is_active",
            "created_at",
        )
        export_order = (
            "id",
            "name",
            "address",
            "opening_time",
            "closing_time",
            "working_days",
            "phone",
            "email",
            "is_active",
        )


@admin.register(Office)
class OfficeAdmin(SimpleHistoryAdmin, ModelAdmin, ImportExportModelAdmin):
    """Admin view for Office model with map widget for location selection."""

    # Import/Export configuration
    resource_class = OfficeResource
    import_form_class = ImportForm
    export_form_class = ExportForm

    list_display = (
        "display_office_header",
        "working_hours_display",
        "working_days",
        "display_status",
        "created_at",
    )
    list_filter = (
        ("is_active", BooleanRadioFilter),
        "working_days",
        ("created_at", RangeDateTimeFilter),
    )
    list_filter_submit = True
    search_fields = ("name", "address", "phone", "email")
    ordering = ("name",)
    paginator = InfinitePaginator
    show_full_result_count = False
    actions_row: ClassVar[list[str]] = ["toggle_active", "view_on_map"]

    @action(
        description=_("Toggle Active"),
        permissions=["toggle_active"],
        url_path="toggle-active",
    )
    def toggle_active(self, request: HttpRequest, object_id: int):
        """Toggle the active status of the office."""
        office = Office.objects.get(pk=object_id)
        office.is_active = not office.is_active
        office.save(update_fields=["is_active"])
        status = "activated" if office.is_active else "deactivated"
        messages.success(request, f"Office '{office.name}' {status}.")
        return redirect(reverse("admin:offices_office_changelist"))

    def has_toggle_active_permission(self, request: HttpRequest) -> bool:
        """Check if user can toggle office status."""
        return request.user.has_perm("offices.change_office")

    @action(
        description=_("View on Map"),
        permissions=["view_on_map"],
        url_path="view-on-map",
        attrs={"target": "_blank"},
    )
    def view_on_map(self, request: HttpRequest, object_id: int):
        """Open the office location in Google Maps."""
        office = Office.objects.get(pk=object_id)
        if office.location:
            lat, lng = office.latitude, office.longitude
            return redirect(f"https://www.google.com/maps?q={lat},{lng}")
        messages.warning(request, f"Office '{office.name}' has no location set.")
        return redirect(reverse("admin:offices_office_changelist"))

    def has_view_on_map_permission(self, request: HttpRequest) -> bool:
        """All users with view permission can see on map."""
        return request.user.has_perm("offices.view_office")

    # Show working hours and contact info only when office is active
    conditional_fields: ClassVar[dict] = {
        "opening_time": "is_active == true",
        "closing_time": "is_active == true",
        "working_days": "is_active == true",
        "phone": "is_active == true",
        "email": "is_active == true",
    }
    readonly_fields = ("created_at", "updated_at", "latitude_display", "longitude_display")

    fieldsets = (
        (
            None,
            {
                "fields": ("name", "is_active"),
            },
        ),
        (
            _("Location"),
            {
                "classes": ["tab"],
                "fields": ("address", "location", "latitude_display", "longitude_display"),
                "description": _("Click on the map to set the office location"),
            },
        ),
        (
            _("Working Hours"),
            {
                "classes": ["tab"],
                "fields": ("opening_time", "closing_time", "working_days"),
            },
        ),
        (
            _("Contact"),
            {
                "classes": ["tab"],
                "fields": ("phone", "email"),
            },
        ),
        (
            _("Timestamps"),
            {
                "classes": ["tab"],
                "fields": ("created_at", "updated_at"),
            },
        ),
    )

    # Use Leaflet map widget for the location field
    formfield_overrides: ClassVar[dict] = {
        # Override PointField to use Leaflet widget
    }

    def get_form(self, request, obj=None, **kwargs):
        """Override form to use Leaflet widget for location field."""
        form = super().get_form(request, obj, **kwargs)
        if "location" in form.base_fields:
            form.base_fields["location"].widget = LeafletPointFieldWidget()
        return form

    @display(header=True)
    def display_office_header(self, obj):
        """Display office name with address as two-line header."""
        # Get initials from office name
        words = obj.name.split()
        initials = "".join(w[0].upper() for w in words[:2]) if words else "O"
        address_max_len = 50
        return [
            obj.name,
            obj.address[:address_max_len] + "..." if len(obj.address) > address_max_len else obj.address,
            initials,
        ]

    @display(
        description=_("Status"),
        label={
            True: "success",
            False: "danger",
        },
    )
    def display_status(self, obj):
        """Display active status with colored label."""
        return obj.is_active, _("Active") if obj.is_active else _("Inactive")

    @display(description=_("Working Hours"))
    def working_hours_display(self, obj):
        """Display formatted working hours."""
        return obj.working_hours

    @display(description=_("Latitude"))
    def latitude_display(self, obj):
        """Display latitude coordinate."""
        return f"{obj.latitude:.6f}" if obj.latitude else "-"

    @display(description=_("Longitude"))
    def longitude_display(self, obj):
        """Display longitude coordinate."""
        return f"{obj.longitude:.6f}" if obj.longitude else "-"
