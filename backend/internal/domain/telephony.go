package domain

type Telephony struct {
	ID   int64
	Name TelephonyName
}

type TelephonyName string

const (
	Twilio  TelephonyName = "Twilio"
	Mango   TelephonyName = "Mango"
	Zadarma TelephonyName = "Zadarma"
	MTS     TelephonyName = "MTS"
	Beeline TelephonyName = "Beeline"
)
