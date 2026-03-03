package middleware

import (
	"context"
	"net/http"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	"telephony/internal/shared/dto"
	"telephony/internal/shared/response"
	srverr "telephony/internal/shared/server_error"
	transperr "telephony/internal/shared/transport_error"
	"telephony/pkg/jwt"
	"telephony/pkg/logger"

	"github.com/gorilla/mux"
)

// AllowedCORSOrigins is a comma-separated list of origins for CORS, or "*" for any.
// Empty means no CORS (browser same-origin only).
// For credentials (cookies) to work, use a specific origin (e.g. "http://localhost:5173"), not "*".
var AllowedCORSOrigins = ""

type Middleware interface {
	PermissionMiddleware(roles ...domain.Role) func(http.Handler) http.Handler
	WithAccess(next http.Handler) http.Handler
	PanicRecovery(next http.Handler) http.Handler
	CORS(next http.Handler) http.Handler
}

type middleware struct {
	log logger.Logger

	cfg *config.Config

	httpResp  response.HttpResponse
	converter transperr.ErrorConverter
}

func NewMiddleware(
	log logger.Logger,
	cfg *config.Config,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
) Middleware {
	return &middleware{
		log:       log,
		cfg:       cfg,
		httpResp:  httpResp,
		converter: converter,
	}
}

type contextKey string

const (
	contextKeyAuth contextKey = "AccountCtxKey"
)

func (m *middleware) WithAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if raw == "" || !strings.HasPrefix(raw, prefix) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(raw, prefix)

		userCtx, err := jwt.ParseAccessToken(tokenString, []byte(m.cfg.Auth.JWTSecret))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), m.cfg.Handler.RequestTimeout)
		defer cancel()

		ctx = context.WithValue(ctx, contextKeyAuth, userCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *middleware) PermissionMiddleware(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acc, ok := r.Context().Value(contextKeyAuth).(*domain.UserFromAccess)
			if !ok {
				h.httpResp.ErrorResponse(w, r,
					dto.TransportErrorToModel(
						h.converter.ToHTTP(srverr.NewServerError(srverr.ErrUnauthorized, "middleware.permissionMiddleware/acc")),
					),
				)
				return
			}
			for _, role := range roles {
				if role == acc.Role {
					next.ServeHTTP(w, r)
					return
				}
			}
			h.httpResp.ErrorResponse(w, r,
				dto.TransportErrorToModel(
					h.converter.ToHTTP(srverr.NewServerError(srverr.ErrForbidden, "middleware.permissionMiddleware/role")),
				),
			)
		})
	}
}

func GetAuthFromContext(r *http.Request) (*domain.UserFromAccess, bool) {
	acc, ok := r.Context().Value(contextKeyAuth).(*domain.UserFromAccess)
	return acc, ok
}

func (m *middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				m.log.Errorf("[%s] Panic recovered: %v", r.URL.String(), rec)
				w.WriteHeader(500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if AllowedCORSOrigins != "" && (AllowedCORSOrigins == "*" || originAllowed(AllowedCORSOrigins, origin)) {
			allowOrigin := AllowedCORSOrigins
			if allowOrigin != "*" {
				allowOrigin = origin
			}
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if allowOrigin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func originAllowed(allowed, origin string) bool {
	if origin == "" {
		return false
	}
	for _, o := range strings.Split(strings.TrimSpace(allowed), ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

// SetupCORS sets AllowedCORSOrigins, applies CORS middleware and registers OPTIONS preflight handler.
// allowedOrigins is from config (e.g. "*" or "http://localhost:5173,https://app.example.com").
// Must be called before registering routes.
func SetupCORS(router *mux.Router, allowedOrigins string, mid Middleware) {
	AllowedCORSOrigins = allowedOrigins
	router.Use(mid.CORS)
	router.Methods(http.MethodOptions).PathPrefix("/").HandlerFunc(corsPreflight)
}

func corsPreflight(w http.ResponseWriter, r *http.Request) {
	if AllowedCORSOrigins == "" {
		return
	}
	allowOrigin := AllowedCORSOrigins
	if o := r.Header.Get("Origin"); o != "" && AllowedCORSOrigins != "*" && originAllowed(AllowedCORSOrigins, o) {
		allowOrigin = o
	}
	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusNoContent)
}
