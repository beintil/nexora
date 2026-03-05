package entity

import (
	"fmt"
	"strconv"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"
	"time"
)

// TwilioVoiceStatusCallbackForm описывает «сырой» payload Twilio вебхука,
// который приходит на endpoint. Используется на transport-слое.
type TwilioVoiceStatusCallbackForm struct {
	CallSid           string `json:"CallSid"`
	ParentCallSid     string `json:"ParentCallSid"`
	AccountSid        string `json:"AccountSid"`
	From              string `json:"From"`
	To                string `json:"To"`
	CallStatus        string `json:"CallStatus"`
	Direction         string `json:"Direction"`
	APIVersion        string `json:"ApiVersion"`
	CallerName        string `json:"CallerName"`
	ForwardedFrom     string `json:"ForwardedFrom"`
	CallbackSource    string `json:"CallbackSource"`
	SequenceNumber    string `json:"SequenceNumber"`
	Timestamp         string `json:"Timestamp"`
	CallDuration      string `json:"CallDuration"`
	Duration          string `json:"Duration"`
	SipResponseCode   string `json:"SipResponseCode"`
	RecordingSid      string `json:"RecordingSid"`
	RecordingURL      string `json:"RecordingUrl"`
	RecordingDuration string `json:"RecordingDuration"`

	Called        string `json:"Called"`
	CalledCity    string `json:"CalledCity"`
	CalledCountry string `json:"CalledCountry"`
	CalledState   string `json:"CalledState"`
	CalledZip     string `json:"CalledZip"`

	Caller        string `json:"Caller"`
	CallerCity    string `json:"CallerCity"`
	CallerCountry string `json:"CallerCountry"`
	CallerState   string `json:"CallerState"`
	CallerZip     string `json:"CallerZip"`

	FromCity    string `json:"FromCity"`
	FromCountry string `json:"FromCountry"`
	FromState   string `json:"FromState"`
	FromZip     string `json:"FromZip"`

	ToCity    string `json:"ToCity"`
	ToCountry string `json:"ToCountry"`
	ToState   string `json:"ToState"`
	ToZip     string `json:"ToZip"`
}

// TwilioCallStatusCallback описывает нормализованный запрос вебхука Twilio
// после маппинга из формы (TwilioVoiceStatusCallbackForm).
type TwilioCallStatusCallback struct {
	CallSid           string
	ParentCallSid     string
	AccountSid        string
	From              string
	To                string
	CallStatus        string
	Direction         string
	ApiVersion        string
	CallerName        string
	ForwardedFrom     string
	CallbackSource    string
	SequenceNumber    string
	Timestamp         string
	CallDuration      string
	Duration          string
	SipResponseCode   string
	RecordingSid      string
	RecordingUrl      string
	RecordingDuration string

	Called        string
	CalledCity    string
	CalledCountry string
	CalledState   string
	CalledZip     string

	Caller        string
	CallerCity    string
	CallerCountry string
	CallerState   string
	CallerZip     string

	FromCity    string
	FromCountry string
	FromState   string
	FromZip     string

	ToCity    string
	ToCountry string
	ToState   string
	ToZip     string
}

// TwilioToCallWorker конвертирует TwilioCallStatusCallback в общий CallWorker.
func TwilioToCallWorker(t *TwilioCallStatusCallback) (*domain.CallWorker, srverr.ServerError) {
	if t == nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.TwilioToCallWorker").
			SetDetails("empty request")
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

	status, ok := eventStatusTwilioToDomain[t.CallStatus]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.TwilioToCallWorker/InvalidEvent").
			SetDetails(fmt.Sprintf("failed parse event status %v", t.CallStatus))
	}
	direction, ok := directionTwilioToDomain[t.Direction]
	if !ok {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.TwilioToCallWorker/InvalidDirection").
			SetDetails(fmt.Sprintf("failed parse event direction %v", t.Direction))
	}

	timestamp, err := time.Parse(time.RFC1123Z, t.Timestamp)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "entity.TwilioToCallWorker/Parse").
			SetError(err.Error()).SetDetails(fmt.Sprintf("failed parse event timestamp %v", t.Timestamp))
	}
	recDuration, err := strconv.Atoi(t.RecordingDuration)
	if err != nil {
		recDuration = 0
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
