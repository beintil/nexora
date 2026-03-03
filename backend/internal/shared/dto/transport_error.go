package dto

import (
	transperr "telephony/internal/shared/transport_error"
	"telephony/models"
	"telephony/pkg/pointer"

	"github.com/go-openapi/strfmt"
)

func TransportErrorToModel(m transperr.TransportError) *models.TransportError {
	return &models.TransportError{
		Error:         m.Error(),
		Code:          pointer.P(int32(m.GetCode())),
		Message:       pointer.P(m.GetMessage()),
		Details:       m.GetDetails(),
		TransactionID: strfmt.UUID(m.GetTransactionID().String()),
	}
}
