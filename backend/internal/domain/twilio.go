package domain

type TwilioCallStatusCallback struct {
	// CallSid — уникальный идентификатор звонка в Twilio.
	// Один и тот же CallSid используется на протяжении всего lifecycle звонка.
	CallSid string

	// ParentCallSid — SID родительского звонка, если этот звонок — leg от <Dial>.
	// Отсутствует для первого (root) звонка.
	ParentCallSid string

	// AccountSid — SID аккаунта Twilio, которому принадлежит звонок.
	// Всегда начинается с "AC".
	AccountSid string

	// From — инициатор звонка.
	// Возможные форматы: E.164 номер (+15551234567) или client идентификатор (client:alice).
	From string

	// To — получатель звонка.
	// Возможные форматы: E.164 номер или client идентификатор.
	To string

	// CallStatus — текущий статус звонка на момент события.
	// Возможные значения: queued, initiated, ringing, in-progress, completed, busy, failed, no-answer.
	CallStatus string

	// Direction — направление звонка.
	// Возможные значения: inbound, outbound-api, outbound-dial.
	Direction string

	// ApiVersion — версия Twilio Voice API, которая обслуживает звонок (например, "2010-04-01").
	ApiVersion string

	// CallerName — CNAM (Caller ID Name).
	// Заполняется, если включён VoiceCallerIdLookup и оператор передал CNAM.
	// Может отсутствовать.
	CallerName string

	// ForwardedFrom — исходный номер, если звонок был переадресован.
	// Зависит от оператора, часто отсутствует.
	ForwardedFrom string

	// CallbackSource — источник webhook.
	// Для StatusCallback всегда "call-progress-events".
	CallbackSource string

	// SequenceNumber — порядковый номер события для данного CallSid.
	// Начинается с 0.
	// НЕ гарантирует порядок доставки HTTP запросов.
	SequenceNumber string

	// Timestamp — время генерации события Twilio.
	// Формат: RFC 2822, UTC (например, "Tue, 06 Jan 2026 10:15:30 +0000").
	Timestamp string

	// CallDuration — длительность разговора в секундах.
	// Присутствует только для события completed.
	CallDuration string

	// Duration — billing duration в минутах.
	// Используется для биллинга.
	// Присутствует только для completed.
	Duration string

	// SipResponseCode — финальный SIP код завершения звонка.
	// Примеры: 200 (успешно), 404 (номер недоступен), 486 (busy), 487 (timeout).
	SipResponseCode string

	// RecordingSid — SID записи разговора.
	// Присутствует только если Record=true.
	RecordingSid string

	// RecordingUrl — URL записи разговора.
	// Может быть недоступен сразу после completed.
	RecordingUrl string

	// RecordingDuration — длительность записи в секундах.
	// Присутствует только если была запись.
	RecordingDuration string

	// Called — номер, на который совершен звонок (целевой номер).
	Called string

	// CalledCity — город, связанный с номером Called.
	// Может отсутствовать.
	CalledCity string

	// CalledCountry — страна, связанная с номером Called.
	// Двухбуквенный код ISO (например, "TH").
	CalledCountry string

	// CalledState — регион или штат, связанный с номером Called.
	// Может отсутствовать.
	CalledState string

	// CalledZip — почтовый индекс, связанный с номером Called.
	// Может отсутствовать.
	CalledZip string

	// Caller — номер инициатора звонка (может совпадать с From).
	Caller string

	// CallerCity — город, связанный с номером Caller.
	CallerCity string

	// CallerCountry — страна, связанная с номером Caller.
	// Двухбуквенный код ISO (например, "US").
	CallerCountry string

	// CallerState — регион или штат, связанный с номером Caller.
	CallerState string

	// CallerZip — почтовый индекс, связанный с номером Caller.
	CallerZip string

	// FromCity — город, связанный с номером From (инициатор звонка).
	FromCity string

	// FromCountry — страна, связанная с номером From.
	FromCountry string

	// FromState — регион или штат, связанный с номером From.
	FromState string

	// FromZip — почтовый индекс, связанный с номером From.
	FromZip string

	// ToCity — город, связанный с номером To (получатель звонка).
	ToCity string

	// ToCountry — страна, связанная с номером To.
	ToCountry string

	// ToState — регион или штат, связанный с номером To.
	ToState string

	// ToZip — почтовый индекс, связанный с номером To.
	ToZip string
}
