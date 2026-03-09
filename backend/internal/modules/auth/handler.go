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
	"telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"
	"telephony/models"
	"telephony/pkg/logger"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type handler struct {
	service Service

	httpResponse response.HttpResponse
	converter    transperr.ErrorConverter
	cfg          config.Config

	log logger.Logger

	redisClient *redis.Client
}

func NewHandler(
	service Service,
	httpResponse response.HttpResponse,
	converter transperr.ErrorConverter,
	cfg config.Config,
	log logger.Logger,
	redisClient *redis.Client,
) Handler {
	return &handler{
		service:      service,
		httpResponse: httpResponse,
		converter:    converter,
		cfg:          cfg,
		log:          log,
		redisClient:  redisClient,
	}
}

func (m *handler) Run(router *mux.Router, mid middleware.Middleware) {
	authRouter := router.PathPrefix("/auth").Subrouter()

	rateLimitRegister := middleware.RateLimitHandler(m.redisClient, 5, time.Minute)
	rateLimitLogin := middleware.RateLimitHandler(m.redisClient, 10, time.Minute)
	rateLimitSendCode := middleware.RateLimitHandler(m.redisClient, 3, time.Minute)

	authRouter.Handle("/register", rateLimitRegister(http.HandlerFunc(m.handleRegister))).Methods(http.MethodPost)
	authRouter.Handle("/login", rateLimitLogin(http.HandlerFunc(m.handleLogin))).Methods(http.MethodPost)
	authRouter.HandleFunc("/refresh", m.handleRefresh).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", m.handleLogout).Methods(http.MethodPost)
	authRouter.HandleFunc("/verify-link", m.handleVerifyLink).Methods(http.MethodGet, http.MethodPost)
	authRouter.Handle("/send-code", rateLimitSendCode(http.HandlerFunc(m.handleSendCode))).Methods(http.MethodPost)

	authRouter.HandleFunc("/google/start", m.handleGoogleStart).Methods(http.MethodGet)
	authRouter.HandleFunc("/google/callback", m.handleGoogleCallback).Methods(http.MethodGet)
	authRouter.HandleFunc("/apple/start", m.handleAppleStart).Methods(http.MethodGet)
	authRouter.HandleFunc("/apple/callback", m.handleAppleCallback).Methods(http.MethodPost)

	authRouter.HandleFunc("/forgot-password", m.handleForgotPassword).Methods(http.MethodPost)
	authRouter.HandleFunc("/reset-password", m.handleResetPassword).Methods(http.MethodPost)
	authRouter.Handle("/change-password", mid.WithAccess(http.HandlerFunc(m.handleChangePassword))).Methods(http.MethodPatch)
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

	res, sErr := m.service.Register(r.Context(), dto.RegisterRequestToDomain(&req))
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.setRefreshTokenCookie(w, res.RefreshToken)
	m.httpResponse.WriteResponse(w, r, http.StatusCreated, &models.RegisterResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	})
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

func (m *handler) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	redirectURL, sErr := m.service.StartGoogleOAuth(r.Context())
	if sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
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
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (m *handler) handleAppleCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		m.redirectOAuthError(w, r, srverr.NewServerError(srverr.ErrBadRequest, "auth.handleAppleCallback/parse").SetError(err.Error()))
		return
	}
	idToken := r.FormValue("id_token")
	state := r.FormValue("state")

	res, sErr := m.service.LoginOrRegisterWithApple(r.Context(), idToken, state)
	if sErr != nil {
		m.redirectOAuthError(w, r, sErr)
		return
	}

	m.setRefreshTokenCookieOAuthCallback(w, res.RefreshToken)
	m.redirectOAuthSuccess(w, r, res.AccessToken, res.RefreshToken)
}

func (m *handler) redirectOAuthSuccess(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	redirectURL := m.cfg.Auth.OAuth.OAuthFrontendSuccessURL
	if redirectURL == "" {
		redirectURL = "/"
	}
	if accessToken != "" && refreshToken != "" {
		// Use fragment for tokens so they aren't sent to server on redirect
		redirectURL += "#access_token=" + accessToken + "&refresh_token=" + refreshToken
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (m *handler) redirectOAuthError(w http.ResponseWriter, r *http.Request, sErr srverr.ServerError) {
	url := m.cfg.Auth.OAuth.OAuthFrontendErrorURL
	if url == "" {
		url = "/"
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (m *handler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   m.cfg.Auth.RefreshTokenCookieMaxAge,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

func (m *handler) setRefreshTokenCookieOAuthCallback(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   m.cfg.Auth.RefreshTokenCookieMaxAge,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteLaxMode, // Lax is required for redirects from OAuth providers
	}
	http.SetCookie(w, &cookie)
}

func (m *handler) clearRefreshTokenCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     m.cfg.Auth.RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cfg.Auth.RefreshTokenCookieSecure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

func (m *handler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrBadRequest, "auth.handleForgotPassword/decode").SetError(err.Error()),
		)))
		return
	}

	if sErr := m.service.SendResetPasswordLink(r.Context(), req.Email); sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusAccepted, nil)
}

func (m *handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrBadRequest, "auth.handleResetPassword/decode").SetError(err.Error()),
		)))
		return
	}

	if sErr := m.service.ResetPassword(r.Context(), req.Token, req.NewPassword.String()); sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusNoContent, nil)
}

func (m *handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrBadRequest, "auth.handleChangePassword/decode").SetError(err.Error()),
		)))
		return
	}

	acc, ok := middleware.GetAuthFromContext(r)
	if !ok {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(
			srverr.NewServerError(srverr.ErrUnauthorized, "auth.handleChangePassword/auth_context"),
		)))
		return
	}

	if sErr := m.service.ChangePassword(r.Context(), acc.ID, req.OldPassword.String(), req.NewPassword.String()); sErr != nil {
		m.httpResponse.ErrorResponse(w, r, dto.TransportErrorToModel(m.converter.ToHTTP(sErr)))
		return
	}

	m.httpResponse.WriteResponse(w, r, http.StatusNoContent, nil)
}
