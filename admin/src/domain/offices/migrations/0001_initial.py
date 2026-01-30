"""Initial migration for Office model."""

import django.contrib.gis.db.models.fields
from django.db import migrations, models


class Migration(migrations.Migration):
    """Create Office model with GeoDjango PointField."""

    initial = True

    dependencies = []

    operations = [
        migrations.CreateModel(
            name="Office",
            fields=[
                (
                    "id",
                    models.BigAutoField(
                        auto_created=True,
                        primary_key=True,
                        serialize=False,
                        verbose_name="ID",
                    ),
                ),
                (
                    "name",
                    models.CharField(
                        help_text="Display name of the office",
                        max_length=255,
                        verbose_name="Office Name",
                    ),
                ),
                (
                    "address",
                    models.TextField(
                        help_text="Full street address of the office",
                        verbose_name="Address",
                    ),
                ),
                (
                    "location",
                    django.contrib.gis.db.models.fields.PointField(
                        help_text="Geographic coordinates (latitude, longitude)",
                        srid=4326,
                        verbose_name="Location",
                    ),
                ),
                (
                    "opening_time",
                    models.TimeField(
                        help_text="Time when the office opens",
                        verbose_name="Opening Time",
                    ),
                ),
                (
                    "closing_time",
                    models.TimeField(
                        help_text="Time when the office closes",
                        verbose_name="Closing Time",
                    ),
                ),
                (
                    "working_days",
                    models.CharField(
                        default="Mon-Fri",
                        help_text="Days of operation (e.g., 'Mon-Fri', 'Mon-Sat')",
                        max_length=100,
                        verbose_name="Working Days",
                    ),
                ),
                (
                    "phone",
                    models.CharField(
                        blank=True,
                        help_text="Contact phone number",
                        max_length=50,
                        verbose_name="Phone",
                    ),
                ),
                (
                    "email",
                    models.EmailField(
                        blank=True,
                        help_text="Contact email address",
                        max_length=254,
                        verbose_name="Email",
                    ),
                ),
                (
                    "is_active",
                    models.BooleanField(
                        default=True,
                        help_text="Whether the office is currently operational",
                        verbose_name="Active",
                    ),
                ),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("updated_at", models.DateTimeField(auto_now=True)),
            ],
            options={
                "verbose_name": "Office",
                "verbose_name_plural": "Offices",
                "ordering": ["name"],
            },
        ),
    ]
