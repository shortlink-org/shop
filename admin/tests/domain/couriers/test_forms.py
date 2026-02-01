"""Tests for courier forms."""

import pytest

from domain.couriers.forms import (
    ArchiveCourierForm,
    ChangeTransportTypeForm,
    CourierFilterForm,
    RegisterCourierForm,
    UpdateContactInfoForm,
    UpdateWorkScheduleForm,
)


class TestCourierFilterForm:
    """Tests for CourierFilterForm."""

    def test_valid_empty_form(self):
        """Empty form should be valid (all fields optional)."""
        form = CourierFilterForm(data={})
        assert form.is_valid()

    def test_valid_with_status(self):
        """Form with valid status should be valid."""
        form = CourierFilterForm(data={"status": "FREE"})
        assert form.is_valid()
        assert form.cleaned_data["status"] == "FREE"

    def test_valid_with_all_statuses(self):
        """All status choices should be valid."""
        for status, _ in CourierFilterForm.STATUS_CHOICES:
            form = CourierFilterForm(data={"status": status})
            assert form.is_valid(), f"Status {status} should be valid"

    def test_valid_with_transport_type(self):
        """Form with valid transport type should be valid."""
        form = CourierFilterForm(data={"transport_type": "BICYCLE"})
        assert form.is_valid()
        assert form.cleaned_data["transport_type"] == "BICYCLE"

    def test_valid_with_all_transport_types(self):
        """All transport type choices should be valid."""
        for transport, _ in CourierFilterForm.TRANSPORT_CHOICES:
            form = CourierFilterForm(data={"transport_type": transport})
            assert form.is_valid(), f"Transport {transport} should be valid"

    def test_valid_with_work_zone(self):
        """Form with work zone should be valid."""
        form = CourierFilterForm(data={"work_zone": "Berlin-Mitte"})
        assert form.is_valid()
        assert form.cleaned_data["work_zone"] == "Berlin-Mitte"

    def test_valid_with_available_only(self):
        """Form with available_only flag should be valid."""
        form = CourierFilterForm(data={"available_only": True})
        assert form.is_valid()
        assert form.cleaned_data["available_only"] is True

    def test_valid_with_all_filters(self):
        """Form with all filters should be valid."""
        form = CourierFilterForm(
            data={
                "status": "FREE",
                "transport_type": "CAR",
                "work_zone": "Berlin-Kreuzberg",
                "available_only": True,
            }
        )
        assert form.is_valid()

    def test_invalid_status_choice(self):
        """Invalid status choice should fail."""
        form = CourierFilterForm(data={"status": "INVALID"})
        assert not form.is_valid()
        assert "status" in form.errors


class TestRegisterCourierForm:
    """Tests for RegisterCourierForm."""

    def test_valid_form(self):
        """Valid form with all required fields."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert form.is_valid(), form.errors

    def test_missing_name(self):
        """Form without name should be invalid."""
        form = RegisterCourierForm(
            data={
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "name" in form.errors

    def test_missing_email(self):
        """Form without email should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "transport_type": "BICYCLE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "email" in form.errors

    def test_invalid_email(self):
        """Form with invalid email should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "not-an-email",
                "transport_type": "BICYCLE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "email" in form.errors

    def test_invalid_transport_type(self):
        """Form with invalid transport type should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "AIRPLANE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "transport_type" in form.errors

    def test_max_distance_too_small(self):
        """Form with max_distance below min should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": 0.05,  # Below 0.1 minimum
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "max_distance_km" in form.errors

    def test_max_distance_too_large(self):
        """Form with max_distance above max should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": 150.0,  # Above 100 maximum
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
                "work_days": [1, 2, 3, 4, 5],
            }
        )
        assert not form.is_valid()
        assert "max_distance_km" in form.errors

    def test_missing_work_days(self):
        """Form without work_days should be invalid."""
        form = RegisterCourierForm(
            data={
                "name": "John Doe",
                "phone": "+49123456789",
                "email": "john@example.com",
                "transport_type": "BICYCLE",
                "max_distance_km": 10.0,
                "work_zone": "Berlin-Mitte",
                "work_start": "09:00",
                "work_end": "18:00",
            }
        )
        assert not form.is_valid()
        assert "work_days" in form.errors

    def test_all_transport_types_valid(self):
        """All transport types should be valid choices."""
        for transport, _ in RegisterCourierForm.TRANSPORT_CHOICES:
            form = RegisterCourierForm(
                data={
                    "name": "John Doe",
                    "phone": "+49123456789",
                    "email": "john@example.com",
                    "transport_type": transport,
                    "max_distance_km": 10.0,
                    "work_zone": "Berlin-Mitte",
                    "work_start": "09:00",
                    "work_end": "18:00",
                    "work_days": [1, 2, 3],
                }
            )
            assert form.is_valid(), f"Transport {transport} should be valid"


