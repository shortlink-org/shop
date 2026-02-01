# Generated manually for adding GoodImage model

from django.db import migrations, models
import django.db.models.deletion


class Migration(migrations.Migration):
    dependencies = [
        ("goods", "0002_add_tags_field"),
    ]

    operations = [
        migrations.AlterModelOptions(
            name="good",
            options={"verbose_name": "Good", "verbose_name_plural": "Goods"},
        ),
        migrations.CreateModel(
            name="GoodImage",
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
                    "image_url",
                    models.URLField(
                        help_text="URL to the product image",
                        verbose_name="Image URL",
                    ),
                ),
                (
                    "alt_text",
                    models.CharField(
                        blank=True,
                        help_text="Alternative text for accessibility",
                        max_length=255,
                        verbose_name="Alt Text",
                    ),
                ),
                (
                    "is_primary",
                    models.BooleanField(
                        default=False,
                        help_text="Mark as the main product image",
                        verbose_name="Primary Image",
                    ),
                ),
                (
                    "sort_order",
                    models.PositiveIntegerField(
                        default=0,
                        help_text="Order in which images are displayed",
                        verbose_name="Sort Order",
                    ),
                ),
                (
                    "created_at",
                    models.DateTimeField(auto_now_add=True),
                ),
                (
                    "good",
                    models.ForeignKey(
                        on_delete=django.db.models.deletion.CASCADE,
                        related_name="images",
                        to="goods.good",
                        verbose_name="Good",
                    ),
                ),
            ],
            options={
                "verbose_name": "Good Image",
                "verbose_name_plural": "Good Images",
                "ordering": ["sort_order", "created_at"],
            },
        ),
    ]
