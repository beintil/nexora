package message_provider

import (
	"context"
	"telephony/internal/domain"
	"telephony/internal/modules/message_delivery"
	"telephony/pkg/client/email_sender"
)

type emailSenderAdapter struct {
	channel domain.DeliveryChannel
	client  email_sender.Sender
}

// NewEmailSenderAdapter возвращает Sender для канала email, оборачивающий pkg/client/email_sender.
func NewEmailSenderAdapter(client email_sender.Sender) message_delivery.Sender {
	if client == nil {
		return nil
	}
	return &emailSenderAdapter{channel: domain.DeliveryChannelEmail, client: client}
}

func (a *emailSenderAdapter) Channel() domain.DeliveryChannel {
	return a.channel
}

func (a *emailSenderAdapter) Send(ctx context.Context, msg *domain.OutgoingMessage) error {
	if msg == nil {
		return nil
	}
	body := msg.HTML
	if body == "" {
		body = msg.Body
	}
	return a.client.Send(ctx, email_sender.Message{
		To:      msg.To,
		Subject: msg.Subject,
		HTML:    body,
		Text:    msg.Body,
	})
}
