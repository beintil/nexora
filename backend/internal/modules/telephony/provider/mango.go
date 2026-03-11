package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"telephony/internal/domain"
	"telephony/internal/modules/telephony/entity"
)

type mangoProvider struct{}

// NewMangoProvider создаёт адаптер Mango Office.
func NewMangoProvider() TelephonyProvider {
	return &mangoProvider{}
}

func (p *mangoProvider) Name() domain.TelephonyName {
	return domain.Mango
}

func (p *mangoProvider) ParseVoiceStatusWebhook(_ context.Context, req *WebhookRequest) (*CallWebhookEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("mango: nil webhook request")
	}

	var payload entity.MangoWebhook
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("mango: empty body")
	}
	if err := json.Unmarshal(req.Body, &payload); err != nil {
		return nil, fmt.Errorf("mango: decode body: %w", err)
	}

	eventStatusMangoToDomain := map[string]domain.CallEventStatus{
		"start":     domain.CallEventStatusInitiated,
		"ringing":   domain.CallEventStatusRinging,
		"answered":  domain.CallEventStatusInProgress,
		"answer":    domain.CallEventStatusInProgress,
		"connected": domain.CallEventStatusInProgress,
		"completed": domain.CallEventStatusCompleted,
		"end":       domain.CallEventStatusCompleted,
		"busy":      domain.CallEventStatusBusy,
		"failed":    domain.CallEventStatusFailed,
		"no-answer": domain.CallEventStatusNoAnswer,
		"abandon":   domain.CallEventStatusNoAnswer,
		"canceled":  domain.CallEventStatusCanceled,
		"cancelled": domain.CallEventStatusCanceled,
		"timeout":   domain.CallEventStatusTimeout,
	}

	directionMangoToDomain := map[string]domain.CallDirection{
		"inbound":  domain.CallDirectionInbound,
		"incoming": domain.CallDirectionInbound,
		"in":       domain.CallDirectionInbound,

		"outbound":  domain.CallDirectionOutboundApi,
		"outgoing":  domain.CallDirectionOutboundApi,
		"out":       domain.CallDirectionOutboundApi,
		"external":  domain.CallDirectionOutboundApi,
		"internal":  domain.CallDirectionInbound,
		"local":     domain.CallDirectionInbound,
		"outbound2": domain.CallDirectionOutboundDial,
	}

	statusKey := strings.ToLower(strings.TrimSpace(payload.Status))
	status, ok := eventStatusMangoToDomain[statusKey]
	if !ok {
		return nil, fmt.Errorf("mango: invalid event status %q", payload.Status)
	}

	dirKey := strings.ToLower(strings.TrimSpace(payload.Direction))
	direction, ok := directionMangoToDomain[dirKey]
	if !ok {
		direction = domain.CallDirectionInbound
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.TimeUTC))
	if err != nil {
		return nil, fmt.Errorf("mango: parse timestamp %q: %w", payload.TimeUTC, err)
	}

	recDuration := 0
	if d := strings.TrimSpace(payload.RecordingDur); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			recDuration = v
		}
	}

	event := &CallWebhookEvent{
		TelephonyName:           domain.Mango,
		TelephonyAccountID:      payload.AccountID,
		ExternalCallID:          payload.CallUID,
		ExternalParentCallID:    payload.ParentCallUID,
		FromNumber:              payload.CallerNumber,
		ToNumber:                payload.NumberDialed,
		Direction:               direction,
		Status:                  status,
		OccurredAt:              timestamp,
		RecordingID:             payload.RecordingID,
		RecordingURL:            payload.RecordingURL,
		RecordingDurationSecond: recDuration,
		FromCountry:             payload.FromCountry,
		FromCity:                payload.FromCity,
		ToCountry:               payload.ToCountry,
		ToCity:                  payload.ToCity,
		Carrier:                 payload.Carrier,
		Trunk:                   payload.Trunk,
		RawPayload:              nil,
	}

	return event, nil
}
