# Migration: replace Good.id (BigAuto) with UUID

import uuid as uuid_lib

from django.db import migrations, models
from django.db.models import deletion


def backfill_uuids(apps, schema_editor):
    Good = apps.get_model("goods", "Good")
    GoodImage = apps.get_model("goods", "GoodImage")
    HistoricalGood = apps.get_model("goods", "HistoricalGood")

    mapping = {}  # old_id (int) -> new_uuid
    for good in Good.objects.all():
        new_uuid = uuid_lib.uuid4()
        good.id_new = new_uuid
        good.save(update_fields=["id_new"])
        mapping[good.id] = new_uuid

    for img in GoodImage.objects.select_related("good").all():
        img.good_new_id = mapping[img.good_id]
        img.save(update_fields=["good_new_id"])

    for hist in HistoricalGood.objects.all():
        if hist.id in mapping:
            hist.id_new = mapping[hist.id]
            hist.save(update_fields=["id_new"])


class Migration(migrations.Migration):
    dependencies = [
        ("goods", "0005_add_history"),
    ]

    operations = [
        migrations.AddField(
            model_name="good",
            name="id_new",
            field=models.UUIDField(db_index=True, null=True, unique=True),
        ),
        migrations.AddField(
            model_name="goodimage",
            name="good_new",
            field=models.ForeignKey(
                null=True,
                on_delete=models.deletion.CASCADE,
                related_name="+",
                to="goods.good",
                to_field="id_new",
                db_column="good_new_id",
            ),
        ),
        migrations.AddField(
            model_name="historicalgood",
            name="id_new",
            field=models.UUIDField(db_index=True, null=True),
        ),
        migrations.RunPython(backfill_uuids, migrations.RunPython.noop),
        migrations.RunSQL(
            "ALTER TABLE goods_good ALTER COLUMN id_new SET NOT NULL",
            reverse_sql=migrations.RunSQL.noop,
        ),
        # Drop PK; CASCADE automatically drops dependent FK (goods_goodimage.good_id -> goods_good.id)
        migrations.RunSQL(
            "ALTER TABLE goods_good DROP CONSTRAINT goods_good_pkey CASCADE",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_good DROP COLUMN id",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_good RENAME COLUMN id_new TO id",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_good ADD PRIMARY KEY (id)",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_goodimage DROP COLUMN good_id CASCADE",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_goodimage RENAME COLUMN good_new_id TO good_id",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_goodimage ADD CONSTRAINT goods_goodimage_good_id_fkey "
            "FOREIGN KEY (good_id) REFERENCES goods_good(id) ON DELETE CASCADE",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "CREATE INDEX IF NOT EXISTS goods_goodimage_good_id_idx ON goods_goodimage(good_id)",
            reverse_sql="DROP INDEX IF EXISTS goods_goodimage_good_id_idx",
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_historicalgood DROP COLUMN id",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "ALTER TABLE goods_historicalgood RENAME COLUMN id_new TO id",
            reverse_sql=migrations.RunSQL.noop,
        ),
        migrations.RunSQL(
            "CREATE INDEX IF NOT EXISTS goods_historicalgood_id_idx ON goods_historicalgood(id)",
            reverse_sql="DROP INDEX IF EXISTS goods_historicalgood_id_idx",
        ),
        migrations.SeparateDatabaseAndState(
            state_operations=[
                migrations.RemoveField(model_name="good", name="id_new"),
                migrations.AlterField(
                    model_name="good",
                    name="id",
                    field=models.UUIDField(
                        default=uuid_lib.uuid4,
                        editable=False,
                        primary_key=True,
                        serialize=False,
                        verbose_name="ID",
                    ),
                ),
                migrations.RemoveField(model_name="goodimage", name="good_new"),
                migrations.AlterField(
                    model_name="historicalgood",
                    name="id",
                    field=models.UUIDField(
                        blank=True, db_index=True, editable=False, verbose_name="ID"
                    ),
                ),
                migrations.RemoveField(model_name="historicalgood", name="id_new"),
            ],
        ),
    ]
