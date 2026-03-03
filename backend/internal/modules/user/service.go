package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	"telephony/internal/modules/company"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"
	"telephony/pkg/client/yandexstorage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ServiceErrorUserNotFound      srverr.ErrorTypeNotFound   = "user_not_found"
	ServiceErrorUserAlreadyExists srverr.ErrorTypeConflict   = "user_already_exists"
	ServiceErrorCompanyNotFound   srverr.ErrorTypeNotFound   = "company_not_found"
	ServiceErrorUserBadRequest    srverr.ErrorTypeBadRequest = "user_bad_request"
)

type service struct {
	repo           repository
	companyService company.Service
	transaction    postgres.Transaction
	s3Storage      yandexstorage.Client
	avatarPrefix   string
}

func NewService(
	repo repository,
	companyService company.Service,
	transaction postgres.Transaction,
	s3Storage yandexstorage.Client,
	avatarPrefix string,
) Service {
	return &service{
		repo:           repo,
		companyService: companyService,
		transaction:    transaction,
		s3Storage:      s3Storage,
		avatarPrefix:   avatarPrefix,
	}
}

func (s *service) CreateUserWithTx(ctx context.Context, tx pgx.Tx, u *domain.User) srverr.ServerError {
	if u == nil {
		return srverr.NewServerError(ServiceErrorUserBadRequest, "user.CreateUserWithTx/nil_user")
	}
	if u.CompanyID == uuid.Nil {
		return srverr.NewServerError(ServiceErrorUserBadRequest, "user.CreateUserWithTx/empty_company_id")
	}
	if u.PasswordHash == "" {
		return srverr.NewServerError(ServiceErrorUserBadRequest, "user.CreateUserWithTx/empty_password_hash")
	}
	if u.Email == nil || strings.TrimSpace(*u.Email) == "" {
		return srverr.NewServerError(ServiceErrorUserBadRequest, "user.CreateUserWithTx/email_required")
	}
	existing, err := s.repo.getUserByEmail(ctx, tx, *u.Email)
	if err != nil && !errors.Is(err, errRepoUserNotFound) {
		return srverr.NewServerError(srverr.ErrInternalServerError, "user.CreateUserWithTx/check_email").SetError(err.Error())
	}
	if existing != nil {
		return srverr.NewServerError(ServiceErrorUserAlreadyExists, "user.CreateUserWithTx/email_exists")
	}
	if err := s.repo.createUser(ctx, tx, u); err != nil {
		if errors.Is(err, errRepoUserDuplicate) {
			return srverr.NewServerError(ServiceErrorUserAlreadyExists, "user.CreateUserWithTx/duplicate")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "user.CreateUserWithTx/create").SetError(err.Error())
	}
	return nil
}

func (s *service) GetUserByEmailWithTx(ctx context.Context, tx pgx.Tx, email string) (*domain.User, srverr.ServerError) {
	u, err := s.repo.getUserByEmail(ctx, tx, email)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.GetUserByEmailWithTx/not_found")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.GetUserByEmailWithTx/repo").SetError(err.Error())
	}
	return u, nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, srverr.ServerError) {
	u, err := s.repo.getUserByID(ctx, nil, id)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.GetUserByID/not_found")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.GetUserByID/repo").SetError(err.Error())
	}
	return u, nil
}

