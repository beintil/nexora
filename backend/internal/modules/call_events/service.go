package call_events

import (
	"context"
	"errors"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	repos repository

	pool postgres.Transaction
}

func NewService(
	repos repository,

	pool postgres.Transaction,
) Service {
	return &service{
		repos: repos,

		pool: pool,
	}
}

const (
	ServiceErrorInvalidEvent srverr.ErrorTypeBadRequest = "event_is_not_valid"
)

func (s *service) SaveCallEvent(ctx context.Context, callEvent *domain.CallEvent) srverr.ServerError {
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call_events.SaveCallEvent/BeginTransaction").
			SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	sErr := s.SaveCallEventWithTx(ctx, tx, callEvent)
	if sErr != nil {
		return sErr
	}
	err = s.pool.Commit(ctx, tx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call_events.SaveCallEvent/Commit").
			SetError(err.Error())
	}
	return nil
}

func (s *service) SaveCallEventWithTx(ctx context.Context, tx pgx.Tx, callEvent *domain.CallEvent) srverr.ServerError {
	if callEvent == nil {
		return srverr.NewServerError(ServiceErrorInvalidEvent, "call_events.SaveCallEventWithTx").
			SetDetails("empty callEvent request")
	}
	if callEvent.CallID == uuid.Nil {
		return srverr.NewServerError(ServiceErrorInvalidEvent, "call_events.SaveCallEventWithTx").
			SetDetails("empty callID")
	}
	if callEvent.Status == "" {
		return srverr.NewServerError(ServiceErrorInvalidEvent, "call_events.SaveCallEventWithTx").
			SetDetails("empty status")
	}
	if callEvent.Timestamp.IsZero() {
		return srverr.NewServerError(ServiceErrorInvalidEvent, "call_events.SaveCallEventWithTx").
			SetDetails("empty timestamp")
	}
	if callEvent.ID == uuid.Nil {
		callEvent.ID = uuid.New()
	}
	// Сначала проверем, существует ли уже такой евент у звонка
	_, err := s.repos.getCallEventByCallIDStatus(ctx, tx, callEvent.CallID, callEvent.Status)
	if err == nil { // Если существует, то пропускаем, сохранять нельзя
		return nil
	}
	if !errors.Is(err, errRepoCallEventNotFound) {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call_events.SaveCallEventWithTx/getCallEventByCallIDStatus").
			SetError(err.Error())
	}

	err = s.repos.saveCallEvent(ctx, tx, callEvent)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call_events.SaveCallEventWithTx/saveCallEvent").
			SetError(err.Error())
	}
	return nil
}

func (s *service) GetEventsByCallIDsWithTx(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, srverr.ServerError) {
	events, err := s.repos.getEventsByCallIDs(ctx, tx, callIDs)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call_events.GetEventsByCallIDsWithTx/getEventsByCallIDs").
			SetError(err.Error())
	}
	return events, nil
}

func (s *service) GetEventsByCallIDs(ctx context.Context, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, srverr.ServerError) {
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call_events.GetEventsByCallIDs/BeginTransaction").
			SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	events, sErr := s.GetEventsByCallIDsWithTx(ctx, tx, callIDs)
	if sErr != nil {
		return nil, sErr
	}
	return events, nil
}

func (s *service) GetCallEventsByCallIDWithTx(ctx context.Context, tx pgx.Tx, callID uuid.UUID) ([]*domain.CallEvent, srverr.ServerError) {
	events, err := s.repos.getEventsByCallID(ctx, tx, callID)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call_events.GetCallEventsByCallIDWithTx/getEventsByCallID").
			SetError(err.Error())
	}
	return events, nil
}

func (s *service) GetCallEventsByCallID(ctx context.Context, callID uuid.UUID) ([]*domain.CallEvent, srverr.ServerError) {
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call_events.GetCallEventsByCallID/BeginTransaction").
			SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	events, sErr := s.GetCallEventsByCallIDWithTx(ctx, tx, callID)
	if sErr != nil {
		return nil, sErr
	}
	return events, nil
}
