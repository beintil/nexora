package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	"telephony/internal/shared/dto"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	srverr "telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"
	"telephony/models"
	"telephony/pkg/logger"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type handler struct {
	service Service

	httpResponse response.HttpResponse
	converter    transperr.ErrorConverter
	cfg          config.Config

	log logger.Logger
}

func NewHandler(
	service Service,
	httpResponse response.HttpResponse,
	converter transperr.ErrorConverter,
	cfg config.Config,
	log logger.Logger,
) Handler {
	return &handler{
		service:      service,
		httpResponse: httpResponse,
		converter:    converter,
		cfg:          cfg,
		log:          log,
	}
}

func (m *handler) Run(router *mux.Router, _ middleware.Middleware) {
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", m.handleRegister).Methods(http.MethodPost)
	authRouter.HandleFunc("/login", m.handleLogin).Methods(http.MethodPost)
	authRouter.HandleFunc("/refresh", m.handleRefresh).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", m.handleLogout).Methods(http.MethodPost)
	authRouter.HandleFunc("/verify-link", m.handleVerifyLink).Methods(http.MethodGet)
	authRouter.HandleFunc("/send-code", m.handleSendCode).Methods(http.MethodPost)
}

func (m *handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleRegister/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleRegister/validate").SetError(err.Error()),
		)))
		return
	}

	_, sErr := m.service.Register(r.Context(), dto.RegisterRequestToDomain(&req))
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusCreated, &models.RegisterResponse{})
}

func (m *handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleLogin/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleLogin/validate").SetError(err.Error()),
		)))
		return
	}

	res, sErr := m.service.Login(r.Context(), dto.LoginRequestToDomain(&req))
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.setRefreshTokenCookie(w, res.RefreshToken)
	m.httpResponse.WriteResponse(w, r, http.StatusOK, &models.LoginResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	})
}

func (m *handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleRefresh/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleRefresh/validate").SetError(err.Error()),
		)))
		return
	}

	res, sErr := m.service.Refresh(r.Context(), dto.RefreshRequestToToken(&req))
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.setRefreshTokenCookie(w, res.RefreshToken)
	m.httpResponse.WriteResponse(w, r, http.StatusOK, &models.LoginResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	})
}

func (m *handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	refreshToken := strings.TrimSpace(dto.RefreshRequestToToken(&req))
	if refreshToken == "" {
		if c, err := r.Cookie(m.cfg.Auth.RefreshTokenCookieName); err == nil {
			refreshToken = c.Value
		}
	}
	if refreshToken != "" {
		if sErr := m.service.Logout(r.Context(), refreshToken); sErr != nil {
			m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
			return
		}
	}
	m.clearRefreshTokenCookie(w)
	m.httpResponse.WriteResponse(w, r, http.StatusNoContent, nil)
}

func (m *handler) handleVerifyLink(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorVerifyLink, "auth.handleVerifyLink/empty_token"),
		)))
		return
	}
	if _, err := uuid.Parse(token); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorVerifyLink, "auth.handleVerifyLink/invalid_uuid").SetError(err.Error()),
		)))
		return
	}

	sErr := m.service.VerifyLink(r.Context(), &domain.VerifyLinkInput{Token: token})
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusNoContent, nil)
}

func (m *handler) handleSendCode(w http.ResponseWriter, r *http.Request) {
	var req models.SendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorSendCode, "auth.handleSendCode/decode").SetError(err.Error()),
		)))
		return
	}
	if err := req.Validate(strfmt.NewFormats()); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorSendCode, "auth.handleSendCode/validate").SetError(err.Error()),
		)))
		return
	}
	domainReq := dto.SendCodeRequestToDomain(&req)
	if strings.TrimSpace(domainReq.Email) == "" {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(ServiceErrorSendCode, "auth.handleSendCode/email_required"),
		)))
		return
	}

	sErr := m.service.SendCode(r.Context(), domainReq)
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusNoContent, nil)
}

/* OAuth HTTP handlers are temporarily disabled.

func (m *handler) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	redirectURL, sErr := m.service.StartGoogleOAuth(r.Context())
	if sErr != nil {
		m.redirectOAuthError(w, r, sErr)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (m *handler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	res, sErr := m.service.LoginOrRegisterWithGoogle(r.Context(), code, state)
	if sErr != nil {
		m.redirectOAuthError(w, r, sErr)
		return
	}
	m.setRefreshTokenCookieOAuthCallback(w, res.RefreshToken)
	m.redirectOAuthSuccess(w, r, res.AccessToken, res.RefreshToken)
}

func (m *handler) handleAppleStart(w http.ResponseWriter, r *http.Request) {
	redirectURL, sErr := m.service.StartAppleOAuth(r.Context())
	if sErr != nil {
		m.redirectOAuthError(w, r, sErr)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (m *handler) handleAppleCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		m.redirectOAuthError(w, r, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleAppleCallback/parse").SetError(err.Error()))
		return
	}
	idToken := r.FormValue("id_token")
	state := r.FormValue("state")
	if idToken == "" {
		m.redirectOAuthError(w, r, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.handleAppleCallback/no_id_token"))
		return
	}
	res, sErr := m.service.LoginOrRegisterWithApple(r.Context(), idToken, state)
	if sErr != nil {
		m.redirectOAuthError(w, r, sErr)
		return
	}
	m.setRefreshTokenCookieOAuthCallback(w, res.RefreshToken)
	m.redirectOAuthSuccess(w, r, res.AccessToken, res.RefreshToken)
}

func (m *handler) redirectOAuthSuccess(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	redirectURL := m.cfg.Auth.OAuthFrontendSuccessURL
	if redirectURL == "" {
		redirectURL = "/"
	}
	// Токены в fragment — фронт может прочитать после редиректа (cross-origin). Fragment не уходит на сервер.
	if accessToken != "" && refreshToken != "" {
		redirectURL += "#access_token=" + url.QueryEscape(accessToken) + "&refresh_token=" + url.QueryEscape(refreshToken)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (m *handler) redirectOAuthError(w http.ResponseWriter, r *http.Request, sErr srverr.ServerError) {
	url := m.cfg.Auth.OAuthFrontendErrorURL
	if url == "" {
		url = m.cfg.Auth.OAuthFrontendSuccessURL
	}
	if url == "" {
		url = "/"
	}
	// Не передаём детали ошибки в URL из соображений безопасности; фронт может показать общее сообщение.
	http.Redirect(w, r, url, http.StatusFound)
}

// setRefreshTokenCookieOAuthCallback — как setRefreshTokenCookie, но SameSite=Lax для приёма редиректа с Google/Apple.
func (m *handler) setRefreshTokenCookieOAuthCallback(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   m.cfg.Auth.RefreshTokenTTLSec,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

*/

func (m *handler) setRefreshTokenCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   m.cfg.Auth.RefreshTokenTTLSec,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (m *handler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}
