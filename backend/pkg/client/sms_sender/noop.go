package sms_sender

import "context"

// NoopSender — заглушка, не отправляет SMS. Используется при отсутствии провайдера.
type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (NoopSender) Send(ctx context.Context, toPhone, body string) error {
	return nil
}
