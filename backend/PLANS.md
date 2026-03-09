## Тарифные планы и лимиты в backend

### 1. Модель данных

- **Таблица `plan`** — справочник тарифных планов:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `name TEXT NOT NULL` — отображаемое имя плана
  - `slug TEXT NOT NULL UNIQUE` — машинный идентификатор (используется в конфиге/админке)
  - `is_active BOOLEAN NOT NULL DEFAULT TRUE` — флаг активности
  - `sort_order INTEGER NOT NULL DEFAULT 0` — порядок сортировки
  - `description TEXT` — описание для UI/документации
  - `visible_limits JSONB` — произвольное представление лимитов для фронта (кэшированное описание)

- **Таблица `plan_limit`** — конкретные лимиты плана:
  - `plan_id UUID NOT NULL REFERENCES plan(id) ON DELETE CASCADE`
  - `limit_type TEXT NOT NULL` — строковый ключ лимита
  - `value BIGINT NOT NULL` — числовое значение (счётчик, уровень и т.п.)
  - `extra JSONB` — доп. параметры для сложных лимитов
  - `PRIMARY KEY (plan_id, limit_type)`

- **Таблица `company_plan`** — назначенный план компании:
  - `company_id UUID NOT NULL REFERENCES company(id) ON DELETE CASCADE`
  - `plan_id UUID NOT NULL REFERENCES plan(id)`
  - `started_at TIMESTAMPTZ NOT NULL`
  - `ends_at TIMESTAMPTZ NULL`
  - `UNIQUE (company_id) WHERE ends_at IS NULL` — не более одного активного плана на компанию

- **Таблица `plan_usage`** — учёт использования лимитов по периодам:
  - `company_id UUID NOT NULL REFERENCES company(id) ON DELETE CASCADE`
  - `period_start DATE NOT NULL` — первый день периода (месяц)
  - `limit_type TEXT NOT NULL`
  - `value BIGINT NOT NULL` — накопленное значение за период
  - `PRIMARY KEY (company_id, period_start, limit_type)`

Доменные структуры описаны в `internal/domain/plan.go`:

- `Plan`, `PlanLimit`, `CompanyPlan`, `PlanUsage`
- `PlanLimitKey` — тип для ключей лимитов, с набором констант:
  - `PlanLimitSmsPerMonth`, `PlanLimitTelegramPerMonth`, `PlanLimitEmailPerMonth`
  - `PlanLimitPhonesRUCount`, `PlanLimitPhonesForeignCount`
  - `PlanLimitCallsPerMonth` — лимит на количество создаваемых звонков в месяц (ingestion из вебхуков)

Хранение в БД происходит как `TEXT`, в коде — как `PlanLimitKey` с приведением типов.

---

### 2. Сервис планов

Модуль: `internal/modules/plan`.

- **Интерфейс `Service` (`interface.go`)**:
  - `GetActivePlanByCompanyID(ctx, companyID)` → `(*CompanyPlan, *Plan, []*PlanLimit, srverr.ServerError)`
    - Находит текущий активный план компании по `company_plan`, загружает сам план и его лимиты.
    - Если у компании нет активного плана — ошибка типа `ServiceErrorPlanNotAssigned`.
  - `CheckLimit(ctx, companyID, key, currentUsage, requiredAmount)` → `(ok bool, err srverr.ServerError)`
    - Достаёт активный план, его лимиты и проверяет, вписывается ли `currentUsage + requiredAmount` в лимит `key`.
    - Если для данного ключа нет лимита в `plan_limit`, по умолчанию считается, что ограничение отсутствует (`ok = true`, `err = nil`).
    - При превышении лимита возвращает `ok = false` и ошибку типа `ServiceErrorLimitExceeded`.
    - При отсутствии плана — `ServiceErrorPlanNotAssigned`.
  - `IncrementUsageWithTx(ctx, tx, companyID, key, delta)` → `srverr.ServerError`
    - Работает внутри уже открытой транзакции.
    - Нормализует период к первому дню текущего месяца (`period_start`).
    - Делает upsert в `plan_usage`: создаёт запись, если её нет, или увеличивает `value` на `delta`.

- **Repository (`repository.go`)**:
  - `getActiveCompanyPlan` — выбор активного `company_plan` с учётом `started_at`/`ends_at`.
  - `getPlanByID` — выбор плана.
  - `getPlanLimitsByPlanID` — все лимиты плана.
  - `getUsageByPeriod` — usage по `(company_id, period_start, limit_type)`.
  - `saveUpdateUsageBulk` — batch upsert `plan_usage` (через `pgx.Batch`).

