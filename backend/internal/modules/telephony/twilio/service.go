package twilio

import (
	"context"
	"fmt"
	"strconv"
	"telephony/internal/domain"
	"telephony/internal/modules/telephony_ingestion_pipeline"
	srverr "telephony/internal/shared/server_error"
	"time"
)

type service struct {
	telephonyIngPipService telephony_ingestion_pipeline.Service
}

func NewService(
	telephonyIngPipService telephony_ingestion_pipeline.Service,
) Service {
	return &service{
		telephonyIngPipService: telephonyIngPipService,
	}
}

func (s *service) VoiceStatus(ctx context.Context, req *domain.TwilioCallStatusCallback) srverr.ServerError {
	callReq, sErr := twilioCallStatusCallbackToCallWorker(req)
	if sErr != nil {
		return sErr
	}
	sErr = s.telephonyIngPipService.CallWorker(ctx, callReq, domain.Twilio)
	if sErr != nil {
		return sErr
	}
	return nil
}

func twilioCallStatusCallbackToCallWorker(t *domain.TwilioCallStatusCallback) (*domain.CallWorker, srverr.ServerError) {
	if t == nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "twilioCallStatusCallbackToCallWorker").
			SetDetails("empty request")
	}
	var eventStatusTwilioToDomain = map[string]domain.CallEventStatus{
		"queued":      domain.CallEventStatusQueued,
		"initiated":   domain.CallEventStatusInitiated,
		"ringing":     domain.CallEventStatusRinging,
		"in-progress": domain.CallEventStatusInProgress,
		"completed":   domain.CallEventStatusCompleted,
		"busy":        domain.CallEventStatusBusy,
		"failed":      domain.CallEventStatusFailed,
		"no-answer":   domain.CallEventStatusNoAnswer,
		"canceled":    domain.CallEventStatusCanceled,
		"timeout":     domain.CallEventStatusTimeout,
	}

	var directionTwilioToDomain = map[string]domain.CallDirection{
		"inbound":       domain.CallDirectionInbound,
		"outbound-api":  domain.CallDirectionOutboundApi,
		"outbound-dial": domain.CallDirectionOutboundDial,
	}

	status, ok := eventStatusTwilioToDomain[t.CallStatus]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "twilioCallStatusCallbackToCallWorker/InvalidEvent").
			SetDetails(fmt.Sprintf("failed parse event status %v", t.CallStatus))
	}
	direction, ok := directionTwilioToDomain[t.Direction]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "twilioCallStatusCallbackToCallWorker/InvalidDirection").
			SetDetails(fmt.Sprintf("failed parse event direction %v", t.Direction))
	}

	timestamp, err := time.Parse(time.RFC1123Z, t.Timestamp)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "twilioCallStatusCallbackToCallWorker/Parse").
			SetError(err.Error()).SetDetails(fmt.Sprintf("failed parse event timestamp %v", t.Timestamp))
	}
	recDuration, err := strconv.Atoi(t.RecordingDuration)
	if err != nil {
		recDuration = 0 // Считаем за 0 секунд
	}

	c := domain.CallWorker{
		Call: &domain.Call{
			ExternalParentCallID: t.ParentCallSid,
			ExternalCallID:       t.CallSid,
			FromNumber:           t.From,
			ToNumber:             t.To,
			Direction:            direction,
			Details: &domain.CallDetails{
				RecordingSid:      t.RecordingSid,
				RecordingURL:      t.RecordingUrl,
				RecordingDuration: recDuration,

				FromCountry: t.FromCountry,
				FromCity:    t.FromCity,

				ToCountry: t.ToCountry,
				ToCity:    t.ToCity,

				Carrier: "",
				Trunk:   "",
			},
		},
		Event: &domain.CallEvent{
			Status:    status,
			Timestamp: timestamp,
		},

		TelephonyAccountID: t.AccountSid,
	}
	return &c, nil
}
