package auth

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/runner"
	"telephony/internal/shared/server_error"
	"time"

	"github.com/google/uuid"
)

type Handler interface {
	handleRegister(w http.ResponseWriter, r *http.Request)
	handleLogin(w http.ResponseWriter, r *http.Request)
	handleRefresh(w http.ResponseWriter, r *http.Request)
	handleLogout(w http.ResponseWriter, r *http.Request)
	handleVerifyLink(w http.ResponseWriter, r *http.Request)
	handleSendCode(w http.ResponseWriter, r *http.Request)

	runner.Runner
}

type Service interface {
	Register(ctx context.Context, req *domain.AuthRegisterInput) (*domain.AuthTokens, srverr.ServerError)
	Login(ctx context.Context, req *domain.AuthLoginInput) (*domain.AuthTokens, srverr.ServerError)
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, srverr.ServerError)
	Logout(ctx context.Context, refreshToken string) srverr.ServerError
	VerifyLink(ctx context.Context, req *domain.VerifyLinkInput) srverr.ServerError
	SendCode(ctx context.Context, req *domain.SendCodeInput) srverr.ServerError

	StartGoogleOAuth(ctx context.Context) (string, srverr.ServerError)
	LoginOrRegisterWithGoogle(ctx context.Context, code, state string) (*domain.AuthTokens, srverr.ServerError)
	StartAppleOAuth(ctx context.Context) (string, srverr.ServerError)
	LoginOrRegisterWithApple(ctx context.Context, idToken, state string) (*domain.AuthTokens, srverr.ServerError)

	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) srverr.ServerError
	SendResetPasswordLink(ctx context.Context, email string) srverr.ServerError
	ResetPassword(ctx context.Context, token, newPassword string) srverr.ServerError
}

type repository interface {
	setVerifyEmailLink(ctx context.Context, token, email string, ttl time.Duration) error
	getVerifyEmailLink(ctx context.Context, token string) (string, error)
	deleteVerifyEmailLink(ctx context.Context, token string) error

	setTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) error
	getUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error)
	deleteTokensByUserID(ctx context.Context, userID uuid.UUID) error
	deleteTokensByRefreshToken(ctx context.Context, refreshToken string) error

	setOAuthState(ctx context.Context, state string, ttl time.Duration) error
	getOAuthState(ctx context.Context, state string) (string, error)
	deleteOAuthState(ctx context.Context, state string) error

	setPasswordResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	getPasswordResetToken(ctx context.Context, token string) (string, error)
	deletePasswordResetToken(ctx context.Context, token string) error
}
