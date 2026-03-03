package twilio

import (
	"encoding/json"
	"net/http"
	"telephony/internal/shared/dto"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	srverr "telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"
	"telephony/models"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

const (
	ServiceErrorTwilioBadRequest srverr.ErrorTypeBadRequest = "twilio_bad_request"
)

type webhookHandler struct {
	service Service

	httpResponse response.HttpResponse
	converter    transperr.ErrorConverter

	validationFormat strfmt.Registry
}

func NewHandler(
	service Service,

	httpResponse response.HttpResponse,
	converter transperr.ErrorConverter,

	validationFormat strfmt.Registry,
) Handler {
	return &webhookHandler{
		service: service,

		httpResponse: httpResponse,
		converter:    converter,

		validationFormat: validationFormat,
	}
}

func (m *webhookHandler) Run(router *mux.Router, mid middleware.Middleware) {
	webhookRouter := router.PathPrefix("/webhook").Subrouter()
	webhookRouter.HandleFunc("/twilio/voice/status", m.handleTwilioVoiceStatus).Methods(http.MethodPost)
}

func (m *webhookHandler) handleTwilioVoiceStatus(w http.ResponseWriter, r *http.Request) {
	var req models.TwilioVoiceStatusCallbackForm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sErr := srverr.NewServerError(ServiceErrorTwilioBadRequest, "twilio.handleTwilioVoiceStatus/decode").SetError(err.Error())
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	msg := dto.TwilioCallStatusFormToDomain(&req)
	sErr := m.service.VoiceStatus(r.Context(), msg)
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusOK, nil)
}
