"""Define the Good model."""

from django.contrib.postgres.fields import ArrayField
from django.db import models
from django.utils.translation import gettext_lazy as _
from djmoney.models.fields import MoneyField
from django_prometheus.models import ExportModelOperationsMixin
from simple_history.models import HistoricalRecords


class Good(ExportModelOperationsMixin("goods"), models.Model):
    """Define the Good model."""

    name = models.CharField(max_length=255)
    price = MoneyField(
        max_digits=10,
        decimal_places=2,
        default_currency="USD",
        verbose_name=_("Price"),
    )
    description = models.TextField()
    tags = ArrayField(
        models.CharField(max_length=50),
        default=list,
        blank=True,
        verbose_name="Tags",
        help_text="Product tags for categorization and filtering",
    )
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    # History tracking
    history = HistoricalRecords()

    class Meta:
        verbose_name = _("Good")
        verbose_name_plural = _("Goods")

    def __str__(self):
        """Return the name of the good."""
        return self.name


class GoodImage(models.Model):
    """Image associated with a Good product."""

    good = models.ForeignKey(
        Good,
        on_delete=models.CASCADE,
        related_name="images",
        verbose_name=_("Good"),
    )
    image_url = models.URLField(
        verbose_name=_("Image URL"),
        help_text=_("URL to the product image"),
    )
    alt_text = models.CharField(
        max_length=255,
        blank=True,
        verbose_name=_("Alt Text"),
        help_text=_("Alternative text for accessibility"),
    )
    is_primary = models.BooleanField(
        default=False,
        verbose_name=_("Primary Image"),
        help_text=_("Mark as the main product image"),
    )
    sort_order = models.PositiveIntegerField(
        default=0,
        verbose_name=_("Sort Order"),
        help_text=_("Order in which images are displayed"),
    )
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        verbose_name = _("Good Image")
        verbose_name_plural = _("Good Images")
        ordering = ["sort_order", "created_at"]

    def __str__(self):
        """Return a description of the image."""
        return f"{self.good.name} - Image {self.sort_order}"
