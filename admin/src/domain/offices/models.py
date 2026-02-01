"""Define the Office model for order pickup locations."""

from django.contrib.gis.db import models
from simple_history.models import HistoricalRecords


class Office(models.Model):
    """Office model representing an order pickup location.

    Stores information about physical offices where customers can pick up their orders.
    Uses GeoDjango PointField for precise geolocation.
    """

    name = models.CharField(
        max_length=255,
        verbose_name="Office Name",
        help_text="Display name of the office",
    )
    address = models.TextField(
        verbose_name="Address",
        help_text="Full street address of the office",
    )
    location = models.PointField(
        verbose_name="Location",
        help_text="Geographic coordinates (latitude, longitude)",
        srid=4326,  # WGS84 coordinate system
    )

    # Working hours
    opening_time = models.TimeField(
        verbose_name="Opening Time",
        help_text="Time when the office opens",
    )
    closing_time = models.TimeField(
        verbose_name="Closing Time",
        help_text="Time when the office closes",
    )
    working_days = models.CharField(
        max_length=100,
        verbose_name="Working Days",
        help_text="Days of operation (e.g., 'Mon-Fri', 'Mon-Sat')",
        default="Mon-Fri",
    )

    # Contact info
    phone = models.CharField(
        max_length=50,
        verbose_name="Phone",
        help_text="Contact phone number",
        blank=True,
    )
    email = models.EmailField(
        verbose_name="Email",
        help_text="Contact email address",
        blank=True,
    )

    # Status
    is_active = models.BooleanField(
        default=True,
        verbose_name="Active",
        help_text="Whether the office is currently operational",
    )

    # Timestamps
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    # History tracking
    history = HistoricalRecords()

    class Meta:
        """Meta options for Office model."""

        verbose_name = "Office"
        verbose_name_plural = "Offices"
        ordering = ["name"]

    def __str__(self):
        """Return the name of the office."""
        return self.name

    @property
    def working_hours(self) -> str:
        """Return formatted working hours string."""
        return f"{self.opening_time.strftime('%H:%M')} - {self.closing_time.strftime('%H:%M')}"

    @property
    def latitude(self) -> float:
        """Return latitude of the office location."""
        return self.location.y if self.location else None

    @property
    def longitude(self) -> float:
        """Return longitude of the office location."""
        return self.location.x if self.location else None
