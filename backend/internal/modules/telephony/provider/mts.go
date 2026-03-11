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

type mtsProvider struct{}

// NewMTSProvider создаёт адаптер МТС.
func NewMTSProvider() TelephonyProvider {
	return &mtsProvider{}
}

func (p *mtsProvider) Name() domain.TelephonyName {
	return domain.MTS
}

func (p *mtsProvider) ParseVoiceStatusWebhook(_ context.Context, req *WebhookRequest) (*CallWebhookEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("mts: nil webhook request")
	}

	var payload entity.MTSWebhook
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("mts: empty body")
	}
	if err := json.Unmarshal(req.Body, &payload); err != nil {
		return nil, fmt.Errorf("mts: decode body: %w", err)
	}

	eventStatusMTSToDomain := map[string]domain.CallEventStatus{
		"queued":      domain.CallEventStatusQueued,
		"initiated":   domain.CallEventStatusInitiated,
		"ringing":     domain.CallEventStatusRinging,
		"in-progress": domain.CallEventStatusInProgress,
		"answered":    domain.CallEventStatusInProgress,
		"completed":   domain.CallEventStatusCompleted,
		"end":         domain.CallEventStatusCompleted,
		"busy":        domain.CallEventStatusBusy,
		"failed":      domain.CallEventStatusFailed,
		"no-answer":   domain.CallEventStatusNoAnswer,
		"canceled":    domain.CallEventStatusCanceled,
		"cancelled":   domain.CallEventStatusCanceled,
		"timeout":     domain.CallEventStatusTimeout,
	}

	directionMTSToDomain := map[string]domain.CallDirection{
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
	status, ok := eventStatusMTSToDomain[statusKey]
	if !ok {
		return nil, fmt.Errorf("mts: invalid event status %q", payload.Status)
	}

	dirKey := strings.ToLower(strings.TrimSpace(payload.Direction))
	direction, ok := directionMTSToDomain[dirKey]
	if !ok {
		direction = domain.CallDirectionInbound
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.Timestamp))
	if err != nil {
		return nil, fmt.Errorf("mts: parse timestamp %q: %w", payload.Timestamp, err)
	}

	recDuration := 0
	if d := strings.TrimSpace(payload.RecordingDur); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			recDuration = v
		}
	}

	event := &CallWebhookEvent{
		TelephonyName:           domain.MTS,
		TelephonyAccountID:      payload.AccountID,
		ExternalCallID:          payload.CallID,
		ExternalParentCallID:    payload.ParentCallID,
		FromNumber:              payload.From,
		ToNumber:                payload.To,
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
