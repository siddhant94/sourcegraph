BEGIN;

CREATE TABLE IF NOT EXISTS webhook_logs (
    id BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    external_service_id INTEGER NULL REFERENCES external_services (id),
    request BYTEA NOT NULL,
    encryption_key_id TEXT NOT NULL,
    error TEXT NULL
);

CREATE INDEX IF NOT EXISTS
    webhook_logs_received_at_idx
ON
    webhook_logs (received_at);

CREATE INDEX IF NOT EXISTS
    webhook_logs_external_service_id_idx
ON
    webhook_logs (external_service_id);

CREATE INDEX IF NOT EXISTS
    webhook_logs_error_idx
ON
    webhook_logs (error)
WHERE
    error IS NOT NULL;

COMMIT;
