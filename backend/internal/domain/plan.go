package domain

import (
	"time"

	"github.com/google/uuid"
)

type CompanyPlan struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	PlanID    uuid.UUID
	IsActive  bool
	StartedAt time.Time
	EndsAt    time.Time

	Plan           *Plan
	PlansUsage     []*PlanUsage
	PlansUsageOver []*PlanUsageOver
}

type Plan struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	IsActive    bool
	SortOrder   int
	Description string

	PlanLimits []*PlanLimit
}

type PlanLimit struct {
	PlanID    uuid.UUID
	LimitType PlanLimitType
	Value     int64
	Extra     map[string]any
}

type PlanUsage struct {
	CompanyPlanID uuid.UUID
	LimitType     PlanLimitType
	Value         int64
}

type PlanUsageOver struct {
	CompanyID   uuid.UUID
	LimitType   PlanLimitType
	IsActive    bool
	Value       int64
	StartPeriod time.Time
	EndPeriod   *time.Time
}
