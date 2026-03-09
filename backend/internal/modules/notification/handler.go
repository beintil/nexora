package notification

import (
	"net/http"
	"telephony/internal/shared/dto"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	srverr "telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type handler struct {
	service   Service
	httpResp  response.HttpResponse
	converter transperr.ErrorConverter
}

func NewHandler(service Service, httpResp response.HttpResponse, converter transperr.ErrorConverter) Handler {
	return &handler{
		service:   service,
		httpResp:  httpResp,
		converter: converter,
	}
}

func (h *handler) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "notification.handleGetNotifications/acc"))))
		return
	}

	nn, sErr := h.service.GetNotifications(r.Context(), userCtx.ID)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.NotificationsDomainToModel(nn))
}

func (h *handler) handleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "notification.handleMarkAsRead/parse_id").SetError(err.Error()))))
		return
	}

	if sErr := h.service.MarkAsRead(r.Context(), id); sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusNoContent, nil)
}