class TestUpdateContactInfoForm:
    """Tests for UpdateContactInfoForm."""

    def test_valid_with_phone_only(self):
        """Form with only phone should be valid."""
        form = UpdateContactInfoForm(data={"phone": "+49111222333"})
        assert form.is_valid()

    def test_valid_with_email_only(self):
        """Form with only email should be valid."""
        form = UpdateContactInfoForm(data={"email": "new@example.com"})
        assert form.is_valid()

    def test_valid_with_both(self):
        """Form with both phone and email should be valid."""
        form = UpdateContactInfoForm(
            data={"phone": "+49111222333", "email": "new@example.com"}
        )
        assert form.is_valid()

    def test_invalid_empty_form(self):
        """Empty form should be invalid (at least one field required)."""
        form = UpdateContactInfoForm(data={})
        assert not form.is_valid()
        assert "__all__" in form.errors or form.non_field_errors()

    def test_invalid_email_format(self):
        """Form with invalid email should be invalid."""
        form = UpdateContactInfoForm(data={"email": "not-an-email"})
        assert not form.is_valid()
        assert "email" in form.errors


class TestUpdateWorkScheduleForm:
    """Tests for UpdateWorkScheduleForm."""

    def test_valid_empty_form(self):
        """Empty form should be valid (all fields optional)."""
        form = UpdateWorkScheduleForm(data={})
        assert form.is_valid()

    def test_valid_with_work_times(self):
        """Form with work times should be valid."""
        form = UpdateWorkScheduleForm(
            data={"work_start": "08:00", "work_end": "20:00"}
        )
        assert form.is_valid()

    def test_valid_with_work_zone(self):
        """Form with work zone should be valid."""
        form = UpdateWorkScheduleForm(data={"work_zone": "Berlin-Kreuzberg"})
        assert form.is_valid()

    def test_valid_with_max_distance(self):
        """Form with max distance should be valid."""
        form = UpdateWorkScheduleForm(data={"max_distance_km": 15.0})
        assert form.is_valid()

    def test_invalid_max_distance_too_small(self):
        """Form with max distance below min should be invalid."""
        form = UpdateWorkScheduleForm(data={"max_distance_km": 0.05})
        assert not form.is_valid()
        assert "max_distance_km" in form.errors

    def test_valid_with_work_days(self):
        """Form with work days should be valid."""
        form = UpdateWorkScheduleForm(data={"work_days": [1, 2, 3]})
        assert form.is_valid()
        assert form.cleaned_data["work_days"] == ["1", "2", "3"]


class TestChangeTransportTypeForm:
    """Tests for ChangeTransportTypeForm."""

    def test_valid_walking(self):
        """Walking transport type should be valid."""
        form = ChangeTransportTypeForm(data={"transport_type": "WALKING"})
        assert form.is_valid()

    def test_valid_bicycle(self):
        """Bicycle transport type should be valid."""
        form = ChangeTransportTypeForm(data={"transport_type": "BICYCLE"})
        assert form.is_valid()

    def test_valid_motorcycle(self):
        """Motorcycle transport type should be valid."""
        form = ChangeTransportTypeForm(data={"transport_type": "MOTORCYCLE"})
        assert form.is_valid()

    def test_valid_car(self):
        """Car transport type should be valid."""
        form = ChangeTransportTypeForm(data={"transport_type": "CAR"})
        assert form.is_valid()

    def test_invalid_transport_type(self):
        """Invalid transport type should fail."""
        form = ChangeTransportTypeForm(data={"transport_type": "HELICOPTER"})
        assert not form.is_valid()
        assert "transport_type" in form.errors

    def test_missing_transport_type(self):
        """Missing transport type should fail."""
        form = ChangeTransportTypeForm(data={})
        assert not form.is_valid()
        assert "transport_type" in form.errors


class TestArchiveCourierForm:
    """Tests for ArchiveCourierForm."""

    def test_valid_with_confirm(self):
        """Form with confirm checked should be valid."""
        form = ArchiveCourierForm(data={"confirm": True})
        assert form.is_valid()

    def test_valid_with_reason_and_confirm(self):
        """Form with reason and confirm should be valid."""
        form = ArchiveCourierForm(
            data={"reason": "Courier resigned", "confirm": True}
        )
        assert form.is_valid()
        assert form.cleaned_data["reason"] == "Courier resigned"

    def test_invalid_without_confirm(self):
        """Form without confirm should be invalid."""
        form = ArchiveCourierForm(data={"reason": "Courier resigned"})
        assert not form.is_valid()
        assert "confirm" in form.errors

    def test_invalid_with_confirm_false(self):
        """Form with confirm=False should be invalid."""
        form = ArchiveCourierForm(data={"confirm": False})
        assert not form.is_valid()
        assert "confirm" in form.errors
