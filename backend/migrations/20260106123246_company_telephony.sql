-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS company_telephony (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES company(id),
    telephony_id INT NOT NULL REFERENCES telephony(id),
    external_account_id TEXT, -- ID компании в конкретной телефонии (например, SID аккаунта Twilio)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (company_id, telephony_id),
    UNIQUE (telephony_id, external_account_id)
);

COMMENT ON COLUMN company_telephony.id IS 'Уникальный идентификатор записи связи компании и телефонии';
COMMENT ON COLUMN company_telephony.company_id IS 'ID компании';
COMMENT ON COLUMN company_telephony.telephony_id IS 'ID телефонии';
COMMENT ON COLUMN company_telephony.external_account_id IS 'Идентификатор компании в системе телефонии (например, Twilio Account SID)';
COMMENT ON COLUMN company_telephony.created_at IS 'Время создания записи';
COMMENT ON COLUMN company_telephony.updated_at IS 'Время последнего обновления записи';

CREATE INDEX idx_company_telephony_company ON company_telephony(company_id);
CREATE INDEX idx_company_telephony_telephony ON company_telephony(telephony_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
