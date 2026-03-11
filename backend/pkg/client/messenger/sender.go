package messenger

import "context"

// Sender отправляет сообщения в мессенджер (Telegram, WhatsApp и т.д.).
type Sender interface {
	Send(ctx context.Context, recipientID, body string) error
}
