package domain

// DeliveryChannel — канал доставки сообщения (email, SMS, мессенджер и т.д.).
type DeliveryChannel string

const (
	DeliveryChannelEmail     DeliveryChannel = "email"
	DeliveryChannelSMS       DeliveryChannel = "sms"
	DeliveryChannelMessenger DeliveryChannel = "messenger"
)

// OutgoingMessage — унифицированное сообщение для отправки по любому каналу.
// Поля заполняются в зависимости от канала: для email нужны Subject и при необходимости HTML,
// для SMS и мессенджера достаточно To и Body.
type OutgoingMessage struct {
	// To — адрес получателя (email, номер телефона, id в мессенджере).
	To string
	// Subject — тема (в основном для email).
	Subject string
	// Body — текстовое тело сообщения.
	Body string
	// HTML — HTML-версия тела (для email; если пусто, используется Body).
	HTML string
}
