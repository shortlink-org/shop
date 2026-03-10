ALTER TABLE oms.order_delivery_info
    ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(32) NOT NULL DEFAULT 'DELIVERY_STATUS_UNSPECIFIED',
    ADD COLUMN IF NOT EXISTS requested_at TIMESTAMPTZ NULL;

COMMENT ON COLUMN oms.order_delivery_info.delivery_status IS 'Delivery lifecycle status from delivery Kafka events';
COMMENT ON COLUMN oms.order_delivery_info.requested_at IS 'When OMS successfully requested delivery and received package_id';
