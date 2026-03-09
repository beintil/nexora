package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeSystem NotificationType = "system"
	NotificationTypeTeam   NotificationType = "team"
	NotificationTypeCall   NotificationType = "call"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CompanyID uuid.UUID
	Type      NotificationType
	Title     string
	Message   string
	IsRead    bool
	CreatedAt time.Time
}
