-- Transactional outbox for domain events (watermill-sql v4 DefaultPostgreSQLSchema).
-- Topic: oms_outbox. Table name: watermill_oms_outbox.
-- Used by go-sdk/cqrs WithTxAwareOutbox and outbox forwarder.
CREATE TABLE IF NOT EXISTS "watermill_oms_outbox" (
    "offset"         BIGSERIAL,
    "uuid"           VARCHAR(36) NOT NULL,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "payload"        JSON DEFAULT NULL,
    "metadata"       JSON DEFAULT NULL,
    "transaction_id" xid8 NOT NULL,
    PRIMARY KEY ("transaction_id", "offset")
);

COMMENT ON TABLE "watermill_oms_outbox" IS 'Outbox for OMS domain events; forwarded to Kafka by RunForwarder';
