package message_provider

import (
	"context"
	"telephony/internal/domain"
	"telephony/internal/modules/message_delivery"
	"telephony/pkg/client/sms_sender"
)

type smsSenderAdapter struct {
	channel domain.DeliveryChannel
	client  sms_sender.Sender
}

// NewSMSSenderAdapter возвращает Sender для канала SMS, оборачивающий pkg/client/sms_sender.
func NewSMSSenderAdapter(client sms_sender.Sender) message_delivery.Sender {
	if client == nil {
		return nil
	}
	return &smsSenderAdapter{channel: domain.DeliveryChannelSMS, client: client}
}

func (a *smsSenderAdapter) Channel() domain.DeliveryChannel {
	return a.channel
}

func (a *smsSenderAdapter) Send(ctx context.Context, msg *domain.OutgoingMessage) error {
	if msg == nil {
		return nil
	}
	return a.client.Send(ctx, msg.To, msg.Body)
}
