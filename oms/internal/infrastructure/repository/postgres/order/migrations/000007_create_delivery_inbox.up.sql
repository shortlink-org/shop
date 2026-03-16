CREATE TABLE IF NOT EXISTS oms.delivery_inbox_messages (
    consumer_name VARCHAR(100) NOT NULL,
    message_id    VARCHAR(255) NOT NULL,
    topic         VARCHAR(255) NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (consumer_name, message_id)
);

COMMENT ON TABLE oms.delivery_inbox_messages IS 'Inbox deduplication for inbound delivery Kafka messages';
COMMENT ON COLUMN oms.delivery_inbox_messages.consumer_name IS 'Logical consumer name that processed the message';
COMMENT ON COLUMN oms.delivery_inbox_messages.message_id IS 'Inbound broker message identifier used for deduplication';
COMMENT ON COLUMN oms.delivery_inbox_messages.topic IS 'Kafka topic where the message was received';
