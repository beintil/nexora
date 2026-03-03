-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS call_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID NOT NULL REFERENCES call(id) ON DELETE CASCADE,
    status TEXT NOT NULL REFERENCES dictionary(value) ON DELETE RESTRICT,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_call_events_call_id_timestamp
    ON call_events (call_id, timestamp);

 CREATE INDEX IF NOT EXISTS idx_call_events_status_timestamp
    ON call_events (status, timestamp);

CREATE INDEX IF NOT EXISTS idx_call_events_call_id_timestamp_desc
    ON call_events (call_id, timestamp DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
