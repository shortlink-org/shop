"""Define the admin view for the Good model."""

from decimal import Decimal

from django.contrib import admin, messages
from django.contrib.postgres.fields import ArrayField
from django.db import models
from django.db.models import TextChoices
from django.http import HttpRequest, HttpResponseRedirect
from django.shortcuts import redirect, render
from django.urls import path, reverse
from django.utils.translation import gettext_lazy as _
from djmoney.money import Money
from import_export import resources
from import_export.admin import ImportExportModelAdmin
from mimesis import Generic
from mimesis import Locale
from simple_history.admin import SimpleHistoryAdmin
from unfold.admin import ModelAdmin, TabularInline
from unfold.contrib.filters.admin import RangeDateTimeFilter
from unfold.contrib.forms.widgets import ArrayWidget, WysiwygWidget
from unfold.contrib.import_export.forms import ExportForm, ImportForm
from unfold.decorators import action, display
from unfold.paginator import InfinitePaginator

from .models import Good, GoodImage


class GoodResource(resources.ModelResource):
    """Resource class for import/export of Good model."""

    class Meta:
        model = Good
        fields = ("id", "name", "price", "price_currency", "description", "tags", "created_at", "updated_at")
        export_order = ("id", "name", "price", "price_currency", "description", "tags", "created_at")


class GoodImageInline(TabularInline):
    """Inline admin for product images."""

    model = GoodImage
    extra = 1
    fields = ("image_url", "alt_text", "is_primary", "sort_order")
    ordering = ("sort_order",)


class TagChoices(TextChoices):
    """Predefined tags for goods categorization."""

    ORGANIC = "organic", _("Organic")
    VEGAN = "vegan", _("Vegan")
    GLUTEN_FREE = "gluten-free", _("Gluten Free")
    LOW_FAT = "low-fat", _("Low Fat")
    SUGAR_FREE = "sugar-free", _("Sugar Free")
    PREMIUM = "premium", _("Premium")
    SALE = "sale", _("On Sale")
    NEW = "new", _("New Arrival")
    BESTSELLER = "bestseller", _("Bestseller")
    LOCAL = "local", _("Local Product")


