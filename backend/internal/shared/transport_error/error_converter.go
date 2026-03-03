package transperr

import (
	"net/http"
	srverr "telephony/internal/shared/server_error"
)

type ErrorConverter interface {
	ToHTTP(err error) TransportError
	// ToGRPC, etc..
}

type errorConverter struct{}

func NewErrorConverter() ErrorConverter {
	return &errorConverter{}
}

func (e *errorConverter) ToHTTP(err error) TransportError {
	var code int

	srvErr, ok := err.(srverr.ServerError)
	if ok {
		switch srvErr.GetServerError().(type) {
		case srverr.ErrorTypeConflict:
			code = http.StatusConflict
		case srverr.ErrorTypeBadRequest:
			code = http.StatusBadRequest
		case srverr.ErrorTypeNotFound:
			code = http.StatusNotFound
		case srverr.ErrorTypeUnauthorized:
			code = http.StatusUnauthorized
		default:
			code = http.StatusInternalServerError
		}
	} else {
		code = http.StatusInternalServerError
	}
	tErr := NewTransportError(err.Error(), code)

	if code != http.StatusInternalServerError {
		_ = tErr.SetMessage(srvErr.GetServerError().String()).SetDetails(srvErr.GetDetails())
	}
	return tErr
}
