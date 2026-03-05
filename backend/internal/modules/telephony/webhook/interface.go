package webhook

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
)

type TelephonyCall interface {
	VoiceStatus(ctx context.Context, telephony domain.TelephonyName, req *domain.CallWorker) srverr.ServerError
}
