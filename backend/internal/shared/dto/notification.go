package dto

import (
	"telephony/internal/domain"
	"telephony/models"

	"github.com/go-openapi/strfmt"
)

// NotificationDomainToModel конвертирует domain.Notification в models.NotificationResponse.
func NotificationDomainToModel(n *domain.Notification) *models.NotificationResponse {
	if n == nil {
		return nil
	}
	return &models.NotificationResponse{
		ID:        strfmt.UUID(n.ID.String()),
		UserID:    strfmt.UUID(n.UserID.String()),
		CompanyID: strfmt.UUID(n.CompanyID.String()),
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: strfmt.DateTime(n.CreatedAt),
	}
}

// NotificationsDomainToModel конвертирует []domain.Notification в models.NotificationsListResponse.
func NotificationsDomainToModel(nn []*domain.Notification) *models.NotificationsListResponse {
	res := &models.NotificationsListResponse{
		Notifications: make([]*models.NotificationResponse, 0, len(nn)),
	}
	for _, n := range nn {
		res.Notifications = append(res.Notifications, NotificationDomainToModel(n))
	}
	return res
}
