-- +goose Up
-- +goose StatementBegin
ALTER TABLE company DROP CONSTRAINT IF EXISTS company_name_key;
COMMENT ON COLUMN company.name IS 'Название компании';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE company ADD CONSTRAINT company_name_key UNIQUE (name);
COMMENT ON COLUMN company.name IS 'Название компании (уникальное)';
-- +goose StatementEnd
