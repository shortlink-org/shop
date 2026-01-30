"""Data migration to create initial Berlin offices for testing.

These offices are located in central Berlin to match the courier emulator
which uses the Berlin map from OSRM.
"""

from datetime import time

from django.contrib.gis.geos import Point
from django.db import migrations


def create_berlin_offices(apps, schema_editor):
    """Create two pickup offices in central Berlin."""
    Office = apps.get_model("offices", "Office")

    # Office 1: Near Alexanderplatz (central Berlin landmark)
    # Coordinates: 52.5219, 13.4132
    Office.objects.create(
        name="ShortLink Berlin Mitte",
        address="Alexanderplatz 1, 10178 Berlin, Germany",
        location=Point(13.4132, 52.5219, srid=4326),  # Note: Point takes (lng, lat)
        opening_time=time(8, 0),
        closing_time=time(20, 0),
        working_days="Mon-Sat",
        phone="+49 30 1234567",
        email="mitte@shortlink.shop",
        is_active=True,
    )

    # Office 2: Near Potsdamer Platz (another central Berlin hub)
    # Coordinates: 52.5096, 13.3761
    Office.objects.create(
        name="ShortLink Berlin Potsdamer Platz",
        address="Potsdamer Platz 5, 10785 Berlin, Germany",
        location=Point(13.3761, 52.5096, srid=4326),  # Note: Point takes (lng, lat)
        opening_time=time(9, 0),
        closing_time=time(21, 0),
        working_days="Mon-Sun",
        phone="+49 30 7654321",
        email="potsdamer@shortlink.shop",
        is_active=True,
    )


def remove_berlin_offices(apps, schema_editor):
    """Remove the initial Berlin offices."""
    Office = apps.get_model("offices", "Office")
    Office.objects.filter(
        name__in=["ShortLink Berlin Mitte", "ShortLink Berlin Potsdamer Platz"]
    ).delete()


class Migration(migrations.Migration):
    """Add initial Berlin offices for testing with courier emulator."""

    dependencies = [
        ("offices", "0001_initial"),
    ]

    operations = [
        migrations.RunPython(create_berlin_offices, remove_berlin_offices),
    ]
