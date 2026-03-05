package twilio

import (
	"encoding/json"
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/modules/telephony/entity"
	"telephony/internal/modules/telephony/webhook"
	"telephony/internal/shared/dto"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	srverr "telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

const (
	ServiceErrorTwilioBadRequest srverr.ErrorTypeBadRequest = "twilio_bad_request"
)

type webhookHandler struct {
	call webhook.TelephonyCall

	httpResponse response.HttpResponse
	converter    transperr.ErrorConverter

	validationFormat strfmt.Registry
}

func NewHandler(
	call webhook.TelephonyCall,

	httpResponse response.HttpResponse,
	converter transperr.ErrorConverter,

	validationFormat strfmt.Registry,
) Handler {
	return &webhookHandler{
		call: call,

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
	var req entity.TwilioVoiceStatusCallbackForm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sErr := srverr.NewServerError(ServiceErrorTwilioBadRequest, "twilio.handleTwilioVoiceStatus/decode").SetError(err.Error())
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	msg := dto.TwilioCallStatusFormToDomain(&req)
	if msg == nil {
		sErr := srverr.NewServerError(ServiceErrorTwilioBadRequest, "twilio.handleTwilioVoiceStatus/validate").SetError("invalid request")
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}
	worker, err := entity.TwilioToCallWorker(msg)
	if err != nil {
		sErr := srverr.NewServerError(ServiceErrorTwilioBadRequest, "twilio.handleTwilioVoiceStatus/convert").SetError(err.Error())
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	sErr := m.call.VoiceStatus(r.Context(), domain.Twilio, worker)
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusOK, nil)
}
