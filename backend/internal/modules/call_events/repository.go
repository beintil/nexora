package call_events

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repositoryImpl struct {
}

func NewRepository() repository {
	return &repositoryImpl{}
}

var (
	errRepoCallEventNotFound = errors.New("call event not found")
)

func (r *repositoryImpl) saveCallEvent(ctx context.Context, tx pgx.Tx, callEvent *domain.CallEvent) error {
	query := `
	INSERT INTO call_events (id, call_id, status, timestamp) VALUES ($1, $2, $3, $4)
	`
	_, err := tx.Exec(ctx, query, callEvent.ID, callEvent.CallID, callEvent.Status, callEvent.Timestamp)
	if err != nil {
		return fmt.Errorf("repository.saveCallEvent/Exec: %w", err)
	}
	return nil
}

func (r *repositoryImpl) getCallEventByCallIDStatus(ctx context.Context, tx pgx.Tx, callID uuid.UUID, eventStatus domain.CallEventStatus) (*domain.CallEvent, error) {
	query := `
SELECT id, status, timestamp FROM call_events WHERE call_id = $1 AND status = $2
`
	var callEvent domain.CallEvent
	row := tx.QueryRow(ctx, query, callID, eventStatus)
	err := row.Scan(&callEvent.ID, &callEvent.Status, &callEvent.Timestamp)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCallEventNotFound
		}
		return nil, fmt.Errorf("repository.getCallEventByCallIDStatus/Scan: %w", err)
	}
	return &callEvent, nil
}

func (r *repositoryImpl) getEventsByCallID(ctx context.Context, tx pgx.Tx, callID uuid.UUID) ([]*domain.CallEvent, error) {
	query := `
SELECT id, status, timestamp FROM call_events WHERE call_id = $1
`
	rows, err := tx.Query(ctx, query, callID)
	if err != nil {
		return nil, fmt.Errorf("repository.getEventsByCallID/Query: %w", err)
	}
	defer rows.Close()

	var callEvents []*domain.CallEvent
	for rows.Next() {
		var callEvent domain.CallEvent
		err = rows.Scan(&callEvent.ID, &callEvent.Status, &callEvent.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("repository.getEventsByCallID/Scan: %w", err)
		}
		callEvents = append(callEvents, &callEvent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.getEventsByCallID/Rows: %w", err)
	}
	return callEvents, nil
}

func (r *repositoryImpl) getEventsByCallIDs(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID][]*domain.CallEvent, error) {
	result := make(map[uuid.UUID][]*domain.CallEvent, len(callIDs))
	if len(callIDs) == 0 {
		return result, nil
	}

	query := `
SELECT call_id, id, status, timestamp
FROM call_events
WHERE call_id = ANY($1)
ORDER BY call_id, timestamp, id
`
	rows, err := tx.Query(ctx, query, callIDs)
	if err != nil {
		return nil, fmt.Errorf("repository.getEventsByCallIDs/Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var callID uuid.UUID
		var ev domain.CallEvent
		if err = rows.Scan(&callID, &ev.ID, &ev.Status, &ev.Timestamp); err != nil {
			return nil, fmt.Errorf("repository.getEventsByCallIDs/Scan: %w", err)
		}
		ev.CallID = callID
		result[callID] = append(result[callID], &ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.getEventsByCallIDs/Rows: %w", err)
	}
	return result, nil
}
