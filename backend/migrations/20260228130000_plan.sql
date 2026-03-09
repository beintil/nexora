-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plan
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name         TEXT        NOT NULL,
    slug         TEXT        NOT NULL UNIQUE,
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order   INTEGER     NOT NULL DEFAULT 0,
    description  TEXT
);

COMMENT ON TABLE plan IS 'Тарифные планы компании';
COMMENT ON COLUMN plan.id IS 'Идентификатор плана';
COMMENT ON COLUMN plan.name IS 'Отображаемое название плана';
COMMENT ON COLUMN plan.slug IS 'Машинное имя плана (уникальное)';
COMMENT ON COLUMN plan.is_active IS 'Флаг активности плана';
COMMENT ON COLUMN plan.sort_order IS 'Порядок сортировки плана';

CREATE TABLE IF NOT EXISTS plan_limit
(
    plan_id    UUID NOT NULL REFERENCES plan (id) ON DELETE CASCADE,
    limit_type TEXT NOT NULL REFERENCES dictionary (value) ON DELETE RESTRICT,
    value      BIGINT NOT NULL,
    extra      JSONB,
    PRIMARY KEY (plan_id, limit_type)
);

COMMENT ON TABLE plan_limit IS 'Лимиты тарифных планов по типам';
COMMENT ON COLUMN plan_limit.plan_id IS 'Идентификатор плана';
COMMENT ON COLUMN plan_limit.limit_type IS 'Тип лимита (ключ справочника plan_limit_type)';
COMMENT ON COLUMN plan_limit.value IS 'Значение лимита (интерпретация зависит от типа лимита)';
COMMENT ON COLUMN plan_limit.extra IS 'Дополнительные параметры лимита, конфиг';

CREATE INDEX IF NOT EXISTS idx_plan_limit_limit_type ON plan_limit (limit_type);

CREATE TABLE IF NOT EXISTS company_plan
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID        NOT NULL REFERENCES company (id) ON DELETE CASCADE,
    plan_id    UUID        NOT NULL REFERENCES plan (id),
    is_active  BOOL NOT NULL DEFAULT FALSE,
    started_at TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    CONSTRAINT company_plan_one_interval CHECK (started_at < ends_at)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_company_plan_company_id_plan_id ON company_plan (company_id, plan_id, is_active);

COMMENT ON TABLE company_plan IS 'Привязка компаний к тарифным планам';
COMMENT ON COLUMN company_plan.id IS 'Идентификатор привязки';
COMMENT ON COLUMN company_plan.company_id IS 'Идентификатор компании';
COMMENT ON COLUMN company_plan.plan_id IS 'Идентификатор тарифа';
COMMENT ON COLUMN company_plan.is_active IS 'Флаг активности плана';
COMMENT ON COLUMN company_plan.started_at IS 'Начало действия плана для компании';
COMMENT ON COLUMN company_plan.ends_at IS 'Окончание действия плана для компании';

CREATE TABLE IF NOT EXISTS plan_usage
(
    company_plan_id UUID NOT NULL REFERENCES company_plan (id) ON DELETE CASCADE,
    limit_type TEXT NOT NULL REFERENCES dictionary (value) ON DELETE RESTRICT,
    value        BIGINT NOT NULL,
    PRIMARY KEY (company_plan_id, limit_type)
);

COMMENT ON TABLE plan_usage IS 'Учёт использования лимитов по периодам (например, в месяц) для компании';
COMMENT ON COLUMN plan_usage.company_plan_id IS 'Идентификатор привязки компании к плану';
COMMENT ON COLUMN plan_usage.limit_type IS 'Тип лимита (ключ справочника plan_limit_type)';
COMMENT ON COLUMN plan_usage.value IS 'Накопленное значение использования за период';


CREATE TABLE IF NOT EXISTS plan_usage_over
(
    company_id   UUID NOT NULL REFERENCES company (id) ON DELETE CASCADE,
    limit_type TEXT NOT NULL REFERENCES dictionary (value) ON DELETE RESTRICT,
    value        BIGINT NOT NULL,
    is_active BOOL NOT NULL DEFAULT FALSE,
    period_start DATE NOT NULL,
    period_end DATE DEFAULT NULL,
    PRIMARY KEY (company_id, limit_type)
);
COMMENT ON TABLE plan_usage_over IS 'Учёт использования лимитов по периодам (например, в месяц) для компаниит в сверх лимитах основного плана';
COMMENT ON COLUMN plan_usage_over.company_id IS 'Идентификатор компании';
COMMENT ON COLUMN plan_usage_over.limit_type IS 'Тип лимита (ключ справочника plan_limit_type)';
COMMENT ON COLUMN plan_usage_over.value IS 'Накопленное значение использования за период';
COMMENT ON COLUMN plan_usage_over.period_start IS 'Начало периода';
COMMENT ON COLUMN plan_usage_over.period_end IS 'Конец периода';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

