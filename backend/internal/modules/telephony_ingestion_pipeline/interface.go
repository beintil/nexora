package telephony_ingestion_pipeline

import (
	"context"

	"telephony/internal/modules/telephony/provider"
	srverr "telephony/internal/shared/server_error"
)

type Service interface {
	// HandleWebhookEvent — новый основной метод ingestion-пайплайна,
	// принимающий нормализованное событие от провайдера телефонии.
	HandleWebhookEvent(ctx context.Context, event *provider.CallWebhookEvent) srverr.ServerError
}
