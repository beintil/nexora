package message_delivery

import (
	"context"
	"sync"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
	"telephony/pkg/logger"
)

type service struct {
	sendersByChannel map[domain.DeliveryChannel]Sender

	log logger.Logger
}

const (
	ServiceErrorInvalidMessage      srverr.ErrorTypeBadRequest = "invalid_message"
	ServiceErrorChannelNotSupported srverr.ErrorTypeBadRequest = "channel_not_supported"
)

// NewService создаёт сервис доставки сообщений. Принимает список отправителей
// (email, SMS, мессенджер и т.д.)
func NewService(log logger.Logger, senders ...Sender) Service {
	sendersByChannel := make(map[domain.DeliveryChannel]Sender, len(senders))
	for _, s := range senders {
		if s != nil {
			sendersByChannel[s.Channel()] = s
		}
	}
	return &service{sendersByChannel: sendersByChannel, log: log}
}

func (s *service) Send(ctx context.Context, msg *domain.OutgoingMessage, channels []domain.DeliveryChannel) srverr.ServerError {
	if msg == nil {
		return srverr.NewServerError(ServiceErrorInvalidMessage, "message_delivery.Send/nil_message")
	}
	if len(channels) == 0 {
		return srverr.NewServerError(ServiceErrorInvalidMessage, "message_delivery.Send/empty_channels")
	}
	errCh := make(chan error, len(channels))
	var wg sync.WaitGroup

	for _, ch := range channels {
		sender, ok := s.sendersByChannel[ch]
		if !ok {
			s.log.Errorf("message_delivery.Send/channel_not_supported: %s", ch)
			continue
		}
		wg.Add(1)
		go func(ch domain.DeliveryChannel, sender Sender) {
			defer wg.Done()
			// Если контекст уже отменён — не шлем
			if err := ctx.Err(); err != nil {
				return
			}
			if err := sender.Send(ctx, msg); err != nil {
				s.log.Errorf("message_delivery.Send/send_failed channel=%s: %v", ch, err)
				select {
				case errCh <- err:
				default:
				}
			}
		}(ch, sender)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return srverr.NewServerError(srverr.ErrInternalServerError, "message_delivery.Send/send").
				SetError(err.Error())
		}
	}
	return nil
}
