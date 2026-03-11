package sms_sender

import "context"

// Sender отправляет SMS на указанный номер.
type Sender interface {
	Send(ctx context.Context, toPhone, body string) error
}
