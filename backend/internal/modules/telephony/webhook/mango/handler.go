package mango

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

	"github.com/gorilla/mux"
)

const (
	ServiceErrorMangoBadRequest srverr.ErrorTypeBadRequest = "mango_bad_request"
)

type webhookHandler struct {
	call webhook.TelephonyCall

	httpResponse response.HttpResponse
	converter    transperr.ErrorConverter
}

func NewHandler(
	call webhook.TelephonyCall,

	httpResponse response.HttpResponse,
	converter transperr.ErrorConverter,
) Handler {
	return &webhookHandler{
		call: call,

		httpResponse: httpResponse,
		converter:    converter,
	}
}

func (h *webhookHandler) Run(router *mux.Router, mid middleware.Middleware) {
	webhookRouter := router.PathPrefix("/webhook").Subrouter()
	webhookRouter.HandleFunc("/mango/voice/status", h.handleMangoVoiceStatus).Methods(http.MethodPost)
}

func (h *webhookHandler) handleMangoVoiceStatus(w http.ResponseWriter, r *http.Request) {
	var req entity.MangoWebhook
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sErr := srverr.NewServerError(ServiceErrorMangoBadRequest, "mango.handleMangoVoiceStatus/decode").SetError(err.Error())
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	worker, convErr := entity.MangoToCallWorker(&req)
	if convErr != nil {
		sErr := srverr.NewServerError(ServiceErrorMangoBadRequest, "mango.handleMangoVoiceStatus/convert").SetError(convErr.Error())
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	sErr := h.call.VoiceStatus(r.Context(), domain.Mango, worker)
	if sErr != nil {
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResponse.WriteResponse(w, r, http.StatusOK, nil)
}
