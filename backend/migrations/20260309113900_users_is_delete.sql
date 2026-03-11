-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_delete BOOLEAN NOT NULL DEFAULT false;
-- +goose Down
