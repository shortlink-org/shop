# Generated manually for adding HistoricalRecords

import django.contrib.gis.db.models.fields
import django.db.models.deletion
import simple_history.models
from django.conf import settings
from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ("offices", "0001_initial"),
        migrations.swappable_dependency(settings.AUTH_USER_MODEL),
    ]

    operations = [
        migrations.CreateModel(
            name="HistoricalOffice",
            fields=[
                (
                    "id",
                    models.BigIntegerField(
                        auto_created=True, blank=True, db_index=True, verbose_name="ID"
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
                ("created_at", models.DateTimeField(blank=True, editable=False)),
                ("updated_at", models.DateTimeField(blank=True, editable=False)),
                ("history_id", models.AutoField(primary_key=True, serialize=False)),
                ("history_date", models.DateTimeField(db_index=True)),
                ("history_change_reason", models.CharField(max_length=100, null=True)),
                (
                    "history_type",
                    models.CharField(
                        choices=[("+", "Created"), ("~", "Changed"), ("-", "Deleted")],
                        max_length=1,
                    ),
                ),
                (
                    "history_user",
                    models.ForeignKey(
                        null=True,
                        on_delete=django.db.models.deletion.SET_NULL,
                        related_name="+",
                        to=settings.AUTH_USER_MODEL,
                    ),
                ),
            ],
            options={
                "verbose_name": "historical Office",
                "verbose_name_plural": "historical Offices",
                "ordering": ("-history_date", "-history_id"),
                "get_latest_by": ("history_date", "history_id"),
            },
            bases=(simple_history.models.HistoricalChanges, models.Model),
        ),
    ]