Ошибки репозитория — обычные `error` с префиксом `errRepo...`, сервис маппит их в `srverr`.

---

### 3. Где и как сейчас используются планы

#### 3.1. Ограничения на создание звонков (ingestion)

Создание/обновление звонков происходит через пайплайн вебхуков:

- Модуль: `internal/modules/telephony_ingestion_pipeline`
- Метод: `service.CallWorker(ctx, call *domain.CallWorker, telephony TelephonyName)`

Порядок работы:

1. Открывается транзакция через `postgres.Transaction`.
2. По `call.TelephonyAccountID` и `telephony` находится `CompanyTelephony` и, соответственно, `company_id`.
3. **Перед сохранением звонка** вызывается проверка лимита:

   ```go
   ok, sErr := s.planService.CheckLimit(ctx, companyTelephony.CompanyID, domain.PlanLimitCallsPerMonth, 0, 1)
   ```

   - Используется ключ `PlanLimitCallsPerMonth`.
   - `currentUsage` сейчас передаётся как `0` — фактическое накопленное использование берётся из `plan_limit`/`plan_usage` внутри сервиса.
   - `requiredAmount = 1` — каждый новый звонок увеличивает счётчик на единицу.

4. Если:
   - `sErr != nil` — ошибка маппится наружу как `srverr`.
   - `ok == false` — возвращается `srverr` с типом `ServiceErrorLimitExceeded` из модуля планов (создание звонка блокируется).
5. Если лимит не превышен — выполняется нормализация стран (`countries.Service`) и сохранение звонка/события через `call.Service.SaveUpdateCallWithTX`.

Таким образом, **лимит применяется на стороне ingestion** (входящие вебхуки телефонии), а не в HTTP‑хендлерах.

#### 3.2. Метрики

- Метрики звонков (`GetCompanyCallMetrics` в `internal/modules/call/service.go`) **не ограничиваются планами**:
  - Проверки через `planService.CheckLimit` удалены по требованию: доступ к метрикам должен быть всегда, независимо от тарифа.
  - Метод по‑прежнему валидирует вход (companyID, период), открывает транзакцию и строит агрегаты, но без тарифных ограничений.

#### 3.3. Сообщения (SMS, мессенджеры, email, телефоны)

На текущий момент:

- Конфиг и заготовки для SMS/email уже есть (`internal/config/sms.go`, клиент `pkg/client/email_sender`).
- Реальной реализации отправки SMS/мессенджеров ещё нет, поэтому лимиты по:
  - `PlanLimitSmsPerMonth`
  - `PlanLimitTelegramPerMonth`
  - `PlanLimitEmailPerMonth`
  - `PlanLimitPhonesRUCount`
  - `PlanLimitPhonesForeignCount`

пока **не используются** в коде.

Планируемое использование (как только появятся соответствующие модули):

- Перед фактической отправкой сообщения:
  - вызывать `planService.CheckLimit(ctx, companyID, <соответствующий ключ>, currentUsage, requiredAmount)` и при `ok == false` возвращать бизнес‑ошибку \"limit exceeded\".
- После успешной отправки (в рамках транзакции use‑case):
  - вызывать `planService.IncrementUsageWithTx(ctx, tx, companyID, <ключ>, delta)` для учёта использования.

Для лимитов по количеству номеров планируется аналогичная проверка при создании/привязке номера компании, когда появится соответствующая сущность.

---

### 4. Сводка по требованиям

- **Метрики**:
  - Не ограничиваются тарифами.
  - Планы могут в будущем управлять уровнем детализации/экспортом, но текущая реализация этого не делает.

- **Лимиты, которые уже работают**:
  - `PlanLimitCallsPerMonth` — ограничение на количество создаваемых звонков через вебхуки телефонии.

- **Лимиты, заложенные в домен, но пока не применённые**:
  - Сообщения: SMS, Telegram, Email.
  - Количество телефонных номеров по странам.

Все новые места использования лимитов должны:

1. Вызывать `CheckLimit` **до** операции.
2. При успехе выполнять бизнес‑действие.
3. В рамках той же транзакции увеличивать usage через `IncrementUsageWithTx`.

