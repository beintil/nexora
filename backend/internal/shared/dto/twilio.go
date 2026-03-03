package dto

import (
	"telephony/internal/domain"
	"telephony/models"
)

// TwilioCallStatusFormToDomain преобразует DTO из Swagger (TwilioVoiceStatusCallbackForm) в доменную модель.
func TwilioCallStatusFormToDomain(d *models.TwilioVoiceStatusCallbackForm) *domain.TwilioCallStatusCallback {
	if d == nil {
		return nil
	}
	return &domain.TwilioCallStatusCallback{
		CallSid:           d.CallSid,
		ParentCallSid:     d.ParentCallSid,
		AccountSid:        d.AccountSid,
		From:              d.From,
		To:                d.To,
		CallStatus:        d.CallStatus,
		Direction:         d.Direction,
		ApiVersion:        d.APIVersion,
		CallerName:        d.CallerName,
		ForwardedFrom:     d.ForwardedFrom,
		CallbackSource:    d.CallbackSource,
		SequenceNumber:    d.SequenceNumber,
		Timestamp:         d.Timestamp,
		CallDuration:      d.CallDuration,
		Duration:          d.Duration,
		SipResponseCode:   d.SipResponseCode,
		RecordingSid:      d.RecordingSid,
		RecordingUrl:      d.RecordingURL,
		RecordingDuration: d.RecordingDuration,

		Called:        d.Called,
		CalledCity:    d.CalledCity,
		CalledCountry: d.CalledCountry,
		CalledState:   d.CalledState,
		CalledZip:     d.CalledZip,

		Caller:        d.Caller,
		CallerCity:    d.CallerCity,
		CallerCountry: d.CallerCountry,
		CallerState:   d.CallerState,
		CallerZip:     d.CallerZip,

		FromCity:    d.FromCity,
		FromCountry: d.FromCountry,
		FromState:   d.FromState,
		FromZip:     d.FromZip,

		ToCity:    d.ToCity,
		ToCountry: d.ToCountry,
		ToState:   d.ToState,
		ToZip:     d.ToZip,
	}
}