class GoodAdmin(SimpleHistoryAdmin, ModelAdmin, ImportExportModelAdmin):
    """Define the admin view for the Good model."""

    # Import/Export configuration
    resource_class = GoodResource
    import_form_class = ImportForm
    export_form_class = ExportForm

    list_display = ("display_good_header", "display_tags", "image_count", "created_at")
    search_fields = ("name", "price", "tags")
    inlines = [GoodImageInline]

    @display(description=_("Images"))
    def image_count(self, obj):
        """Display number of images."""
        count = obj.images.count()
        return count if count > 0 else "-"
    paginator = InfinitePaginator
    show_full_result_count = False
    list_filter = (
        ("created_at", RangeDateTimeFilter),
        ("updated_at", RangeDateTimeFilter),
    )
    list_filter_submit = True
    ordering = ("created_at",)
    change_list_template = "admin/goods/good/change_list.html"
    actions_row = ["duplicate_good"]

    # Custom widgets for fields
    formfield_overrides = {
        ArrayField: {
            "widget": ArrayWidget,
        },
        models.TextField: {
            "widget": WysiwygWidget,
        },
    }

    def get_form(self, request, obj=None, change=False, **kwargs):
        """Override form to configure ArrayWidget with choices."""
        form = super().get_form(request, obj, change, **kwargs)
        if "tags" in form.base_fields:
            form.base_fields["tags"].widget = ArrayWidget(choices=TagChoices)
        return form

    @display(header=True)
    def display_good_header(self, obj):
        """Display good name with price as two-line header."""
        # Get first letter of first two words for initials
        words = obj.name.split()
        initials = "".join(w[0].upper() for w in words[:2]) if words else "G"
        return [
            obj.name,
            str(obj.price),
            initials,
        ]

    @display(
        description=_("Tags"),
        label={
            "organic": "success",
            "vegan": "success",
            "gluten-free": "info",
            "low-fat": "info",
            "sugar-free": "info",
            "premium": "warning",
            "sale": "danger",
            "new": "success",
            "bestseller": "warning",
            "local": "info",
        },
    )
    def display_tags(self, obj):
        """Display tags as colored labels."""
        if obj.tags:
            # Return list of (value, label) tuples for multiple labels
            return [(tag, tag.replace("-", " ").title()) for tag in obj.tags]
        return "-"

    @action(
        description=_("Duplicate"),
        permissions=["duplicate"],
        url_path="duplicate",
    )
    def duplicate_good(self, request: HttpRequest, object_id: int):
        """Create a copy of the selected good."""
        original = Good.objects.get(pk=object_id)
        Good.objects.create(
            name=f"{original.name} (Copy)",
            price=original.price,
            description=original.description,
            tags=original.tags.copy() if original.tags else [],
        )
        messages.success(request, f"Good '{original.name}' duplicated successfully.")
        return redirect(reverse("admin:goods_good_changelist"))

    def has_duplicate_permission(self, request: HttpRequest) -> bool:
        """Check if user can duplicate goods."""
        return request.user.has_perm("goods.add_good")

    def get_urls(self):
        """Add custom URLs for generating goods."""
        urls = super().get_urls()
        custom_urls = [
            path(
                "generate/",
                self.admin_site.admin_view(self.generate_goods_view),
                name="goods_good_generate",
            ),
        ]
        return custom_urls + urls

    def generate_goods_view(self, request):
        """View for generating fake goods."""
        if request.method == "POST":
            raw_count = request.POST.get("count", 10)
            try:
                count = int(raw_count)
            except (TypeError, ValueError):
                count = 10
            count = max(1, min(count, 1000))
            locale_str = request.POST.get("locale", "en")

            # Map locale string to Mimesis Locale
            locale_map = {
                "en": Locale.EN,
                "ru": Locale.RU,
                "de": Locale.DE,
                "fr": Locale.FR,
                "es": Locale.ES,
                "it": Locale.IT,
                "ja": Locale.JA,
                "zh": Locale.ZH,
            }
            locale = locale_map.get(locale_str, Locale.EN)

            # Generate goods using Mimesis
            gen = Generic(locale=locale)

            # Product templates for more realistic names
            product_templates = [
                lambda: f"{gen.food.fruit()} {gen.choice(['Jam', 'Juice', 'Smoothie', 'Preserve', 'Syrup'])}",
                lambda: f"{gen.choice(['Organic', 'Premium', 'Fresh', 'Artisan', 'Natural'])} {gen.food.vegetable()}",
                lambda: f"{gen.food.dish()} {gen.choice(['Mix', 'Kit', 'Set', 'Pack', 'Box'])}",
                lambda: f"{gen.choice(['Gourmet', 'Classic', 'Traditional', 'Homestyle', "Chef's"])} {gen.food.dish()}",
                lambda: f"{gen.food.spices()} {gen.choice(['Blend', 'Seasoning', 'Mix', 'Powder'])}",
                lambda: f"{gen.choice(['Sparkling', 'Craft', 'Imported', 'Local', 'Organic'])} {gen.food.drink()}",
                lambda: f"{gen.food.fruit()} & {gen.food.fruit()} {gen.choice(['Salad', 'Mix', 'Combo', 'Duo'])}",
            ]

            # Description templates (HTML formatted for WYSIWYG)
            def generate_description():
                intro = gen.choice([
                    f"<p><strong>Premium Quality Product</strong></p>",
                    f"<p><strong>Artisan Selection</strong></p>",
                    f"<p><strong>Fresh & Natural</strong></p>",
                ])
                features = [
                    f"Made with {gen.choice(['100%', 'premium', 'finest', 'selected', 'organic'])} ingredients.",
                    f"Perfect for {gen.choice(['breakfast', 'lunch', 'dinner', 'snacking', 'parties', 'everyday meals'])}.",
                    f"{gen.choice(['Rich in', 'Contains', 'Source of', 'Packed with'])} {gen.choice(['vitamins', 'nutrients', 'antioxidants', 'fiber', 'protein'])}.",
                    f"Best served {gen.choice(['chilled', 'warm', 'at room temperature', 'fresh', 'immediately'])}.",
                ]
                selected_features = gen.random.choices(features, k=3)
                features_html = "<ul>" + "".join(f"<li>{f}</li>" for f in selected_features) + "</ul>"
                footer = f"<p><em>Net weight: {gen.numeric.integer_number(start=100, end=1000)}g</em></p>"
                return intro + features_html + footer

            # Generate random tags
            all_tags = [choice.value for choice in TagChoices]

            def generate_tags():
                num_tags = gen.random.randint(0, 3)
                return gen.random.sample(all_tags, k=min(num_tags, len(all_tags)))

            goods = []
            for _ in range(count):
                template = gen.choice(product_templates)
                price_value = Decimal(str(gen.finance.price(minimum=1.99, maximum=99.99)))
                goods.append(
                    Good(
                        name=template(),
                        price=Money(price_value, "USD"),
                        description=generate_description(),
                        tags=generate_tags(),
                    )
                )

            Good.objects.bulk_create(goods)
            messages.success(request, f"Successfully generated {count} goods.")
            return HttpResponseRedirect("../")

        context = {
            **self.admin_site.each_context(request),
            "title": "Generate Goods",
            "opts": self.model._meta,
        }
        return render(request, "admin/goods/good/generate_form.html", context)


admin.site.register(Good, GoodAdmin)
