# Generated manually for converting price to MoneyField

import djmoney.models.fields
from decimal import Decimal
from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ("goods", "0003_add_goodimage"),
    ]

    operations = [
        # Remove the old price field
        migrations.RemoveField(
            model_name="good",
            name="price",
        ),
        # Add the new MoneyField (creates price and price_currency fields)
        migrations.AddField(
            model_name="good",
            name="price",
            field=djmoney.models.fields.MoneyField(
                decimal_places=2,
                default=Decimal("0"),
                default_currency="USD",
                max_digits=10,
                verbose_name="Price",
            ),
        ),
        migrations.AddField(
            model_name="good",
            name="price_currency",
            field=djmoney.models.fields.CurrencyField(
                choices=[
                    ("USD", "US Dollar"),
                    ("EUR", "Euro"),
                    ("GBP", "British Pound"),
                    ("RUB", "Russian Ruble"),
                ],
                default="USD",
                editable=False,
                max_length=3,
            ),
        ),
    ]
