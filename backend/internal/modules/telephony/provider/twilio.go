package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"telephony/internal/domain"
	"telephony/internal/modules/telephony/entity"
)

type twilioProvider struct{}

// NewTwilioProvider создаёт адаптер Twilio, реализующий TelephonyProvider.
func NewTwilioProvider() TelephonyProvider {
	return &twilioProvider{}
}

func (p *twilioProvider) Name() domain.TelephonyName {
	return domain.Twilio
}

func (p *twilioProvider) ParseVoiceStatusWebhook(_ context.Context, req *WebhookRequest) (*CallWebhookEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("twilio: nil webhook request")
	}

	var form entity.TwilioVoiceStatusCallbackForm
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("twilio: empty body")
	}
	if err := json.Unmarshal(req.Body, &form); err != nil {
		return nil, fmt.Errorf("twilio: decode body: %w", err)
	}

	// Нормализуем форму в структуру, с которой удобно работать.
	msg := &entity.TwilioCallStatusCallback{
		CallSid:           form.CallSid,
		ParentCallSid:     form.ParentCallSid,
		AccountSid:        form.AccountSid,
		From:              form.From,
		To:                form.To,
		CallStatus:        form.CallStatus,
		Direction:         form.Direction,
		ApiVersion:        form.APIVersion,
		CallerName:        form.CallerName,
		ForwardedFrom:     form.ForwardedFrom,
		CallbackSource:    form.CallbackSource,
		SequenceNumber:    form.SequenceNumber,
		Timestamp:         form.Timestamp,
		CallDuration:      form.CallDuration,
		Duration:          form.Duration,
		SipResponseCode:   form.SipResponseCode,
		RecordingSid:      form.RecordingSid,
		RecordingUrl:      form.RecordingURL,
		RecordingDuration: form.RecordingDuration,
		Called:            form.Called,
		CalledCity:        form.CalledCity,
		CalledCountry:     form.CalledCountry,
		CalledState:       form.CalledState,
		CalledZip:         form.CalledZip,
		Caller:            form.Caller,
		CallerCity:        form.CallerCity,
		CallerCountry:     form.CallerCountry,
		CallerState:       form.CallerState,
		CallerZip:         form.CallerZip,
		FromCity:          form.FromCity,
		FromCountry:       form.FromCountry,
		FromState:         form.FromState,
		FromZip:           form.FromZip,
		ToCity:            form.ToCity,
		ToCountry:         form.ToCountry,
		ToState:           form.ToState,
		ToZip:             form.ToZip,
	}

	eventStatusTwilioToDomain := map[string]domain.CallEventStatus{
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

	directionTwilioToDomain := map[string]domain.CallDirection{
		"inbound":       domain.CallDirectionInbound,
		"outbound-api":  domain.CallDirectionOutboundApi,
		"outbound-dial": domain.CallDirectionOutboundDial,
	}

	status, ok := eventStatusTwilioToDomain[msg.CallStatus]
	if !ok {
		return nil, fmt.Errorf("twilio: invalid event status %q", msg.CallStatus)
	}

	direction, ok := directionTwilioToDomain[msg.Direction]
	if !ok {
		return nil, fmt.Errorf("twilio: invalid direction %q", msg.Direction)
	}

	timestamp, err := time.Parse(time.RFC1123Z, msg.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("twilio: parse timestamp %q: %w", msg.Timestamp, err)
	}

	recDuration, err := strconv.Atoi(msg.RecordingDuration)
	if err != nil {
		recDuration = 0
	}

	event := &CallWebhookEvent{
		TelephonyName:           domain.Twilio,
		TelephonyAccountID:      msg.AccountSid,
		ExternalCallID:          msg.CallSid,
		ExternalParentCallID:    msg.ParentCallSid,
		FromNumber:              msg.From,
		ToNumber:                msg.To,
		Direction:               direction,
		Status:                  status,
		OccurredAt:              timestamp,
		RecordingID:             msg.RecordingSid,
		RecordingURL:            msg.RecordingUrl,
		RecordingDurationSecond: recDuration,
		FromCountry:             msg.FromCountry,
		FromCity:                msg.FromCity,
		ToCountry:               msg.ToCountry,
		ToCity:                  msg.ToCity,
		Carrier:                 "",
		Trunk:                   "",
		RawPayload:              nil,
	}

	return event, nil
}
