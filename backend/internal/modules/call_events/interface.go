package call_events

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	SaveCallEventWithTx(ctx context.Context, tx pgx.Tx, callEvent *domain.CallEvent) srverr.ServerError
	GetEventsByCallIDsWithTx(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, srverr.ServerError)
	GetEventsByCallIDs(ctx context.Context, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, srverr.ServerError)
	GetCallEventsByCallIDWithTx(ctx context.Context, tx pgx.Tx, callID uuid.UUID) ([]*domain.CallEvent, srverr.ServerError)
	GetCallEventsByCallID(ctx context.Context, callID uuid.UUID) ([]*domain.CallEvent, srverr.ServerError)
}

type repository interface {
	getEventsByCallID(ctx context.Context, tx pgx.Tx, callID uuid.UUID) ([]*domain.CallEvent, error)
	saveCallEvent(ctx context.Context, tx pgx.Tx, callEvent *domain.CallEvent) error
	getEventsByCallIDs(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, error)
	getCallEventByCallIDStatus(ctx context.Context, tx pgx.Tx, callID uuid.UUID, eventStatus domain.CallEventStatus) (*domain.CallEvent, error)
}
