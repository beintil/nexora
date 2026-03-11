-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS currency (
    id SERIAL PRIMARY KEY,
    -- code — ISO 4217 код валюты (например, RUB, USD, EUR)
    code CHAR(3) NOT NULL UNIQUE,
    -- name — название валюты на языке системы (Russian за основу)
    name VARCHAR(100) NOT NULL,
    -- symbol — графический символ валюты (например, ₽, $)
    symbol VARCHAR(10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Комментарии
COMMENT ON TABLE currency IS 'Справочник валют (ISO 4217), используемый в биллинге';
COMMENT ON COLUMN currency.code IS 'ISO 4217 код валюты';
COMMENT ON COLUMN currency.name IS 'Название валюты';
COMMENT ON COLUMN currency.symbol IS 'Символ валюты';

-- Индексы
CREATE UNIQUE INDEX IF NOT EXISTS idx_currency_code ON currency(code);

-- Начальные данные
INSERT INTO currency (code, name, symbol) VALUES
    ('RUB', 'Российский рубль', '₽')
ON CONFLICT (code) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS currency;
-- +goose StatementEnd
