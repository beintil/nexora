-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_registration BOOLEAN NOT NULL DEFAULT false;
COMMENT ON COLUMN users.verified_registration IS 'Подтвердил ли пользователь email (переход по ссылке из письма)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS verified_registration;
-- +goose StatementEnd
