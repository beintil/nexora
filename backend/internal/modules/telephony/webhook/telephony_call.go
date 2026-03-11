package webhook

import (
	"context"

	"telephony/internal/domain"
	"telephony/internal/modules/telephony/provider"
	"telephony/internal/modules/telephony_ingestion_pipeline"
	srverr "telephony/internal/shared/server_error"
)

type telephonyCall struct {
	telephonyIngPipService telephony_ingestion_pipeline.Service
	providerRegistry       provider.Registry
}

func NewTelephonyCall(
	telephonyIngPipService telephony_ingestion_pipeline.Service,
	providerRegistry provider.Registry,
) TelephonyCall {
	return &telephonyCall{
		telephonyIngPipService: telephonyIngPipService,
		providerRegistry:       providerRegistry,
	}
}

var (
	telephonyErrCallNotValid srverr.ErrorTypeBadRequest = "telephony request is not valid"
)

// HandleWebhookRoute — Здесь выбирается провайдер, парсится вебхук и вызывается ingestion-пайплайн.
func (m *telephonyCall) HandleWebhookRoute(ctx context.Context, telephony domain.TelephonyName, req *provider.WebhookRequest) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.HandleWebhookRoute/nil_request")
	}

	p, ok := m.providerRegistry.GetProvider(telephony)
	if !ok {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.HandleWebhookRoute/provider_not_found")
	}

	event, err := p.ParseVoiceStatusWebhook(ctx, req)
	if err != nil {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.HandleWebhookRoute/parse").
			SetError(err.Error())
	}

	sErr := m.telephonyIngPipService.HandleWebhookEvent(ctx, event)
	if sErr != nil {
		return sErr
	}

	return nil
}
