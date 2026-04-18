-- 000003_create_outbox_messages_table.up.sql
CREATE TABLE IF NOT EXISTS outbox_messages (
                                               id BIGSERIAL PRIMARY KEY,
                                               event_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
    );

CREATE INDEX idx_outbox_messages_status ON outbox_messages(status);
CREATE INDEX idx_outbox_messages_created_at ON outbox_messages(created_at);
CREATE INDEX idx_outbox_messages_aggregate_id ON outbox_messages(aggregate_id);