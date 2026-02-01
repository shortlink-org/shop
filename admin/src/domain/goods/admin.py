"""Define the admin view for the Good model."""

from decimal import Decimal

from django.contrib import admin, messages
from django.http import HttpResponseRedirect
from django.shortcuts import render
from django.urls import path
from mimesis import Generic
from mimesis import Locale

from .models import Good


class GoodAdmin(admin.ModelAdmin):
    """Define the admin view for the Good model."""

    list_display = ("name", "price", "created_at")
    search_fields = ("name", "price")
    list_filter = ("created_at", "updated_at")
    ordering = ("created_at",)
    change_list_template = "admin/goods/good/change_list.html"

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
            count = int(request.POST.get("count", 10))
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

            # Description templates
            def generate_description():
                features = [
                    f"Made with {gen.choice(['100%', 'premium', 'finest', 'selected', 'organic'])} ingredients.",
                    f"Perfect for {gen.choice(['breakfast', 'lunch', 'dinner', 'snacking', 'parties', 'everyday meals'])}.",
                    f"{gen.choice(['Rich in', 'Contains', 'Source of', 'Packed with'])} {gen.choice(['vitamins', 'nutrients', 'antioxidants', 'fiber', 'protein'])}.",
                    f"Best served {gen.choice(['chilled', 'warm', 'at room temperature', 'fresh', 'immediately'])}.",
                    f"{gen.choice(['Family recipe', 'Handcrafted', 'Locally sourced', 'Sustainably produced', 'Award-winning'])} quality.",
                    f"Net weight: {gen.numeric.integer_number(start=100, end=1000)}g.",
                ]
                return " ".join(gen.random.choices(features, k=3))

            goods = []
            for _ in range(count):
                template = gen.choice(product_templates)
                goods.append(
                    Good(
                        name=template(),
                        price=Decimal(str(gen.finance.price(minimum=1.99, maximum=99.99))),
                        description=generate_description(),
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
