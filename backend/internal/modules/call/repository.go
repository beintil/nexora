package call

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"
	"telephony/pkg/pointer"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

var (
	errRepoCallNotFound     = errors.New("call not found")
	errRepoCallMetricsQuery = errors.New("call metrics query failed")
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

func (r *repositoryImpl) getCallTreeByCallUUIDForCompany(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, callID uuid.UUID) (*domain.CallTree, error) {
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
  JOIN company_telephony ct ON ct.id = c.company_telephony_id
  WHERE c.id = $1 AND ct.company_id = $2

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
	rows, err := tx.Query(ctx, sql, callID, companyID)
	if err != nil {
		return nil, fmt.Errorf("call.getCallTreeByCallUUIDForCompany: %w", err)
	}
	defer rows.Close()

	nodesByExternalID := make(map[string]*domain.CallTree)
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
			return nil, fmt.Errorf("call.getCallTreeByCallUUIDForCompany: %w", err)
		}
		call.ParentCallID = pointer.FromPtr(parentCallID)

		node := &domain.CallTree{
			Call:     &call,
			Children: nil,
		}

		nodesByExternalID[call.ExternalCallID] = node

		if call.ID == callID {
			root = node
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("call.getCallTreeByCallUUIDForCompany: %w", err)
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
		calls = append(calls, &call)
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

func (r *repositoryImpl) listCompanyCalls(ctx context.Context, tx pgx.Tx, filters *domain.CallListFilters) ([]*domain.CallSummary, int, error) {
	if filters == nil {
		return nil, 0, fmt.Errorf("call.listCompanyCalls: filters is nil")
	}

	args := make([]any, 0, 8)
	conditions := make([]string, 0, 8)

	// company_id is mandatory; только корневые (родительские) звонки — дочерние показываются внутри карточки
	args = append(args, filters.CompanyID)
	conditions = append(conditions, "ct.company_id = $1")
	conditions = append(conditions, "c.parent_call_id IS NULL")
	argPos := 2

	if filters.From != nil {
		conditions = append(conditions, fmt.Sprintf("c.created_at >= $%d", argPos))
		args = append(args, *filters.From)
		argPos++
	}
	if filters.To != nil {
		conditions = append(conditions, fmt.Sprintf("c.created_at <= $%d", argPos))
		args = append(args, *filters.To)
		argPos++
	}
	if filters.Direction != nil {
		conditions = append(conditions, fmt.Sprintf("c.direction = $%d", argPos))
		args = append(args, *filters.Direction)
		argPos++
	}
	if filters.CompanyTelephonyID != nil {
		conditions = append(conditions, fmt.Sprintf("c.company_telephony_id = $%d", argPos))
		args = append(args, *filters.CompanyTelephonyID)
		argPos++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("le.status = $%d", argPos))
		args = append(args, *filters.Status)
		argPos++
	}

	queryBuilder := &strings.Builder{}
	queryBuilder.WriteString(`
WITH last_events AS (
  SELECT DISTINCT ON (ce.call_id)
    ce.call_id,
    ce.status,
    ce.timestamp
  FROM call_events ce
  ORDER BY ce.call_id, ce.timestamp DESC, ce.id DESC
)
SELECT
  c.id,
  c.company_telephony_id,
  c.from_number,
  c.to_number,
  c.direction,
  c.created_at,
  c.updated_at,
  COALESCE(le.status, '') AS last_status,
  COUNT(*) OVER() AS total_count,
  EXISTS (SELECT 1 FROM call c2 WHERE c2.parent_call_id = c.id) AS has_children
FROM call c
JOIN company_telephony ct ON ct.id = c.company_telephony_id
LEFT JOIN last_events le ON le.call_id = c.id
`)

	if len(conditions) > 0 {
		queryBuilder.WriteString("WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
		queryBuilder.WriteString("\n")
	}

	queryBuilder.WriteString("ORDER BY c.created_at DESC\n")

	if filters.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf("LIMIT $%d\n", argPos))
		args = append(args, filters.Limit)
		argPos++
	}
	if filters.Offset > 0 {
		queryBuilder.WriteString(fmt.Sprintf("OFFSET $%d\n", argPos))
		args = append(args, filters.Offset)
		argPos++
	}

	rows, err := tx.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("call.listCompanyCalls: %w", err)
	}
	defer rows.Close()

	summaries := make([]*domain.CallSummary, 0)
	var total int

	for rows.Next() {
		var s domain.CallSummary
		if err := rows.Scan(
			&s.ID,
			&s.CompanyTelephonyID,
			&s.FromNumber,
			&s.ToNumber,
			&s.Direction,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.LastStatus,
			&total,
			&s.HasChildren,
		); err != nil {
			return nil, 0, fmt.Errorf("call.listCompanyCalls/scan: %w", err)
		}
		summaries = append(summaries, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("call.listCompanyCalls/rows: %w", err)
	}

	return summaries, total, nil
}

