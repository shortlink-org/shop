ALTER TABLE oms.order_delivery_info
    DROP COLUMN IF EXISTS requested_at,
    DROP COLUMN IF EXISTS delivery_status;
