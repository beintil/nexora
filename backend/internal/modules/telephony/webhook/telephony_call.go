package webhook

import (
	"context"
	"telephony/internal/domain"
	"telephony/internal/modules/telephony_ingestion_pipeline"
	srverr "telephony/internal/shared/server_error"
)

type telephonyCall struct {
	telephonyIngPipService telephony_ingestion_pipeline.Service
}

func NewTelephonyCall(
	telephonyIngPipService telephony_ingestion_pipeline.Service,
) TelephonyCall {
	return &telephonyCall{
		telephonyIngPipService: telephonyIngPipService,
	}
}

var (
	telephonyErrCallNotValid srverr.ErrorTypeBadRequest = "telephony request is not valid"
)

func (m *telephonyCall) VoiceStatus(ctx context.Context, telephony domain.TelephonyName, req *domain.CallWorker) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/nil_req")
	}
	if req.TelephonyAccountID == "" {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/empty_telephony_account_id")
	}
	if req.ExternalCallID == "" {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/empty_external_call_id")
	}
	if req.FromNumber == "" {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/empty_from_number")
	}
	if req.ToNumber == "" {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/empty_to_number")
	}
	if req.Event == nil {
		return srverr.NewServerError(telephonyErrCallNotValid, "telephony_call.VoiceStatus/nil_event")
	}

	sErr := m.telephonyIngPipService.CallWorker(ctx, req, telephony)
	if sErr != nil {
		return sErr
	}
	return nil

}
