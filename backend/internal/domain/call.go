package domain

import (
	"time"

	"github.com/google/uuid"
)

type CallTree struct {
	Call     *Call
	Children []*CallTree
}

func (t *CallTree) ApplyDetails(details map[uuid.UUID]*CallDetails) {
	if t == nil {
		return
	}
	t.applyDetails(details)
}

func (t *CallTree) applyDetails(details map[uuid.UUID]*CallDetails) {
	if t == nil || t.Call == nil {
		return
	}
	if details != nil {
		if d, ok := details[t.Call.ID]; ok {
			t.Call.Details = d
		}
	}
	for _, ch := range t.Children {
		ch.applyDetails(details)
	}
}

func (t *CallTree) ApplyEvents(events map[uuid.UUID][]*CallEvent) {
	if t == nil {
		return
	}
	t.applyEvents(events)
}

func (t *CallTree) applyEvents(events map[uuid.UUID][]*CallEvent) {
	if t == nil || t.Call == nil {
		return
	}
	if events != nil {
		if evs, ok := events[t.Call.ID]; ok {
			t.Call.Events = evs
		}
	}
	for _, ch := range t.Children {
		ch.applyEvents(events)
	}
}

func (t *CallTree) CallIDs() []uuid.UUID {
	if t == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0)
	t.collectIDs(&ids)
	return ids
}

func (t *CallTree) collectIDs(out *[]uuid.UUID) {
	if t == nil || t.Call == nil {
		return
	}
	*out = append(*out, t.Call.ID)
	for _, ch := range t.Children {
		ch.collectIDs(out)
	}
}

type CallWorker struct {
	*Call
	Event *CallEvent // Событие звонка

	// Служебные поля. Не сохранять в бд
	TelephonyAccountID string // ID аккаунта телефонии, например Twilio AccountSid
}

type Call struct {
	ID                 uuid.UUID // Идентификатор звонка в нашей системе
	ParentCallID       uuid.UUID // ID родительского звонка, если есть перенапровление родительского звонка. Пусто в случае родительского звонка
	CompanyTelephonyID uuid.UUID // ID связи компании с телефонией

	ExternalCallID       string // Уникальный ID звонка, полученный от провайдера телефонии (например, Twilio CallSid)
	ExternalParentCallID string // SID родительского звонка с телефонии, если есть перенаправление родительского звонка. Пусто в случае родительского звонка

	WaitingForParent bool          // Ждёт ли звонок родительского звонка
	FromNumber       string        // Номер телефона кто звонит
	ToNumber         string        // Номер телефона кому звонит
	Direction        CallDirection // Направление звонка
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Events  []*CallEvent // Список событий звонка
	Details *CallDetails // Детали звонка
}

type CallEvent struct {
	ID        uuid.UUID       // Идентификатор события
	CallID    uuid.UUID       // ID звонка к которому относится событие
	Status    CallEventStatus // Статус события
	Timestamp time.Time       // Время события
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CallDetails struct {
	CallID            uuid.UUID // ID звонка к которому относятся детали
	RecordingSid      string    // Идентификатор записи разговора у провайдера телефонии
	RecordingURL      string    // Ссылка на файл записи разговора
	RecordingDuration int       // Длительность записи разговора в секундах
	FromCountry       string    // Страна номера инициатора звонка (ISO 3166-1 alpha-2)
	FromCity          string    // Город инициатора звонка
	ToCountry         string    // Страна номера получателя звонка (ISO 3166-1 alpha-2)
	ToCity            string    // Город получателя звонка
	Carrier           string    // Оператор связи (carrier)
	Trunk             string    // trunk / SIP-линия / DID, через которую прошёл звонок
	CreatedAt         time.Time // Время создания записи
	UpdatedAt         time.Time // Время последнего обновления
}
