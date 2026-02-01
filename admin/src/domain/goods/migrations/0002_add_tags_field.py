# Generated manually for adding tags ArrayField

import django.contrib.postgres.fields
from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ("goods", "0001_initial"),
    ]

    operations = [
        migrations.AddField(
            model_name="good",
            name="tags",
            field=django.contrib.postgres.fields.ArrayField(
                base_field=models.CharField(max_length=50),
                blank=True,
                default=list,
                help_text="Product tags for categorization and filtering",
                size=None,
                verbose_name="Tags",
            ),
        ),
    ]
