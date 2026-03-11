package messenger

import "context"

// NoopSender — заглушка, не отправляет сообщения. Используется при отсутствии провайдера.
type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (NoopSender) Send(ctx context.Context, recipientID, body string) error {
	return nil
}
