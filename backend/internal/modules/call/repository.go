package call

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"
	"telephony/pkg/pointer"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

var (
	errRepoCallNotFound = errors.New("call not found")
)

// ------- Call Tree -------

func (r *repositoryImpl) getCallTreeByCallUUID(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallTree, error) {
	// TODO: Возможно стоит в будущем сделать защиту от цикла в каком нибудь кроне, который будет проверять дерево
	const sql = `
WITH RECURSIVE call_tree AS (
  SELECT
      c.id,
      c.company_telephony_id,
      c.parent_call_id,
      c.waiting_for_parent,
      c.external_call_id,
      c.external_parent_call_id,
      c.from_number,
      c.to_number,
      c.direction,
      c.created_at,
      c.updated_at,
      0 AS depth
  FROM call c
  WHERE c.id = $1

  UNION ALL

  SELECT
      ch.id,
      ch.company_telephony_id,
      ch.parent_call_id,
      ch.waiting_for_parent,
      ch.external_call_id,
      ch.external_parent_call_id,
      ch.from_number,
      ch.to_number,
      ch.direction,
      ch.created_at,
      ch.updated_at,
      ct.depth + 1 AS depth
  FROM call ch
  JOIN call_tree ct
    ON ch.company_telephony_id = ct.company_telephony_id
   AND ch.external_parent_call_id = ct.external_call_id
)
SELECT
      ct.id,
      ct.company_telephony_id,
      ct.parent_call_id,
      ct.waiting_for_parent,
      ct.external_call_id,
      ct.external_parent_call_id,
      ct.from_number,
      ct.to_number,
      ct.direction,
      ct.created_at,
      ct.updated_at,
      ct.depth
  FROM call_tree ct
ORDER BY depth, created_at
`
	rows, err := tx.Query(ctx, sql, callID)
	if err != nil {
		return nil, fmt.Errorf("call.getCallTreeByCallUUID: %w", err)
	}
	defer rows.Close()

	nodesByExternalID := make(map[string]*domain.CallTree)
	nodesByID := make(map[uuid.UUID]*domain.CallTree)
	var root *domain.CallTree

	for rows.Next() {
		var call domain.Call
		var parentCallID *uuid.UUID
		var depth int

		err = rows.Scan(
			&call.ID,
			&call.CompanyTelephonyID,
			&parentCallID,
			&call.WaitingForParent,
			&call.ExternalCallID,
			&call.ExternalParentCallID,
			&call.FromNumber,
			&call.ToNumber,
			&call.Direction,
			&call.CreatedAt,
			&call.UpdatedAt,
			&depth,
		)
		if err != nil {
			return nil, fmt.Errorf("call.getCallTreeByCallUUID: %w", err)
		}
		call.ParentCallID = pointer.FromPtr(parentCallID)

		node := &domain.CallTree{
			Call:     &call,
			Children: nil,
		}

		nodesByExternalID[call.ExternalCallID] = node
		nodesByID[call.ID] = node

		if call.ID == callID {
			root = node
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("call.getCallTreeByCallUUID: %w", err)
	}
	if root == nil {
		return nil, errRepoCallNotFound
	}

	for _, node := range nodesByExternalID {
		parentExternalID := node.Call.ExternalParentCallID
		if parentExternalID == "" {
			continue
		}
		parent := nodesByExternalID[parentExternalID]
		if parent == nil {
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return root, nil

}

// ------- Calls -------

func (r *repositoryImpl) saveCall(ctx context.Context, tx pgx.Tx, call *domain.Call) error {
	var query = `
	INSERT INTO call 
	    (id, company_telephony_id, external_call_id, external_parent_call_id, to_number, from_number, parent_call_id, waiting_for_parent, direction)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var parentCallID *uuid.UUID
	if call.ParentCallID != uuid.Nil {
		parentCallID = &call.ParentCallID
	}

	_, err := tx.Exec(
		ctx,
		query,
		call.ID,
		call.CompanyTelephonyID,
		call.ExternalCallID,
		call.ExternalParentCallID,
		call.ToNumber,
		call.FromNumber,
		parentCallID,
		call.WaitingForParent,
		call.Direction,
	)
	if err != nil {
		return fmt.Errorf("call.saveCall: %w", err)
	}
	return nil
}

func (r *repositoryImpl) getCallByCompanyTelephonyIDAndExternalCallID(
	ctx context.Context,
	tx pgx.Tx,
	companyTelephonyID uuid.UUID,
	externalCallID string,
) (*domain.Call, error) {
	var query = `
	SELECT 
		id,
		company_telephony_id,
		external_call_id, 
		external_parent_call_id,
		from_number,
		to_number,
		parent_call_id,
		waiting_for_parent,
		direction,
		created_at,
		updated_at
    FROM call 
    	WHERE company_telephony_id = $1 AND external_call_id = $2
`
	row := tx.QueryRow(ctx, query, companyTelephonyID, externalCallID)
	var call domain.Call

	var parentCallID *uuid.UUID

	err := row.Scan(
		&call.ID,
		&call.CompanyTelephonyID,
		&call.ExternalCallID,
		&call.ExternalParentCallID,
		&call.FromNumber,
		&call.ToNumber,
		&parentCallID,
		&call.WaitingForParent,
		&call.Direction,
		&call.CreatedAt,
		&call.UpdatedAt,
	)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCallNotFound
		}
		return nil, fmt.Errorf("call.getCallByExternalCallID: %w", err)
	}
	call.ParentCallID = pointer.FromPtr(parentCallID)

	return &call, nil
}

func (r *repositoryImpl) getChillCallsByCompanyTelephonyIDAndExternalCallID(
	ctx context.Context,
	tx pgx.Tx,
	companyTelephonyID uuid.UUID,
	externalCallID string,
) ([]*domain.Call, error) {
	const query = `
	SELECT 
	    id,
		company_telephony_id,
		external_call_id, 
		external_parent_call_id,
		from_number,
		to_number,
		parent_call_id,
		waiting_for_parent,
		direction,
		created_at,
		updated_at
	FROM call
		WHERE company_telephony_id = $1 AND external_parent_call_id = $2
`
	rows, err := tx.Query(ctx, query, companyTelephonyID, externalCallID)
	if err != nil {
		return nil, fmt.Errorf("call.getChillCallsByCompanyTelephonyIDAndCallID: %w", err)
	}
	defer rows.Close()

	var calls []*domain.Call
	for rows.Next() {
		var call domain.Call
		err = rows.Scan(
			&call.ID,
			&call.CompanyTelephonyID,
			&call.ExternalCallID,
			&call.ExternalParentCallID,
			&call.FromNumber,
			&call.ToNumber,
			&call.ParentCallID,
			&call.WaitingForParent,
			&call.Direction,
			&call.CreatedAt,
			&call.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("call.getChillCallsByCompanyTelephonyIDAndCallID: %w", err)
		}
		callCopy := call
		calls = append(calls, &callCopy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("call.getChillCallsByCompanyTelephonyIDAndCallID: %w", err)
	}
	return calls, nil
}

func (r *repositoryImpl) updateCallsWaitingForParentAndParentID(ctx context.Context, tx pgx.Tx, calls []*domain.Call) error {
	batch := &pgx.Batch{}
	for _, call := range calls {
		batch.Queue(`UPDATE call SET waiting_for_parent = $1, parent_call_id = $2 WHERE id = $3`, call.WaitingForParent, call.ParentCallID, call.ID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range calls {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("call.UpdateCallsWaitingForParent: %w", err)
		}
	}
	return nil
}

func (r *repositoryImpl) getCall(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Call, error) {
	const query = `
	SELECT 
	    id,
		company_telephony_id,
		external_call_id, 
		external_parent_call_id,
		from_number,
		to_number,
		parent_call_id,
		waiting_for_parent,
		direction,
		created_at,
		updated_at
	FROM call
		WHERE id = $1
`
	row := tx.QueryRow(ctx, query, id)
	var call domain.Call

	var parentCallID *uuid.UUID

	err := row.Scan(
		&call.ID,
		&call.CompanyTelephonyID,
		&call.ExternalCallID,
		&call.ExternalParentCallID,
		&call.FromNumber,
		&call.ToNumber,
		&parentCallID,
		&call.WaitingForParent,
		&call.Direction,
		&call.CreatedAt,
		&call.UpdatedAt,
	)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCallNotFound
		}
		return nil, fmt.Errorf("call.getCall: %w", err)
	}
	call.ParentCallID = pointer.FromPtr(parentCallID)

	return &call, nil
}

// ------- Call Details -------

func (r *repositoryImpl) saveOrUpdateCallDetails(ctx context.Context, tx pgx.Tx, details *domain.CallDetails) error {
	if details == nil {
		return nil
	}
	query := `
	INSERT INTO call_detail (
	  call_id, recording_sid, recording_url, recording_duration,
	  from_country, from_city, to_country, to_city,
	  carrier, trunk
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (call_id) DO UPDATE SET
	  recording_sid      = COALESCE(EXCLUDED.recording_sid, call_detail.recording_sid),
	  recording_url      = COALESCE(EXCLUDED.recording_url, call_detail.recording_url),
	  recording_duration = COALESCE(EXCLUDED.recording_duration, call_detail.recording_duration),
	  updated_at         = NOW()
	RETURNING created_at, updated_at;		
`

	row := tx.QueryRow(
		ctx,
		query,
		details.CallID,
		details.RecordingSid,
		details.RecordingURL,
		details.RecordingDuration,
		details.FromCountry,
		details.FromCity,
		details.ToCountry,
		details.ToCity,
		details.Carrier,
		details.Trunk,
	)

	err := row.Scan(&details.CreatedAt, &details.UpdatedAt)
	if err != nil {
		return fmt.Errorf("call.saveOrUpdateCallDetails: %w", err)
	}

	return nil
}

func (r *repositoryImpl) getCallsDetails(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID]*domain.CallDetails, error) {
	result := make(map[uuid.UUID]*domain.CallDetails, len(callIDs))
	if len(callIDs) == 0 {
		return result, nil
	}

	const query = `
	SELECT 
		call_id, recording_sid, recording_url, recording_duration,
		from_country, from_city, to_country, to_city,
		carrier, trunk, created_at, updated_at
	FROM call_detail
	WHERE call_id = ANY($1)
`
	rows, err := tx.Query(ctx, query, callIDs)
	if err != nil {
		return nil, fmt.Errorf("call.getCallsDetails: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var details domain.CallDetails
		if err = rows.Scan(
			&details.CallID,
			&details.RecordingSid,
			&details.RecordingURL,
			&details.RecordingDuration,
			&details.FromCountry,
			&details.FromCity,
			&details.ToCountry,
			&details.ToCity,
			&details.Carrier,
			&details.Trunk,
			&details.CreatedAt,
			&details.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("call.getCallsDetails: %w", err)
		}
		detailsCopy := details
		result[details.CallID] = &detailsCopy
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("call.getCallsDetails: %w", err)
	}

	return result, nil
}

func (r *repositoryImpl) getCallDetails(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallDetails, error) {
	const query = `
	SELECT 
		call_id, recording_sid, recording_url, recording_duration,
		from_country, from_city, to_country, to_city,
		carrier, trunk, created_at, updated_at
	FROM call_detail
	WHERE call_id = $1
`
	row := tx.QueryRow(ctx, query, callID)
	var details domain.CallDetails
	err := row.Scan(
		&details.CallID,
		&details.RecordingSid,
		&details.RecordingURL,
		&details.RecordingDuration,
		&details.FromCountry,
		&details.FromCity,
		&details.ToCountry,
		&details.ToCity,
		&details.Carrier,
		&details.Trunk,
		&details.CreatedAt,
		&details.UpdatedAt,
	)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCallNotFound
		}
		return nil, fmt.Errorf("call.getCallDetails: %w", err)
	}
	return &details, nil
}
