package entity

import (
	"fmt"
	"strconv"
	"strings"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
	"time"
)

// ZadarmaWebhook описывает формат webhook-а Zadarma с call_status.
type ZadarmaWebhook struct {
	PBXID        string `json:"pbx_id"`
	CallStatus   string `json:"call_status"`
	CallerNumber string `json:"caller_number"`
	AnswerNumber string `json:"answer_number"`
	Duration     string `json:"duration"`
	Time         string `json:"time"`
}

// ZadarmaToCallWorker конвертирует ZadarmaWebhook в общий CallWorker.
func ZadarmaToCallWorker(z *ZadarmaWebhook) (*domain.CallWorker, srverr.ServerError) {
	if z == nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.ZadarmaToCallWorker").
			SetDetails("empty request")
	}

	eventStatusZadarmaToDomain := map[string]domain.CallEventStatus{
		"start":   domain.CallEventStatusInitiated,
		"answer":  domain.CallEventStatusInProgress,
		"end":     domain.CallEventStatusCompleted,
		"busy":    domain.CallEventStatusBusy,
		"abandon": domain.CallEventStatusNoAnswer,
	}

	statusKey := strings.ToLower(strings.TrimSpace(z.CallStatus))
	status, ok := eventStatusZadarmaToDomain[statusKey]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.ZadarmaToCallWorker/InvalidEvent").
			SetDetails(fmt.Sprintf("failed parse event status %v", z.CallStatus))
	}

	rawTime := strings.TrimSpace(z.Time)
	timestamp, err := time.Parse(time.RFC3339, rawTime)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02 15:04:05", rawTime); err2 == nil {
			timestamp = t2.UTC()
		} else {
			return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.ZadarmaToCallWorker/Parse").
				SetError(err.Error()).
				SetDetails(fmt.Sprintf("failed parse event timestamp %v", z.Time))
		}
	}

	recDuration := 0
	if d := strings.TrimSpace(z.Duration); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			recDuration = v
		}
	}

	telephonyAccountID := strings.TrimSpace(z.AnswerNumber)

	c := domain.CallWorker{
		Call: &domain.Call{
			ExternalParentCallID: "",
			ExternalCallID:       z.PBXID,
			FromNumber:           z.CallerNumber,
			ToNumber:             z.AnswerNumber,
			Direction:            domain.CallDirectionInbound,
			Details: &domain.CallDetails{
				RecordingSid:      "",
				RecordingURL:      "",
				RecordingDuration: recDuration,

				FromCountry: "",
				FromCity:    "",

				ToCountry: "",
				ToCity:    "",

				Carrier: "",
				Trunk:   "",
			},
		},
		Event: &domain.CallEvent{
			Status:    status,
			Timestamp: timestamp,
		},

		TelephonyAccountID: telephonyAccountID,
	}
	return &c, nil
}
