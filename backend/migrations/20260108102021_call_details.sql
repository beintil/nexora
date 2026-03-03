-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS call_detail (
    -- call_id — ссылка на звонок
    -- 1:1 отношение с таблицей call
    call_id UUID PRIMARY KEY REFERENCES call(id) ON DELETE CASCADE,

    -- recording_sid — идентификатор записи разговора у провайдера телефонии
    -- Например: Twilio RecordingSid, Asterisk mixmonitor id
    recording_sid TEXT,

    -- recording_url — ссылка на файл записи разговора
    -- Может появиться с задержкой после завершения звонка
    recording_url TEXT,

    -- recording_duration — длительность записи разговора в секундах
    recording_duration INTEGER,

    -- from_country — страна номера инициатора звонка (ISO 3166-1 alpha-2)
    -- Например: US, TH, RU
    from_country CHAR(2) NOT NULL REFERENCES country(code) ON DELETE RESTRICT,

    -- from_city — город инициатора звонка
    from_city TEXT,

    -- to_country — страна номера получателя звонка (ISO 3166-1 alpha-2)
    to_country CHAR(2) NOT NULL REFERENCES country(code) ON DELETE RESTRICT,

    -- to_city — город получателя звонка
    to_city TEXT,

    -- carrier — оператор связи (carrier), если предоставляется провайдером
    -- Используется для аналитики и диагностики
    carrier TEXT,

    -- trunk — trunk / SIP-линия / DID, через которую прошёл звонок
    -- Полезно для биллинга и маршрутизации
    trunk TEXT,

    -- created_at — время создания записи в таблице call_details
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- updated_at — время последнего обновления дополнительных данных звонка
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Комментарии к таблице
COMMENT ON TABLE call_detail IS
    'Дополнительные атрибуты телефонного звонка: запись разговора, география номеров, оператор и trunk. Не является событиями звонка.';

COMMENT ON COLUMN call_detail.call_id IS
    'Идентификатор звонка. Связь 1:1 с таблицей call.';

COMMENT ON COLUMN call_detail.recording_sid IS
    'Идентификатор записи разговора у провайдера телефонии.';

COMMENT ON COLUMN call_detail.recording_url IS
    'URL файла записи разговора. Может быть недоступен сразу после завершения звонка.';

COMMENT ON COLUMN call_detail.recording_duration IS
    'Длительность записи разговора в секундах.';

COMMENT ON COLUMN call_detail.from_country IS
    'Страна номера инициатора звонка (ISO 3166-1 alpha-2).';

COMMENT ON COLUMN call_detail.from_city IS
    'Город инициатора звонка.';

COMMENT ON COLUMN call_detail.to_country IS
    'Страна номера получателя звонка (ISO 3166-1 alpha-2).';

COMMENT ON COLUMN call_detail.to_city IS
    'Город получателя звонка.';

COMMENT ON COLUMN call_detail.carrier IS
    'Оператор связи (carrier), если предоставляется провайдером телефонии.';

COMMENT ON COLUMN call_detail.trunk IS
    'SIP trunk / линия / DID, через которую прошёл звонок.';

COMMENT ON COLUMN call_detail.created_at IS
    'Время создания записи в таблице call_details.';

COMMENT ON COLUMN call_detail.updated_at IS
    'Время последнего обновления дополнительных данных звонка.';

-- Индексы для аналитики и фильтрации

-- По стране инициатора
CREATE INDEX IF NOT EXISTS idx_call_details_from_country
    ON call_detail (from_country);

-- По стране получателя
CREATE INDEX IF NOT EXISTS idx_call_details_to_country
    ON call_detail (to_country);

-- По оператору связи
CREATE INDEX IF NOT EXISTS idx_call_details_carrier
    ON call_detail (carrier);

-- По trunk / линии
CREATE INDEX IF NOT EXISTS idx_call_details_trunk
    ON call_detail (trunk);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
