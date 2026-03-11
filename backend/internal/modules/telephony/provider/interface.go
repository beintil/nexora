package provider

import (
	"context"
	"time"

	"telephony/internal/domain"
)

// CallWebhookEvent описывает уже нормализованное событие от провайдера телефонии.
// Это интеграционная модель: сюда приводится всё многообразие форматов вебхуков
// (JSON, form-data и т.п.), но в едином контракте, удобном для ingestion-пайплайна.
type CallWebhookEvent struct {
	// Провайдер, от которого пришёл вебхук (Twilio, Mango, MTS, Zadarma и т.п.).
	TelephonyName domain.TelephonyName

	// Идентификатор аккаунта/АТС у провайдера.
	// Например, для Mango — PBX_ID, для MTS — account_id, для Twilio — AccountSid.
	TelephonyAccountID string

	// Внешние идентификаторы звонка в системе провайдера.
	ExternalCallID       string
	ExternalParentCallID string

	// Номера телефона в едином, уже нормализованном формате.
	FromNumber string
	ToNumber   string

	// Нормализованное направление и статус звонка.
	Direction domain.CallDirection
	Status    domain.CallEventStatus

	// Время, к которому относится событие провайдера (ringing, answered, finished и т.п.).
	OccurredAt time.Time

	// Информация о записи разговора.
	RecordingID             string
	RecordingURL            string
	RecordingDurationSecond int

	// География, оператор и линия.
	FromCountry string
	FromCity    string
	ToCountry   string
	ToCity      string
	Carrier     string
	Trunk       string

	// Сырой payload для отладки/аудита (если нужно сохранить).
	RawPayload map[string]any
}

// WebhookRequest описывает вебхук провайдера в обобщённом виде,
// без привязки к конкретному HTTP-фреймворку.
type WebhookRequest struct {
	Headers map[string]string
	Query   map[string]string
	Form    map[string]string
	Body    []byte
}

// TelephonyProvider описывает адаптер конкретного провайдера телефонии.
// Его задача — из специфичного вебхука собрать нормализованный CallWebhookEvent.
type TelephonyProvider interface {
	// Name возвращает доменное имя провайдера (Twilio, Mango, MTS, Zadarma и т.п.).
	Name() domain.TelephonyName

	// ParseVoiceStatusWebhook парсит вебхук статуса голосового звонка
	// и возвращает нормализованное событие.
	ParseVoiceStatusWebhook(ctx context.Context, req *WebhookRequest) (*CallWebhookEvent, error)
}

// Registry описывает реестр провайдеров телефонии.
// Через него TelephonyCall и ingestion-пайплайн находят нужную реализацию по имени.
type Registry interface {
	GetProvider(name domain.TelephonyName) (TelephonyProvider, bool)
}

// registry — базовая in-memory реализация Registry.
type registry struct {
	providers map[domain.TelephonyName]TelephonyProvider
}

// NewRegistry создаёт новый реестр провайдеров.
func NewRegistry(providers ...TelephonyProvider) Registry {
	r := &registry{
		providers: make(map[domain.TelephonyName]TelephonyProvider, len(providers)),
	}
	for _, p := range providers {
		if p == nil {
			continue
		}
		r.providers[p.Name()] = p
	}
	return r
}

// GetProvider возвращает провайдера по имени, если он зарегистрирован.
func (r *registry) GetProvider(name domain.TelephonyName) (TelephonyProvider, bool) {
	if r == nil {
		return nil, false
	}
	p, ok := r.providers[name]
	return p, ok
}
