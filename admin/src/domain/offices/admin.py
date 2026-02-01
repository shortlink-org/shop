"""Define the admin view for the Office model."""

from django.contrib import admin
from mapwidgets import LeafletPointFieldWidget
from unfold.admin import ModelAdmin
from unfold.contrib.filters.admin import BooleanRadioFilter, RangeDateTimeFilter
from unfold.paginator import InfinitePaginator

from .models import Office


@admin.register(Office)
class OfficeAdmin(ModelAdmin):
    """Admin view for Office model with map widget for location selection."""

    list_display = (
        "name",
        "address",
        "working_hours_display",
        "working_days",
        "is_active",
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

    # Show working hours and contact info only when office is active
    conditional_fields = {
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
            "Location",
            {
                "fields": ("address", "location", "latitude_display", "longitude_display"),
                "description": "Click on the map to set the office location",
            },
        ),
        (
            "Working Hours",
            {
                "fields": ("opening_time", "closing_time", "working_days"),
            },
        ),
        (
            "Contact Information",
            {
                "fields": ("phone", "email"),
                "classes": ("collapse",),
            },
        ),
        (
            "Timestamps",
            {
                "fields": ("created_at", "updated_at"),
                "classes": ("collapse",),
            },
        ),
    )

    # Use Leaflet map widget for the location field
    formfield_overrides = {
        # Override PointField to use Leaflet widget
    }

    def get_form(self, request, obj=None, **kwargs):
        """Override form to use Leaflet widget for location field."""
        form = super().get_form(request, obj, **kwargs)
        if "location" in form.base_fields:
            form.base_fields["location"].widget = LeafletPointFieldWidget()
        return form

    @admin.display(description="Working Hours")
    def working_hours_display(self, obj):
        """Display formatted working hours."""
        return obj.working_hours

    @admin.display(description="Latitude")
    def latitude_display(self, obj):
        """Display latitude coordinate."""
        return f"{obj.latitude:.6f}" if obj.latitude else "-"

    @admin.display(description="Longitude")
    def longitude_display(self, obj):
        """Display longitude coordinate."""
        return f"{obj.longitude:.6f}" if obj.longitude else "-"
