"""Forms for courier management."""

from django import forms


class CourierFilterForm(forms.Form):
    """Form for filtering courier list."""

    STATUS_CHOICES = [
        ("", "All Statuses"),
        ("FREE", "Free"),
        ("BUSY", "Busy"),
        ("UNAVAILABLE", "Unavailable"),
        ("ARCHIVED", "Archived"),
    ]

    TRANSPORT_CHOICES = [
        ("", "All Transport"),
        ("WALKING", "Walking"),
        ("BICYCLE", "Bicycle"),
        ("MOTORCYCLE", "Motorcycle"),
        ("CAR", "Car"),
    ]

    status = forms.ChoiceField(
        choices=STATUS_CHOICES,
        required=False,
        widget=forms.Select(attrs={"class": "form-control"}),
    )
    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        required=False,
        widget=forms.Select(attrs={"class": "form-control"}),
    )
    work_zone = forms.CharField(
        max_length=100,
        required=False,
        widget=forms.TextInput(attrs={"class": "form-control", "placeholder": "Work Zone"}),
    )
    available_only = forms.BooleanField(
        required=False,
        widget=forms.CheckboxInput(attrs={"class": "form-check-input"}),
    )


class RegisterCourierForm(forms.Form):
    """Form for registering a new courier."""

    TRANSPORT_CHOICES = [
        ("WALKING", "Walking"),
        ("BICYCLE", "Bicycle"),
        ("MOTORCYCLE", "Motorcycle"),
        ("CAR", "Car"),
    ]

    WEEKDAY_CHOICES = [
        (1, "Monday"),
        (2, "Tuesday"),
        (3, "Wednesday"),
        (4, "Thursday"),
        (5, "Friday"),
        (6, "Saturday"),
        (7, "Sunday"),
    ]

    name = forms.CharField(
        max_length=200,
        widget=forms.TextInput(attrs={"class": "form-control"}),
    )
    phone = forms.CharField(
        max_length=20,
        widget=forms.TextInput(
            attrs={"class": "form-control", "placeholder": "+49123456789"}
        ),
        help_text="Phone number in international format (starting with +)",
    )
    email = forms.EmailField(
        widget=forms.EmailInput(attrs={"class": "form-control"}),
    )
    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        widget=forms.Select(attrs={"class": "form-control"}),
    )
    max_distance_km = forms.FloatField(
        min_value=0.1,
        max_value=100,
        initial=10,
        widget=forms.NumberInput(attrs={"class": "form-control", "step": "0.1"}),
    )
    work_zone = forms.CharField(
        max_length=100,
        widget=forms.TextInput(
            attrs={"class": "form-control", "placeholder": "Berlin-Mitte"}
        ),
    )
    work_start = forms.TimeField(
        initial="09:00",
        widget=forms.TimeInput(attrs={"class": "form-control", "type": "time"}),
    )
    work_end = forms.TimeField(
        initial="18:00",
        widget=forms.TimeInput(attrs={"class": "form-control", "type": "time"}),
    )
    work_days = forms.MultipleChoiceField(
        choices=WEEKDAY_CHOICES,
        initial=[1, 2, 3, 4, 5],
        widget=forms.CheckboxSelectMultiple(),
    )


class UpdateContactInfoForm(forms.Form):
    """Form for updating courier contact information."""

    phone = forms.CharField(
        max_length=20,
        required=False,
        widget=forms.TextInput(attrs={"class": "form-control"}),
    )
    email = forms.EmailField(
        required=False,
        widget=forms.EmailInput(attrs={"class": "form-control"}),
    )

    def clean(self):
        """Validate that at least one field is provided."""
        cleaned_data = super().clean()
        if not cleaned_data.get("phone") and not cleaned_data.get("email"):
            raise forms.ValidationError("At least one field must be provided.")
        return cleaned_data


class UpdateWorkScheduleForm(forms.Form):
    """Form for updating courier work schedule."""

    WEEKDAY_CHOICES = [
        (1, "Monday"),
        (2, "Tuesday"),
        (3, "Wednesday"),
        (4, "Thursday"),
        (5, "Friday"),
        (6, "Saturday"),
        (7, "Sunday"),
    ]

    work_start = forms.TimeField(
        required=False,
        widget=forms.TimeInput(attrs={"class": "form-control", "type": "time"}),
    )
    work_end = forms.TimeField(
        required=False,
        widget=forms.TimeInput(attrs={"class": "form-control", "type": "time"}),
    )
    work_days = forms.MultipleChoiceField(
        choices=WEEKDAY_CHOICES,
        required=False,
        widget=forms.CheckboxSelectMultiple(),
    )
    work_zone = forms.CharField(
        max_length=100,
        required=False,
        widget=forms.TextInput(attrs={"class": "form-control"}),
    )
    max_distance_km = forms.FloatField(
        min_value=0.1,
        max_value=100,
        required=False,
        widget=forms.NumberInput(attrs={"class": "form-control", "step": "0.1"}),
    )


class ChangeTransportTypeForm(forms.Form):
    """Form for changing courier transport type."""

    TRANSPORT_CHOICES = [
        ("WALKING", "Walking (max 1 package)"),
        ("BICYCLE", "Bicycle (max 2 packages)"),
        ("MOTORCYCLE", "Motorcycle (max 3 packages)"),
        ("CAR", "Car (max 5 packages)"),
    ]

    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        widget=forms.Select(attrs={"class": "form-control"}),
    )


class ArchiveCourierForm(forms.Form):
    """Form for archiving a courier."""

    reason = forms.CharField(
        required=False,
        widget=forms.Textarea(
            attrs={"class": "form-control", "rows": 3, "placeholder": "Reason for archival (optional)"}
        ),
    )
    confirm = forms.BooleanField(
        required=True,
        label="I confirm that I want to archive this courier",
        widget=forms.CheckboxInput(attrs={"class": "form-check-input"}),
    )
