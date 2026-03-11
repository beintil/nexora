package webhook

import (
	"context"

	"telephony/internal/domain"
	"telephony/internal/modules/telephony/provider"
	srverr "telephony/internal/shared/server_error"
)

type TelephonyCall interface {
	// HandleWebhookRoute — новый контракт, принимающий нормализованный WebhookRequest
	// и делегирующий обработку конкретному провайдеру и ingestion-пайплайну.
	HandleWebhookRoute(ctx context.Context, telephony domain.TelephonyName, req *provider.WebhookRequest) srverr.ServerError
}
