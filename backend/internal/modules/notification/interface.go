package notification

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
)

type Service interface {
	GetNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.Notification, srverr.ServerError)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) srverr.ServerError
	CreateNotification(ctx context.Context, n *domain.Notification) srverr.ServerError
}

type Handler interface {
	handleGetNotifications(w http.ResponseWriter, r *http.Request)
	handleMarkAsRead(w http.ResponseWriter, r *http.Request)
}
