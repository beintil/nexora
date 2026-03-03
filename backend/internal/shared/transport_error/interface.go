package transperr

import "github.com/google/uuid"

type TransportError interface {
	Error() string

	SetMessage(msg string) TransportError
	SetDetails(details string) TransportError

	GetMessage() string
	GetDetails() string

	GetCode() int

	GetTransactionID() uuid.UUID
}