// ------- Call Metrics -------

func (r *repositoryImpl) getCallMetrics(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, from, to time.Time, answered domain.CallEventStatus, missed []domain.CallEventStatus) (*domain.CallMetrics, error) {
	const summaryQuery = `
WITH last_events AS (
  SELECT DISTINCT ON (ce.call_id)
    ce.call_id,
    ce.status
  FROM call_events ce
  ORDER BY ce.call_id, ce.timestamp DESC, ce.id DESC
)
SELECT
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE le.status = $4) AS answered,
  COUNT(*) FILTER (WHERE le.status = ANY($5)) AS missed
FROM call c
JOIN company_telephony ct ON ct.id = c.company_telephony_id
LEFT JOIN last_events le ON le.call_id = c.id
WHERE ct.company_id = $1
  AND c.created_at >= $2
  AND c.created_at <= $3
  AND c.parent_call_id IS NULL
`
	var total, answeredCount, missedCount int
	row := tx.QueryRow(ctx, summaryQuery, companyID, from, to, answered, missed)
	if err := row.Scan(&total, &answeredCount, &missedCount); err != nil {
		return nil, errRepoCallMetricsQuery
	}

	const byDirQuery = `
SELECT c.direction, COUNT(*)
FROM call c
JOIN company_telephony ct ON ct.id = c.company_telephony_id
WHERE ct.company_id = $1
  AND c.created_at >= $2
  AND c.created_at <= $3
  AND c.parent_call_id IS NULL
GROUP BY c.direction
`
	rows, err := tx.Query(ctx, byDirQuery, companyID, from, to)
	if err != nil {
		return nil, errRepoCallMetricsQuery
	}
	defer rows.Close()

	byDirection := make(map[domain.CallDirection]int)
	for rows.Next() {
		var dir domain.CallDirection
		var cnt int
		if err := rows.Scan(&dir, &cnt); err != nil {
			return nil, errRepoCallMetricsQuery
		}
		byDirection[dir] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, errRepoCallMetricsQuery
	}

	return &domain.CallMetrics{
		CompanyID:   companyID,
		From:        from,
		To:          to,
		Total:       total,
		Answered:    answeredCount,
		Missed:      missedCount,
		ByDirection: byDirection,
	}, nil
}

func (r *repositoryImpl) getCallMetricsTimeseries(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, from, to time.Time, answered domain.CallEventStatus, missed []domain.CallEventStatus) (*domain.CallMetricsTimeseries, error) {
	const query = `
WITH last_events AS (
  SELECT DISTINCT ON (ce.call_id)
    ce.call_id,
    ce.status
  FROM call_events ce
  ORDER BY ce.call_id, ce.timestamp DESC, ce.id DESC
),
day_agg AS (
  SELECT
    date_trunc('day', c.created_at) AS day,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE le.status = $4) AS answered,
    COUNT(*) FILTER (WHERE le.status = ANY($5)) AS missed
  FROM call c
  JOIN company_telephony ct ON ct.id = c.company_telephony_id
  LEFT JOIN last_events le ON le.call_id = c.id
  WHERE ct.company_id = $1
    AND c.created_at >= $2
    AND c.created_at <= $3
    AND c.parent_call_id IS NULL
  GROUP BY date_trunc('day', c.created_at)
  ORDER BY day
)
SELECT day, total, answered, missed FROM day_agg
`
	rows, err := tx.Query(ctx, query, companyID, from, to, answered, missed)
	if err != nil {
		return nil, errRepoCallMetricsQuery
	}
	defer rows.Close()

	var points []*domain.CallMetricsTimeseriesPoint
	for rows.Next() {
		var p domain.CallMetricsTimeseriesPoint
		if err := rows.Scan(&p.Date, &p.Total, &p.Answered, &p.Missed); err != nil {
			return nil, errRepoCallMetricsQuery
		}
		points = append(points, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errRepoCallMetricsQuery
	}
	if points == nil {
		points = []*domain.CallMetricsTimeseriesPoint{}
	}
	return &domain.CallMetricsTimeseries{Points: points}, nil
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
