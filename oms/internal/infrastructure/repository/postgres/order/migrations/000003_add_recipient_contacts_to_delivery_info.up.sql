-- Add recipient contact columns to order_delivery_info
ALTER TABLE oms.order_delivery_info
    ADD COLUMN IF NOT EXISTS recipient_name  VARCHAR(255),
    ADD COLUMN IF NOT EXISTS recipient_phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS recipient_email VARCHAR(255);

COMMENT ON COLUMN oms.order_delivery_info.recipient_name IS 'Name of the person receiving the delivery';
COMMENT ON COLUMN oms.order_delivery_info.recipient_phone IS 'Phone number for delivery contact';
COMMENT ON COLUMN oms.order_delivery_info.recipient_email IS 'Email for delivery notifications';
