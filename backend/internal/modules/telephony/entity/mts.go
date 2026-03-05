package entity

import (
	"fmt"
	"strconv"
	"strings"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
	"time"
)

// MTSWebhook — нормализованное представление webhook-а МТС.
type MTSWebhook struct {
	CallID       string `json:"call_id"`
	ParentCallID string `json:"parent_call_id"`
	AccountID    string `json:"account_id"`

	From string `json:"from"`
	To   string `json:"to"`

	Status    string `json:"status"`
	Direction string `json:"direction"`

	Timestamp string `json:"timestamp"`

	RecordingID  string `json:"recording_id"`
	RecordingURL string `json:"recording_url"`
	// Строка с длительностью записи в секундах.
	RecordingDur string `json:"recording_duration"`

	FromCountry string `json:"from_country"`
	FromCity    string `json:"from_city"`
	ToCountry   string `json:"to_country"`
	ToCity      string `json:"to_city"`
	Carrier     string `json:"carrier"`
	Trunk       string `json:"trunk"`
}

// MTSToCallWorker конвертирует MTSWebhook в общий CallWorker.
func MTSToCallWorker(m *MTSWebhook) (*domain.CallWorker, srverr.ServerError) {
	if m == nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MTSToCallWorker").
			SetDetails("empty request")
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

	statusKey := strings.ToLower(strings.TrimSpace(m.Status))
	status, ok := eventStatusMTSToDomain[statusKey]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MTSToCallWorker/InvalidEvent").
			SetDetails(fmt.Sprintf("failed parse event status %v", m.Status))
	}

	dirKey := strings.ToLower(strings.TrimSpace(m.Direction))
	direction, ok := directionMTSToDomain[dirKey]
	if !ok {
		direction = domain.CallDirectionInbound
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(m.Timestamp))
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MTSToCallWorker/Parse").
			SetError(err.Error()).
			SetDetails(fmt.Sprintf("failed parse event timestamp %v", m.Timestamp))
	}

	recDuration := 0
	if d := strings.TrimSpace(m.RecordingDur); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			recDuration = v
		}
	}

	c := domain.CallWorker{
		Call: &domain.Call{
			ExternalParentCallID: m.ParentCallID,
			ExternalCallID:       m.CallID,
			FromNumber:           m.From,
			ToNumber:             m.To,
			Direction:            direction,
			Details: &domain.CallDetails{
				RecordingSid:      m.RecordingID,
				RecordingURL:      m.RecordingURL,
				RecordingDuration: recDuration,

				FromCountry: m.FromCountry,
				FromCity:    m.FromCity,

				ToCountry: m.ToCountry,
				ToCity:    m.ToCity,

				Carrier: m.Carrier,
				Trunk:   m.Trunk,
			},
		},
		Event: &domain.CallEvent{
			Status:    status,
			Timestamp: timestamp,
		},

		TelephonyAccountID: m.AccountID,
	}
	return &c, nil
}
