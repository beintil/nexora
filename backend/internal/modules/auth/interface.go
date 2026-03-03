package auth

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/runner"
	srverr "telephony/internal/shared/server_error"
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
}
