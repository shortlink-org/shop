"""Forms for courier management with Crispy Forms and Unfold styling."""

from crispy_forms.helper import FormHelper
from crispy_forms.layout import Div, Field, Fieldset, Layout
from django import forms
from django.utils.translation import gettext_lazy as _
from unfold.widgets import (
    UnfoldAdminEmailInputWidget,
    UnfoldAdminSelectWidget,
    UnfoldAdminTextInputWidget,
    UnfoldBooleanSwitchWidget,
)


class CourierFilterForm(forms.Form):
    """Form for filtering courier list."""

    STATUS_CHOICES = [
        ("", _("All Statuses")),
        ("FREE", _("Free")),
        ("BUSY", _("Busy")),
        ("UNAVAILABLE", _("Unavailable")),
        ("ARCHIVED", _("Archived")),
    ]

    TRANSPORT_CHOICES = [
        ("", _("All Transport")),
        ("WALKING", _("Walking")),
        ("BICYCLE", _("Bicycle")),
        ("MOTORCYCLE", _("Motorcycle")),
        ("CAR", _("Car")),
    ]

    status = forms.ChoiceField(
        choices=STATUS_CHOICES,
        required=False,
        label=_("Status"),
        widget=UnfoldAdminSelectWidget(),
    )
    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        required=False,
        label=_("Transport Type"),
        widget=UnfoldAdminSelectWidget(),
    )
    work_zone = forms.CharField(
        max_length=100,
        required=False,
        label=_("Work Zone"),
        widget=UnfoldAdminTextInputWidget(attrs={"placeholder": _("Work Zone")}),
    )
    available_only = forms.BooleanField(
        required=False,
        label=_("Available Only"),
        widget=UnfoldBooleanSwitchWidget(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_method = "get"
        self.helper.form_class = "flex flex-wrap gap-4 items-end"
        self.helper.layout = Layout(
            Div(
                Field("status", wrapper_class="w-40"),
                Field("transport_type", wrapper_class="w-40"),
                Field("work_zone", wrapper_class="w-48"),
                Field("available_only", wrapper_class="w-32"),
                css_class="flex flex-wrap gap-4 items-end",
            ),
        )


class RegisterCourierForm(forms.Form):
    """Form for registering a new courier."""

    TRANSPORT_CHOICES = [
        ("WALKING", _("Walking")),
        ("BICYCLE", _("Bicycle")),
        ("MOTORCYCLE", _("Motorcycle")),
        ("CAR", _("Car")),
    ]

    WEEKDAY_CHOICES = [
        (1, _("Monday")),
        (2, _("Tuesday")),
        (3, _("Wednesday")),
        (4, _("Thursday")),
        (5, _("Friday")),
        (6, _("Saturday")),
        (7, _("Sunday")),
    ]

    name = forms.CharField(
        max_length=200,
        label=_("Name"),
        widget=UnfoldAdminTextInputWidget(),
    )
    phone = forms.CharField(
        max_length=20,
        label=_("Phone"),
        help_text=_("Phone number in international format (starting with +)"),
        widget=UnfoldAdminTextInputWidget(attrs={"placeholder": "+49123456789"}),
    )
    email = forms.EmailField(
        label=_("Email"),
        widget=UnfoldAdminEmailInputWidget(),
    )
    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        label=_("Transport Type"),
        widget=UnfoldAdminSelectWidget(),
    )
    max_distance_km = forms.FloatField(
        min_value=0.1,
        max_value=100,
        initial=10,
        label=_("Max Distance (km)"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "number", "step": "0.1"}),
    )
    work_zone = forms.CharField(
        max_length=100,
        label=_("Work Zone"),
        widget=UnfoldAdminTextInputWidget(attrs={"placeholder": "Berlin-Mitte"}),
    )
    work_start = forms.TimeField(
        initial="09:00",
        label=_("Work Start"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "time"}),
    )
    work_end = forms.TimeField(
        initial="18:00",
        label=_("Work End"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "time"}),
    )
    work_days = forms.MultipleChoiceField(
        choices=WEEKDAY_CHOICES,
        initial=[1, 2, 3, 4, 5],
        label=_("Work Days"),
        widget=forms.CheckboxSelectMultiple(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Fieldset(
                _("Personal Information"),
                Div(
                    Field("name"),
                    Field("phone"),
                    Field("email"),
                    css_class="grid grid-cols-1 md:grid-cols-3 gap-4",
                ),
            ),
            Fieldset(
                _("Transport & Zone"),
                Div(
                    Field("transport_type"),
                    Field("max_distance_km"),
                    Field("work_zone"),
                    css_class="grid grid-cols-1 md:grid-cols-3 gap-4",
                ),
            ),
            Fieldset(
                _("Work Schedule"),
                Div(
                    Field("work_start"),
                    Field("work_end"),
                    css_class="grid grid-cols-1 md:grid-cols-2 gap-4",
                ),
                Field("work_days"),
            ),
        )


class UpdateContactInfoForm(forms.Form):
    """Form for updating courier contact information."""

    phone = forms.CharField(
        max_length=20,
        required=False,
        label=_("Phone"),
        widget=UnfoldAdminTextInputWidget(),
    )
    email = forms.EmailField(
        required=False,
        label=_("Email"),
        widget=UnfoldAdminEmailInputWidget(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Div(
                Field("phone"),
                Field("email"),
                css_class="grid grid-cols-1 md:grid-cols-2 gap-4",
            ),
        )

    def clean(self):
        """Validate that at least one field is provided."""
        cleaned_data = super().clean()
        if not cleaned_data.get("phone") and not cleaned_data.get("email"):
            raise forms.ValidationError(_("At least one field must be provided."))
        return cleaned_data


class UpdateWorkScheduleForm(forms.Form):
    """Form for updating courier work schedule."""

    WEEKDAY_CHOICES = [
        (1, _("Monday")),
        (2, _("Tuesday")),
        (3, _("Wednesday")),
        (4, _("Thursday")),
        (5, _("Friday")),
        (6, _("Saturday")),
        (7, _("Sunday")),
    ]

    work_start = forms.TimeField(
        required=False,
        label=_("Work Start"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "time"}),
    )
    work_end = forms.TimeField(
        required=False,
        label=_("Work End"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "time"}),
    )
    work_days = forms.MultipleChoiceField(
        choices=WEEKDAY_CHOICES,
        required=False,
        label=_("Work Days"),
        widget=forms.CheckboxSelectMultiple(),
    )
    work_zone = forms.CharField(
        max_length=100,
        required=False,
        label=_("Work Zone"),
        widget=UnfoldAdminTextInputWidget(),
    )
    max_distance_km = forms.FloatField(
        min_value=0.1,
        max_value=100,
        required=False,
        label=_("Max Distance (km)"),
        widget=UnfoldAdminTextInputWidget(attrs={"type": "number", "step": "0.1"}),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Fieldset(
                _("Working Hours"),
                Div(
                    Field("work_start"),
                    Field("work_end"),
                    css_class="grid grid-cols-1 md:grid-cols-2 gap-4",
                ),
                Field("work_days"),
            ),
            Fieldset(
                _("Work Area"),
                Div(
                    Field("work_zone"),
                    Field("max_distance_km"),
                    css_class="grid grid-cols-1 md:grid-cols-2 gap-4",
                ),
            ),
        )


class ChangeTransportTypeForm(forms.Form):
    """Form for changing courier transport type."""

    TRANSPORT_CHOICES = [
        ("WALKING", _("Walking (max 1 package)")),
        ("BICYCLE", _("Bicycle (max 2 packages)")),
        ("MOTORCYCLE", _("Motorcycle (max 3 packages)")),
        ("CAR", _("Car (max 5 packages)")),
    ]

    transport_type = forms.ChoiceField(
        choices=TRANSPORT_CHOICES,
        label=_("Transport Type"),
        widget=UnfoldAdminSelectWidget(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Field("transport_type"),
        )


class ArchiveCourierForm(forms.Form):
    """Form for archiving a courier."""

    reason = forms.CharField(
        required=False,
        label=_("Reason"),
        widget=forms.Textarea(attrs={"rows": 3, "placeholder": _("Reason for archival (optional)")}),
    )
    confirm = forms.BooleanField(
        required=True,
        label=_("I confirm that I want to archive this courier"),
        widget=UnfoldBooleanSwitchWidget(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Field("reason"),
            Field("confirm"),
        )
