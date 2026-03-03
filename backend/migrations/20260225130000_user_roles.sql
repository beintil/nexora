-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS role
(
    id   SMALLINT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

COMMENT ON TABLE role IS 'Роли пользователей: Admin/Support — админка, Owner — при регистрации (создание компании), Manager — создаётся Owner-ом';
COMMENT ON COLUMN role.id IS 'Идентификатор роли: 0=Admin, 1=Support, 2=Owner, 3=Manager';
COMMENT ON COLUMN role.name IS 'Название роли';

INSERT INTO role (id, name) VALUES
    (0, 'Admin'),
    (1, 'Support'),
    (2, 'Owner'),
    (3, 'Manager')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id SMALLINT NOT NULL DEFAULT 2 REFERENCES role(id);
COMMENT ON COLUMN users.role_id IS 'Роль пользователя (2=Owner по умолчанию при регистрации)';
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
DROP TABLE IF EXISTS role;
-- +goose StatementEnd
