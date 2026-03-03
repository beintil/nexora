package domain

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time

	TelephonesMap map[TelephonyName]*CompanyTelephone
}

type CompanyTelephone struct {
	ID        uuid.UUID
	CompanyID uuid.UUID

	Telephone         *Telephony
	ExternalAccountID string

	CreatedAt time.Time
	UpdatedAt time.Time
}
