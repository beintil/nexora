-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    company_id    UUID         NOT NULL REFERENCES company (id) ON DELETE CASCADE,
    email        TEXT UNIQUE,
    phone        TEXT UNIQUE,
    password_hash TEXT        NOT NULL,
    full_name    TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT users_email_or_phone_required CHECK (
        (email IS NOT NULL AND trim(email) != '') OR
        (phone IS NOT NULL AND trim(phone) != '')
    )
);

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users (company_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL;

COMMENT ON TABLE users IS 'Пользователи системы; обязательно email или телефон';
COMMENT ON COLUMN users.company_id IS 'Идентификатор компании';
COMMENT ON COLUMN users.email IS 'Email (опционально, если указан телефон)';
COMMENT ON COLUMN users.phone IS 'Телефон в формате E.164 (опционально, если указан email)';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля (bcrypt); пароль обязателен при регистрации';
COMMENT ON COLUMN users.full_name IS 'ФИО сотрудника (опционально)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
