package mts

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
	ServiceErrorMTSBadRequest srverr.ErrorTypeBadRequest = "mts_bad_request"
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
	webhookRouter.HandleFunc("/mts/voice/status", h.handleMTSVoiceStatus).Methods(http.MethodPost)
}

func (h *webhookHandler) handleMTSVoiceStatus(w http.ResponseWriter, r *http.Request) {
	var req entity.MTSWebhook
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sErr := srverr.NewServerError(ServiceErrorMTSBadRequest, "mts.handleMTSVoiceStatus/decode").SetError(err.Error())
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	worker, convErr := entity.MTSToCallWorker(&req)
	if convErr != nil {
		sErr := srverr.NewServerError(ServiceErrorMTSBadRequest, "mts.handleMTSVoiceStatus/convert").SetError(convErr.Error())
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	sErr := h.call.VoiceStatus(r.Context(), domain.MTS, worker)
	if sErr != nil {
		h.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResponse.WriteResponse(w, r, http.StatusOK, nil)
}
