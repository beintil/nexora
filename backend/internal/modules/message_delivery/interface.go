package message_delivery

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
)

// Sender — контракт отправителя по одному каналу (email, SMS, мессенджер и т.д.).
// Реализации живут в pkg/client; модуль принимает список Sender при создании сервиса.
type Sender interface {
	// Channel возвращает канал, который реализует этот отправитель.
	Channel() domain.DeliveryChannel
	// Send отправляет сообщение. Ошибки — типизированные error (из pkg/client).
	Send(ctx context.Context, msg *domain.OutgoingMessage) error
}

// Service — сервис доставки сообщений по выбранным каналам.
type Service interface {
	// Send отправляет сообщение по указанным каналам. Для каждого канала вызывается
	// соответствующий Sender из переданного при инициализации списка.
	// Если канал не зарегистрирован, возвращается ошибка.
	Send(ctx context.Context, msg *domain.OutgoingMessage, channels []domain.DeliveryChannel) srverr.ServerError
}
