package message_provider

import (
	"context"
	"telephony/internal/domain"
	"telephony/internal/modules/message_delivery"
	"telephony/pkg/client/messenger"
)

type messengerSenderAdapter struct {
	channel domain.DeliveryChannel
	client  messenger.Sender
}

// NewMessengerSenderAdapter возвращает Sender для канала мессенджера, оборачивающий pkg/client/messenger.
func NewMessengerSenderAdapter(client messenger.Sender) message_delivery.Sender {
	if client == nil {
		return nil
	}
	return &messengerSenderAdapter{channel: domain.DeliveryChannelMessenger, client: client}
}

func (a *messengerSenderAdapter) Channel() domain.DeliveryChannel {
	return a.channel
}

func (a *messengerSenderAdapter) Send(ctx context.Context, msg *domain.OutgoingMessage) error {
	if msg == nil {
		return nil
	}
	return a.client.Send(ctx, msg.To, msg.Body)
}
