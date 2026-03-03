-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS company
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON COLUMN company.id IS 'Уникальный идентификатор компании';
COMMENT ON COLUMN company.name IS 'Название компании (уникальное)';
COMMENT ON COLUMN company.created_at IS 'Время создания записи';
COMMENT ON COLUMN company.updated_at IS 'Время последнего обновления записи';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
