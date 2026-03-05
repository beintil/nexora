package call

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/runner"
	srverr "telephony/internal/shared/server_error"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler interface {
	handleListCalls(w http.ResponseWriter, r *http.Request)
	handleGetCallByID(w http.ResponseWriter, r *http.Request)

	runner.Runner
}

type Service interface {
	GetCallTreeByCallUUIDWithTx(ctx context.Context, tx pgx.Tx, callUUID uuid.UUID) (*domain.CallTree, srverr.ServerError)
	SaveUpdateCall(ctx context.Context, call *domain.CallWorker) srverr.ServerError
	SaveUpdateCallWithTX(ctx context.Context, tx pgx.Tx, call *domain.CallWorker) srverr.ServerError

	// For read APIs
	GetCallTreeByCallUUIDByCompanyUUID(ctx context.Context, companyID uuid.UUID, callUUID uuid.UUID) (*domain.CallTree, srverr.ServerError)
	ListCompanyCalls(ctx context.Context, filters *domain.CallListFilters) (*domain.CallListPage, srverr.ServerError)

	// Metrics
	GetCompanyCallMetrics(ctx context.Context, companyID uuid.UUID, from, to time.Time) (*domain.CallMetrics, *domain.CallMetricsTimeseries, srverr.ServerError)
}

type repository interface {
	getCallTreeByCallUUID(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallTree, error)
	getCallTreeByCallUUIDForCompany(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, callID uuid.UUID) (*domain.CallTree, error)
	getChillCallsByCompanyTelephonyIDAndExternalCallID(ctx context.Context, tx pgx.Tx, companyTelephonyID uuid.UUID, externalCallID string) ([]*domain.Call, error)
	updateCallsWaitingForParentAndParentID(ctx context.Context, tx pgx.Tx, calls []*domain.Call) error
	saveOrUpdateCallDetails(ctx context.Context, tx pgx.Tx, details *domain.CallDetails) error
	getCallByCompanyTelephonyIDAndExternalCallID(ctx context.Context, tx pgx.Tx, companyTelephonyID uuid.UUID, externalCallID string) (*domain.Call, error)
	getCall(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Call, error)
	saveCall(ctx context.Context, tx pgx.Tx, call *domain.Call) error
	getCallDetails(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallDetails, error)
	getCallsDetails(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID]*domain.CallDetails, error)

	listCompanyCalls(ctx context.Context, tx pgx.Tx, filters *domain.CallListFilters) ([]*domain.CallSummary, int, error)

	// Metrics
	getCallMetrics(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, from, to time.Time, answered domain.CallEventStatus, missed []domain.CallEventStatus) (*domain.CallMetrics, error)
	getCallMetricsTimeseries(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, from, to time.Time, answered domain.CallEventStatus, missed []domain.CallEventStatus) (*domain.CallMetricsTimeseries, error)
}
