package domain

type Telephony struct {
	ID   int64
	Name TelephonyName
}

type TelephonyName string

const (
	Twilio TelephonyName = "Twilio"
)
