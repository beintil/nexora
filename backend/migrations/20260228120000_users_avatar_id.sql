-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_id TEXT;
COMMENT ON COLUMN users.avatar_id IS 'Идентификатор файла аватара в хранилище';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS avatar_id;
-- +goose StatementEnd
