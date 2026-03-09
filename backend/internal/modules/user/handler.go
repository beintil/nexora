package user

import (
	"encoding/json"
	"io"
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

func (h *handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "user.handleGetProfile/acc")),
			),
		)
		return
	}
	profile, sErr := h.service.GetProfile(r.Context(), userCtx.ID.String())
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.ProfileDomainToModel(profile))
}

func (h *handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "user.handleUpdateProfile/acc")),
			),
		)
		return
	}
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleUpdateProfile/decode").SetError(err.Error()),
		)))
		return
	}
	profile, sErr := h.service.UpdateProfile(r.Context(), userCtx.ID.String(), dto.UpdateProfileRequestToDomain(&req))
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.ProfileDomainToModel(profile))
}

const multipartMaxBytes = 6 * 1024 * 1024 // 6 MiB для формы с файлом

func (h *handler) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r,
			dto.TransportErrorToModel(
				h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "user.handleUploadAvatar/acc")),
			),
		)
		return
	}
	if err := r.ParseMultipartForm(multipartMaxBytes); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleUploadAvatar/parse").SetError(err.Error()),
		)))
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleUploadAvatar/missing_avatar").SetError("missing multipart field 'avatar'"),
		)))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleUploadAvatar/read").SetError(err.Error()),
		)))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	profile, sErr := h.service.UploadAvatar(r.Context(), userCtx.ID.String(), data, contentType)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}
	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.ProfileDomainToModel(profile))
}

func (h *handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "user.handleListUsers/acc"))))
		return
	}

	users, sErr := h.service.GetCompanyUsers(r.Context(), userCtx.CompanyID)
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusOK, dto.UsersDomainToModel(users))
}

func (h *handler) handleCreateStaff(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetAuthFromContext(r)
	if !ok {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(srverr.ErrInternalServerError, "user.handleCreateStaff/acc"))))
		return
	}

	var req models.CreateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleCreateStaff/decode").SetError(err.Error()))))
		return
	}

	if err := req.Validate(strfmt.NewFormats()); err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleCreateStaff/validate").SetError(err.Error()))))
		return
	}

	u, sErr := h.service.CreateStaff(r.Context(), userCtx.CompanyID, req.Email.String(), req.FullName, domain.Role(req.RoleID))
	if sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusCreated, dto.UserToProfileModel(u))
}

func (h *handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(srverr.NewServerError(ServiceErrorUserBadRequest, "user.handleDeleteUser/parse_id").SetError(err.Error()))))
		return
	}

	if sErr := h.service.DeleteUser(r.Context(), uid); sErr != nil {
		h.httpResp.ErrorResponse(w, r, dto.TransportErrorToModel(h.converter.ToHTTP(sErr)))
		return
	}

	h.httpResp.WriteResponse(w, r, http.StatusNoContent, nil)
}
