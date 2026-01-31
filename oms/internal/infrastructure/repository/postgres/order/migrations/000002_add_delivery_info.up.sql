-- Add delivery info table for orders
CREATE TABLE IF NOT EXISTS oms.order_delivery_info (
    order_id             UUID PRIMARY KEY REFERENCES oms.orders(id) ON DELETE CASCADE,
    -- Pickup address (where the package will be picked up)
    pickup_street        VARCHAR(255),
    pickup_city          VARCHAR(100),
    pickup_postal_code   VARCHAR(20),
    pickup_country       VARCHAR(100),
    pickup_latitude      DECIMAL(10, 7),
    pickup_longitude     DECIMAL(10, 7),
    -- Delivery address (where the package should be delivered)
    delivery_street      VARCHAR(255) NOT NULL,
    delivery_city        VARCHAR(100) NOT NULL,
    delivery_postal_code VARCHAR(20),
    delivery_country     VARCHAR(100) NOT NULL,
    delivery_latitude    DECIMAL(10, 7),
    delivery_longitude   DECIMAL(10, 7),
    -- Delivery period
    period_start         TIMESTAMPTZ NOT NULL,
    period_end           TIMESTAMPTZ NOT NULL,
    -- Package info
    weight_kg            DECIMAL(8, 3),
    dimensions           VARCHAR(50),
    -- Priority
    priority             VARCHAR(20) NOT NULL DEFAULT 'NORMAL',
    -- Package ID from delivery service (set after order is sent to delivery)
    package_id           UUID
);

COMMENT ON TABLE oms.order_delivery_info IS 'Delivery information for orders';
COMMENT ON COLUMN oms.order_delivery_info.pickup_street IS 'Street address for package pickup';
COMMENT ON COLUMN oms.order_delivery_info.delivery_street IS 'Street address for package delivery';
COMMENT ON COLUMN oms.order_delivery_info.period_start IS 'Start of desired delivery time window';
COMMENT ON COLUMN oms.order_delivery_info.period_end IS 'End of desired delivery time window';
COMMENT ON COLUMN oms.order_delivery_info.priority IS 'Delivery priority: NORMAL, URGENT';
COMMENT ON COLUMN oms.order_delivery_info.package_id IS 'Package ID assigned by delivery service';

CREATE INDEX IF NOT EXISTS order_delivery_info_package_id_idx ON oms.order_delivery_info(package_id) WHERE package_id IS NOT NULL;
