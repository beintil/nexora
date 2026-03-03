package telephony_ingestion_pipeline

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
)

type Service interface {
	CallWorker(ctx context.Context, call *domain.CallWorker, telephony domain.TelephonyName) srverr.ServerError
}
