-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dictionary (
    main_type TEXT NOT NULL,
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL UNIQUE,
    comment TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS dictionary_key_uindex ON dictionary (main_type, key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