func (s *service) GetUserByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.User, srverr.ServerError) {
	u, err := s.repo.getUserByID(ctx, tx, id)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.GetUserByIDWithTx/not_found")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.GetUserByIDWithTx/repo").SetError(err.Error())
	}
	return u, nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*domain.Profile, srverr.ServerError) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.GetProfileWithTx/parse_user_id").SetError(err.Error())
	}
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.GetProfileWithTx/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	u, err := s.repo.getUserByID(ctx, tx, uid)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.GetProfile/get_user")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.GetProfile/user_repo").SetError(err.Error())
	}
	c, sErr := s.companyService.GetCompanyByIDWithTx(ctx, tx, u.CompanyID)
	if sErr != nil {
		if sErr.GetServerError() == company.ServiceErrorCompanyNotFound {
			return nil, srverr.NewServerError(ServiceErrorCompanyNotFound, "user.GetProfile/company")
		}
		return nil, sErr
	}
	return &domain.Profile{
		ID:          u.ID,
		CompanyID:   u.CompanyID,
		CompanyName: c.Name,
		RoleID:      u.RoleID,
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		AvatarID:    u.AvatarID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, input *domain.UpdateProfileInput) (*domain.Profile, srverr.ServerError) {
	if input == nil {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UpdateProfileWithTx/nil_input")
	}
	if input.FullName != nil && len(*input.FullName) > config.UserMaxFullNameLength {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UpdateProfileWithTx/full_name_too_long")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UpdateProfileWithTx/parse_user_id").SetError(err.Error())
	}
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UpdateProfileWithTx/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	if err := s.repo.updateUserProfile(ctx, tx, uid, input.FullName, nil, nil); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UpdateProfile/user_repo").SetError(err.Error())
	}
	u, err := s.repo.getUserByID(ctx, tx, uid)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.UpdateProfile/get_user")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UpdateProfile/user_repo").SetError(err.Error())
	}
	c, sErr := s.companyService.GetCompanyByIDWithTx(ctx, tx, u.CompanyID)
	if sErr != nil {
		if sErr.GetServerError() == company.ServiceErrorCompanyNotFound {
			return nil, srverr.NewServerError(ServiceErrorCompanyNotFound, "user.UpdateProfile/company")
		}
		return nil, sErr
	}
	if err := s.transaction.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UpdateProfileWithTx/commit").SetError(err.Error())
	}
	return &domain.Profile{
		ID:          u.ID,
		CompanyID:   u.CompanyID,
		CompanyName: c.Name,
		RoleID:      u.RoleID,
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		AvatarID:    u.AvatarID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (s *service) UploadAvatar(ctx context.Context, userID string, data []byte, contentType string) (*domain.Profile, srverr.ServerError) {
	ext, ok := config.UserAllowedAvatarContentTypes[contentType]
	if !ok {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UploadAvatarWithTx/unsupported_type").SetError("allowed: image/jpeg, image/png, image/webp")
	}
	if len(data) > config.UserMaxAvatarSize {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UploadAvatarWithTx/too_large").SetError("max 5 MiB")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorUserBadRequest, "user.UploadAvatarWithTx/parse_user_id").SetError(err.Error())
	}
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatarWithTx/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	u, err := s.repo.getUserByID(ctx, tx, uid)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.UploadAvatar/get_user")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatar/user_repo").SetError(err.Error())
	}
	if u.AvatarID != nil {
		err = s.s3Storage.Delete(ctx, fmt.Sprintf("%s/%s", s.avatarPrefix, *u.AvatarID))
		if err != nil {
			return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatar/delete_avatar").SetError(err.Error())
		}
	}

	avatarID := uuid.New().String() + ext
	objectKey := fmt.Sprintf("%s/%s", s.avatarPrefix, avatarID)
	publicURL, err := s.s3Storage.Upload(ctx, objectKey, data, contentType)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatarWithTx/upload").SetError(err.Error())
	}
	if err := s.repo.updateUserProfile(ctx, tx, uid, nil, &publicURL, &avatarID); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatar/user_repo").SetError(err.Error())
	}
	u, err = s.repo.getUserByID(ctx, tx, uid)
	if err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return nil, srverr.NewServerError(ServiceErrorUserNotFound, "user.UploadAvatar/get_user")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatar/user_repo").SetError(err.Error())
	}
	c, sErr := s.companyService.GetCompanyByIDWithTx(ctx, tx, u.CompanyID)
	if sErr != nil {
		if sErr.GetServerError() == company.ServiceErrorCompanyNotFound {
			return nil, srverr.NewServerError(ServiceErrorCompanyNotFound, "user.UploadAvatar/company")
		}
		return nil, sErr
	}
	if err := s.transaction.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "user.UploadAvatarWithTx/commit").SetError(err.Error())
	}
	return &domain.Profile{
		ID:          u.ID,
		CompanyID:   u.CompanyID,
		CompanyName: c.Name,
		RoleID:      u.RoleID,
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		AvatarID:    u.AvatarID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (s *service) SetUserIsVerifiedWithTx(ctx context.Context, tx pgx.Tx, userID string) srverr.ServerError {
	id, err := uuid.Parse(userID)
	if err != nil {
		return srverr.NewServerError(ServiceErrorUserBadRequest, "user.SetUserIsVerifiedWithTx/parse_user_id").SetError(err.Error())
	}
	if err := s.repo.setUserVerified(ctx, tx, id); err != nil {
		if errors.Is(err, errRepoUserNotFound) {
			return srverr.NewServerError(ServiceErrorUserNotFound, "user.SetUserIsVerifiedWithTx/not_found")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "user.SetUserIsVerifiedWithTx/repo").SetError(err.Error())
	}
	return nil
}
