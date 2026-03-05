package entity

import (
	"fmt"
	"strconv"
	"strings"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
	"time"
)

// MangoWebhook описывает формат webhook-а Mango Office.
// Эти поля напрямую маппятся в общий CallWorker.
type MangoWebhook struct {
	CallUID       string `json:"CALL_UID"`
	ParentCallUID string `json:"PARENT_CALL_UID"`
	AccountID     string `json:"PBX_ID"`

	CallerNumber string `json:"CALLER_NUMBER"`
	NumberDialed string `json:"NUMBER_DIALED"`

	Direction string `json:"DIRECTION"`
	Status    string `json:"STATUS"`

	TimeUTC string `json:"TIME_UTC"`

	RecordingID  string `json:"RECORDING_ID"`
	RecordingURL string `json:"RECORDING_URL"`
	// В документации Mango обычно строка с продолжительностью в секундах.
	RecordingDur string `json:"RECORDING_DURATION"`

	FromCountry string `json:"FROM_COUNTRY"`
	FromCity    string `json:"FROM_CITY"`
	ToCountry   string `json:"TO_COUNTRY"`
	ToCity      string `json:"TO_CITY"`
	Carrier     string `json:"CARRIER"`
	Trunk       string `json:"TRUNK"`
}

// MangoToCallWorker конвертирует MangoWebhook в общий CallWorker.
func MangoToCallWorker(m *MangoWebhook) (*domain.CallWorker, srverr.ServerError) {
	if m == nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MangoToCallWorker").
			SetDetails("empty request")
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

	statusKey := strings.ToLower(strings.TrimSpace(m.Status))
	status, ok := eventStatusMangoToDomain[statusKey]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MangoToCallWorker/InvalidEvent").
			SetDetails(fmt.Sprintf("failed parse event status %v", m.Status))
	}

	dirKey := strings.ToLower(strings.TrimSpace(m.Direction))
	direction, ok := directionMangoToDomain[dirKey]
	if !ok {
		direction = domain.CallDirectionInbound
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(m.TimeUTC))
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.MangoToCallWorker/Parse").
			SetError(err.Error()).
			SetDetails(fmt.Sprintf("failed parse event timestamp %v", m.TimeUTC))
	}

	recDuration := 0
	if d := strings.TrimSpace(m.RecordingDur); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			recDuration = v
		}
	}

	c := domain.CallWorker{
		Call: &domain.Call{
			ExternalParentCallID: m.ParentCallUID,
			ExternalCallID:       m.CallUID,
			FromNumber:           m.CallerNumber,
			ToNumber:             m.NumberDialed,
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
