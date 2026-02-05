ALTER TABLE oms.order_delivery_info
    DROP COLUMN IF EXISTS recipient_name,
    DROP COLUMN IF EXISTS recipient_phone,
    DROP COLUMN IF EXISTS recipient_email;
