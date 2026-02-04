"""Forms for order management with Crispy Forms and Unfold styling."""

from typing import ClassVar

from crispy_forms.helper import FormHelper
from crispy_forms.layout import Div, Field, Layout
from django import forms
from django.utils.translation import gettext_lazy as _
from unfold.widgets import (
    UnfoldAdminSelectWidget,
    UnfoldAdminTextInputWidget,
    UnfoldBooleanSwitchWidget,
)


class OrderFilterForm(forms.Form):
    """Form for filtering order list."""

    STATUS_CHOICES: ClassVar[list[tuple[str, str]]] = [
        ("", _("All Statuses")),
        ("1", _("Pending")),
        ("2", _("Processing")),
        ("3", _("Completed")),
        ("4", _("Cancelled")),
    ]

    status = forms.ChoiceField(
        choices=STATUS_CHOICES,
        required=False,
        label=_("Status"),
        widget=UnfoldAdminSelectWidget(),
    )
    customer_id = forms.CharField(
        max_length=100,
        required=False,
        label=_("Customer ID"),
        widget=UnfoldAdminTextInputWidget(attrs={"placeholder": _("Customer ID")}),
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
                Field("customer_id", wrapper_class="w-48"),
                css_class="flex flex-wrap gap-4 items-end",
            ),
        )


class CancelOrderForm(forms.Form):
    """Form for cancelling an order."""

    confirm = forms.BooleanField(
        required=True,
        label=_("I confirm that I want to cancel this order"),
        widget=UnfoldBooleanSwitchWidget(),
    )

    def __init__(self, *args, **kwargs):
        """Initialize form with Crispy Forms helper."""
        super().__init__(*args, **kwargs)
        self.helper = FormHelper()
        self.helper.form_tag = False
        self.helper.layout = Layout(
            Field("confirm"),
        )
