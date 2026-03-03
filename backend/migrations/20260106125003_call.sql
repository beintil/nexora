-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS call (
    id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),                       -- Внутренний уникальный ID звонка в нашей системе
    company_telephony_id UUID NOT NULL REFERENCES company_telephony(id),            -- ID связи компании с телефонией (внешний ключ)
    parent_call_id UUID NULL REFERENCES call(id) ON DELETE CASCADE,
    external_parent_call_id TEXT NULL, -- Уникальный идентификатор родительского звонка из телефонии
    waiting_for_parent BOOLEAN NOT NULL DEFAULT FALSE, -- Ждет ли родительский звонок
    external_call_id TEXT NOT NULL,                -- Уникальный идентификатор звонка от телефонии (например, CallSid у Twilio)
    from_number TEXT NOT NULL, -- Номер кто звонит
    to_number TEXT NOT NULL, -- Номер кто получает звонок
    direction TEXT NOT NULL REFERENCES dictionary (value) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), -- Время создания записи
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()  -- Время последнего обновления записи
);

-- Уникальный индекс на пару (company_telephony_id, external_call_id), чтобы исключить дубли звонков
CREATE UNIQUE INDEX uq_call_company_telephony_external_call_id ON call (company_telephony_id, external_call_id);
-- Индекс для ускорения поиска дочерних звонков по родителю
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS call_parent_call_id_idx
--     ON call (parent_call_id);
--
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS call_company_external_parent_idx
--     ON call (company_telephony_id, external_parent_call_id);


COMMENT ON COLUMN call.parent_call_id IS 'ID родительского звонка (NULL для корневого)';
COMMENT ON COLUMN call.id IS 'Внутренний уникальный идентификатор звонка';
COMMENT ON COLUMN call.company_telephony_id IS 'ID связи компании с телефонией (внешний ключ)';
COMMENT ON COLUMN call.external_call_id IS 'Уникальный ID звонка, полученный от провайдера телефонии (например, Twilio CallSid)';
COMMENT ON COLUMN call.from_number IS 'Номер кто звонит';
COMMENT ON COLUMN call.to_number IS 'Номер кто получает звонок';
COMMENT ON COLUMN call.created_at IS 'Время создания записи о звонке';
COMMENT ON COLUMN call.updated_at IS 'Время последнего обновления записи о звонке';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
