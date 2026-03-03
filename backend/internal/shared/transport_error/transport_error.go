package transperr

import "github.com/google/uuid"

type transportError struct {
	error   string
	message string
	details string

	code          int // HTTP | gRPC status code
	transactionID uuid.UUID
}

func NewTransportError(error string, code int) TransportError {
	return &transportError{
		error:         error,
		code:          code,
		transactionID: uuid.New(),
	}
}

func (t *transportError) Error() string {
	return t.error
}

func (t *transportError) SetMessage(msg string) TransportError {
	t.message = msg
	return t
}

func (t *transportError) SetDetails(details string) TransportError {
	t.details = details
	return t
}

func (t *transportError) GetMessage() string {
	return t.message
}

func (t *transportError) GetDetails() string {
	return t.details
}

func (t *transportError) GetCode() int {
	return t.code
}

func (t *transportError) GetTransactionID() uuid.UUID {
	return t.transactionID
}
