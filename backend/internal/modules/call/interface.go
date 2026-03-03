package call

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	GetCallTreeByCallUUIDWithTx(ctx context.Context, tx pgx.Tx, callUUID uuid.UUID) (*domain.CallTree, srverr.ServerError)
	SaveUpdateCall(ctx context.Context, call *domain.CallWorker) srverr.ServerError
	SaveUpdateCallWithTX(ctx context.Context, tx pgx.Tx, call *domain.CallWorker) srverr.ServerError
}

type repository interface {
	getCallTreeByCallUUID(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallTree, error)
	getChillCallsByCompanyTelephonyIDAndExternalCallID(ctx context.Context, tx pgx.Tx, companyTelephonyID uuid.UUID, externalCallID string) ([]*domain.Call, error)
	updateCallsWaitingForParentAndParentID(ctx context.Context, tx pgx.Tx, calls []*domain.Call) error
	saveOrUpdateCallDetails(ctx context.Context, tx pgx.Tx, details *domain.CallDetails) error
	getCallByCompanyTelephonyIDAndExternalCallID(ctx context.Context, tx pgx.Tx, companyTelephonyID uuid.UUID, externalCallID string) (*domain.Call, error)
	getCall(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Call, error)
	saveCall(ctx context.Context, tx pgx.Tx, call *domain.Call) error
	getCallDetails(ctx context.Context, tx pgx.Tx, callID uuid.UUID) (*domain.CallDetails, error)
	getCallsDetails(ctx context.Context, tx pgx.Tx, callIDs []uuid.UUID) (map[uuid.UUID]*domain.CallDetails, error)
}
