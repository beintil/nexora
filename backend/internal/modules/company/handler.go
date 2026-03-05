package company

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

func (h *handler) handleListCompanyTelephony(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "company.handleListCompanyTelephony/auth")),
			),
		)
		return
	}
	list, sErr := h.service.ListTelephonyByCompanyID(r.Context(), auth.CompanyID)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.CompanyTelephonyListDomainToModel(list))
}

func (h *handler) handleAttachCompanyTelephony(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "company.handleAttachCompanyTelephony/auth")),
			),
		)
		return
	}
	var req models.CompanyTelephonyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleAttachCompanyTelephony/decode").SetError(err.Error())),
			),
		)
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleAttachCompanyTelephony/validate").SetError(err.Error())),
			),
		)
		return
	}
	telephonyName, externalAccountID := dto.CompanyTelephonyCreateRequestToParams(&req)
	if telephonyName == "" {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleAttachCompanyTelephony/empty_telephony_name")),
			),
		)
		return
	}
	if externalAccountID == "" {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleAttachCompanyTelephony/empty_external_account_id")),
			),
		)
		return
	}
	ct, sErr := h.service.AttachTelephonyToCompany(r.Context(), auth.CompanyID, telephonyName, externalAccountID)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusCreated, dto.CompanyTelephonyDomainToModel(ct))
}

func (h *handler) handleDetachCompanyTelephony(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "company.handleDetachCompanyTelephony/auth")),
			),
		)
		return
	}
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleDetachCompanyTelephony/invalid_id")),
			),
		)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.handleDetachCompanyTelephony/invalid_uuid").SetError(err.Error())),
			),
		)
		return
	}
	sErr := h.service.DetachTelephonyFromCompany(r.Context(), auth.CompanyID, id)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) handleListTelephonyDictionary(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "company.handleListTelephonyDictionary/auth")),
			),
		)
		return
	}
	list, sErr := h.service.ListTelephonyDictionary(r.Context())
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.TelephonyDictionaryDomainToModel(list))
}
