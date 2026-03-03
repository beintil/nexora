-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS telephony (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

COMMENT ON COLUMN telephony.id IS 'Уникальный идентификатор телефонии';
COMMENT ON COLUMN telephony.name IS 'Название интегрированной телефонии (уникально в системе)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
