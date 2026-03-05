package call

import (
	"encoding/json"
	"net/http"
	"strings"
	"telephony/internal/domain"
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

func pointerInt32(v int32) *int32 {
	return &v
}

type handler struct {
	service   Service
	httpResp  response.HttpResponse
	converter transperr.ErrorConverter
}

func NewHandler(
	service Service,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
) Handler {
	return &handler{
		service:   service,
		httpResp:  httpResp,
		converter: converter,
	}
}

func (r *handler) Run(router *mux.Router, mid middleware.Middleware) {
	router.Handle("/calls",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner, domain.RoleManager)(http.HandlerFunc(r.handleListCalls)),
		),
	).Methods(http.MethodPost)

	router.Handle("/calls/{id}",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner, domain.RoleManager)(http.HandlerFunc(r.handleGetCallByID)),
		),
	).Methods(http.MethodGet)

	router.Handle("/metrics/calls",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner, domain.RoleManager)(http.HandlerFunc(r.handleGetCallMetrics)),
		),
	).Methods(http.MethodPost)
}

func (h *handler) handleListCalls(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "call.handleListCalls/auth")),
			),
		)
		return
	}

	var req models.CallsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorCallIsNotValid, "call.handleListCalls/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorCallIsNotValid, "call.handleListCalls/validate").SetError(err.Error()),
		)))
		return
	}

	filters := dto.CallsListRequestToFilters(auth.CompanyID, &req)
	page, sErr := h.service.ListCompanyCalls(r.Context(), filters)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	if page == nil {
		h.httpResp.WriteResponse(w, r, http.StatusOK, &models.CallsListResponse{
			Items: []*models.CallsListItem{},
			Meta:  &models.PaginationMeta{Limit: pointerInt32(0), Offset: pointerInt32(0), Page: pointerInt32(1), Total: pointerInt32(0)},
		})
		return
	}

	items := dto.MapSlice(page.Items, dto.CallSummaryDomainToModel)
	limit32 := int32(page.Meta.Limit)
	offset32 := int32(page.Meta.Offset)
	page32 := int32(page.Meta.Page)
	total32 := int32(page.Meta.Total)
	h.httpResp.WriteResponse(w, r, http.StatusOK, &models.CallsListResponse{
		Items: items,
		Meta: &models.PaginationMeta{
			Limit:  &limit32,
			Offset: &offset32,
			Page:   &page32,
			Total:  &total32,
		},
	})
}

func (h *handler) handleGetCallByID(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "call.handleGetCallByID/auth")),
			),
		)
		return
	}

	vars := mux.Vars(r)
	idStr := strings.TrimSpace(vars["id"])
	callID, err := uuid.Parse(idStr)
	if err != nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(ServiceErrorCallIsNotValid, "call.handleGetCallByID/invalid_uuid").SetError(err.Error())),
			),
		)
		return
	}

	tree, sErr := h.service.GetCallTreeByCallUUIDByCompanyUUID(r.Context(), auth.CompanyID, callID)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.CallTreeDomainToModel(tree))
}

func (h *handler) handleGetCallMetrics(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthFromContext(r)
	if !ok || auth == nil || auth.CompanyID == uuid.Nil {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "call.handleGetCallMetrics/auth")),
			),
		)
		return
	}

	var req models.CallMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrInternalServerError, "call.handleGetCallMetrics/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrInternalServerError, "call.handleGetCallMetrics/validate").SetError(err.Error()),
		)))
		return
	}
	if req.DateFrom == nil || req.DateTo == nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrorTypeBadRequest("date_from and date_to are required"), "call.handleGetCallMetrics/required_dates"),
		)))
		return
	}

	from, to := dto.CallMetricsRequestToFromTo(&req)
	if from.After(to) {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrorTypeBadRequest("date_from must be before or equal to date_to"), "call.handleGetCallMetrics/from_after_to")),
			),
		)
		return
	}

	summary, timeseries, sErr := h.service.GetCompanyCallMetrics(r.Context(), auth.CompanyID, from, to)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.CallMetricsDomainToModel(summary, timeseries))
}
