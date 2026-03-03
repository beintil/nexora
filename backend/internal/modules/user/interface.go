package user

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service — контракт модуля user. В т.ч. используется в auth для создания и поиска учётной записи (регистрация, логин, refresh).
type Service interface {
	SetUserIsVerifiedWithTx(ctx context.Context, tx pgx.Tx, userID string) srverr.ServerError
	CreateUserWithTx(ctx context.Context, tx pgx.Tx, u *domain.User) srverr.ServerError
	GetUserByEmailWithTx(ctx context.Context, tx pgx.Tx, email string) (*domain.User, srverr.ServerError)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, srverr.ServerError)
	GetUserByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.User, srverr.ServerError)

	GetProfile(ctx context.Context, userID string) (*domain.Profile, srverr.ServerError)
	UpdateProfile(ctx context.Context, userID string, input *domain.UpdateProfileInput) (*domain.Profile, srverr.ServerError)
	UploadAvatar(ctx context.Context, userID string, data []byte, contentType string) (*domain.Profile, srverr.ServerError)
}

type Handler interface {
	handleGetProfile(w http.ResponseWriter, r *http.Request)
	handleUpdateProfile(w http.ResponseWriter, r *http.Request)
	handleUploadAvatar(w http.ResponseWriter, r *http.Request)
}
