package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	"telephony/internal/modules/company"
	usermod "telephony/internal/modules/user"
	"telephony/internal/shared/cache"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"
	"telephony/internal/shared/templates"
	"telephony/pkg/client/email_sender"
	"telephony/pkg/jwt"
	"telephony/pkg/password"
	"telephony/pkg/validator"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	userSvc     usermod.Service
	companySvc  company.Service
	transaction postgres.Transaction
	store       cache.Cache
	cfg         config.Config

	emailSender email_sender.Sender
}

func NewService(
	userSvc usermod.Service,
	companySvc company.Service,
	transaction postgres.Transaction,
	store cache.Cache,
	cfg config.Config,

	emailSender email_sender.Sender,
) Service {
	return &service{
		userSvc:     userSvc,
		companySvc:  companySvc,
		transaction: transaction,
		store:       store,
		cfg:         cfg,

		emailSender: emailSender,
	}
}

const (
	authRefreshTokenKey         = "auth:refresh_token:"
	authAccessTokenKey          = "auth:access_token:"
	authRefreshTokenByUserIDKey = "auth:refresh_token_by_user_id:"
	authAccessTokenByUserIDKey  = "auth:access_token_by_user_id:"

	authConfirmEmailKey = "auth:link:"
)

const (
	ServiceErrorAuthInvalidBrainPrivilegedUser srverr.ErrorTypeUnauthorized = "Долбаеб, который заходит с пользователя повышенных прав на сторону клиентского сайта, пиздуй в админку. Еще один запрос и вьебу у тебя права кретин, все логируется"

	ServiceErrorAuthRequestIsNotValid  srverr.ErrorTypeBadRequest   = "auth_request_is_not_valid"
	ServiceErrorAuthInvalidCompanyName srverr.ErrorTypeBadRequest   = "Company name is invalid. Must be at least 5 characters long"
	ServiceErrorAuthInvalidPassword    srverr.ErrorTypeBadRequest   = "Password is invalid. Must be at least 8 characters long, contain at least one uppercase letter, one lowercase letter, one number, and one special character."
	ServiceErrorAuthInvalidEmail       srverr.ErrorTypeBadRequest   = "Email is invalid"
	ServiceErrorAuthInvalidCredentials srverr.ErrorTypeUnauthorized = "auth_invalid_credentials"
	ServiceErrorAuthInvalidRefresh     srverr.ErrorTypeUnauthorized = "auth_invalid_refresh"
	ServiceErrorAuthConfirmEmail       srverr.ErrorTypeBadRequest   = "Your account is registered, but you have not confirmed your email address. Please check your email for a confirmation link."
)

const (
	ServiceErrorSendCode   srverr.ErrorTypeBadRequest = "verification_send_code_invalid"
	ServiceErrorVerifyLink srverr.ErrorTypeBadRequest = "verification_verify_link_invalid"
)

func (s *service) Register(ctx context.Context, req *domain.AuthRegisterInput) (*domain.AuthTokens, srverr.ServerError) {
	if req == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.Register/nil_request")
	}
	emailNorm, err := validator.ValidateEmail(req.Email)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidEmail, "auth.Register/email").SetError(err.Error())
	}
	err = validator.ValidatePassword(req.Password)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidPassword, "auth.Register/password").SetError(err.Error())
	}

	companyName := strings.TrimSpace(req.CompanyName)
	if companyName == "" || len(companyName) > 100 || len(companyName) < 5 {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidCompanyName, "auth.Register/empty_company")
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.Register/begin").SetError(err.Error())
	}
	defer s.transaction.Rollback(ctx, tx)

	{
		u, serverErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, emailNorm)
		if serverErr != nil && serverErr.GetServerError() != usermod.ServiceErrorUserNotFound {
			return nil, serverErr
		}
		if u != nil {
			if u.VerifiedRegistration {
				return nil, srverr.NewServerError(ServiceErrorAuthInvalidCredentials, "auth.Register/user_already_verified")
			} else {
				serverErr = s.sendCodeWithTx(ctx, tx, &domain.SendCodeInput{Email: emailNorm})
				if serverErr != nil {
					return nil, serverErr
				}
				return nil, srverr.NewServerError(ServiceErrorAuthConfirmEmail, "auth.Register/user_already_registered")
			}
		}
	}

	companyEntity, serverErr := s.companySvc.CreateCompanyWithTx(ctx, tx, companyName)
	if serverErr != nil {
		return nil, serverErr
	}

	passwordHash, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.RegisterWithTx/hash").SetError(err.Error())
	}

	user := &domain.User{
		ID:           uuid.New(),
		CompanyID:    companyEntity.ID,
		RoleID:       domain.RoleOwner,
		PasswordHash: passwordHash,
	}
	user.Email = &emailNorm
	if fullName := strings.TrimSpace(req.FullName); fullName != "" {
		user.FullName = &fullName
	}
	if user.FullName == nil || strings.TrimSpace(*user.FullName) == "" {
		user.FullName = user.Email
	}
	serverErr = s.userSvc.CreateUserWithTx(ctx, tx, user)
	if serverErr != nil {
		return nil, serverErr
	}
	serverErr = s.sendCodeWithTx(ctx, tx, &domain.SendCodeInput{Email: emailNorm})
	if serverErr != nil {
		return nil, serverErr
	}

	if err = s.transaction.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.RegisterWithTx/commit").SetError(err.Error())
	}

	return s.issueTokens(ctx, user)
}

func (s *service) VerifyLink(ctx context.Context, req *domain.VerifyLinkInput) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLinkWithTx/nil")
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLinkWithTx/empty_token")
	}
	if _, err := uuid.Parse(token); err != nil {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/invalid_uuid").SetError(err.Error())
	}
	var email string
	err := s.store.Get(ctx, authConfirmEmailKey+token, &email)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/link_expired_or_invalid")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/link_store").SetError(err.Error())
	}
	err = s.store.Delete(ctx, authConfirmEmailKey+token)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/link_store").SetError(err.Error())
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/begin").SetError(err.Error())
	}
	defer s.transaction.Rollback(ctx, tx)

	u, serverErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, email)
	if serverErr != nil {
		if serverErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/user_not_found")
		}
		return serverErr
	}
	if u.VerifiedRegistration {
		return nil
	}

	servErr := s.userSvc.SetUserIsVerifiedWithTx(ctx, tx, u.ID.String())
	if servErr != nil {
		if servErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/user_not_found")
		}
		return servErr
	}

	if err = s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/commit").SetError(err.Error())
	}
	return nil
}

func (s *service) SendCode(ctx context.Context, req *domain.SendCodeInput) srverr.ServerError {
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendCodeWithTx/begin").SetError(err.Error())
	}
	defer s.transaction.Rollback(ctx, tx)

	return s.sendCodeWithTx(ctx, tx, req)
}

func (s *service) sendCodeWithTx(ctx context.Context, tx pgx.Tx, req *domain.SendCodeInput) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(ServiceErrorSendCode, "auth.SendCodeWithTx/nil")
	}
	if strings.TrimSpace(req.Email) == "" {
		return srverr.NewServerError(ServiceErrorSendCode, "auth.SendCode/email_required")
	}
	emailNorm, err := validator.ValidateEmail(req.Email)
	if err != nil {
		return srverr.NewServerError(ServiceErrorSendCode, "auth.SendCode/email").SetError(err.Error())
	}
	user, serverErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, emailNorm)
	if serverErr != nil {
		if serverErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			return srverr.NewServerError(ServiceErrorSendCode, "auth.SendCode/user_not_found")
		}
		return serverErr
	}
	if user.VerifiedRegistration {
		return srverr.NewServerError(ServiceErrorSendCode, "auth.SendCode/user_already_verified")
	}

	tokenUUID := uuid.New().String()
	authLinkTTL := time.Duration(s.cfg.Auth.AuthLinkTTLSec) * time.Second
	if err := s.store.Set(ctx, authConfirmEmailKey+tokenUUID, emailNorm, authLinkTTL); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendCode/create_link").SetError(err.Error())
	}

	link := s.cfg.Auth.AuthLinkBaseURL + tokenUUID + "&email=" + url.QueryEscape(emailNorm)
	render, err := templates.NewRenderer()
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/renderer").SetError(err.Error())
	}
	body, err := render.Render(templates.HTMLFileRegister, map[string]interface{}{
		"Link": link,
	})
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/render").SetError(err.Error())
	}

	if emailNorm != "" {
		if err := s.emailSender.Send(ctx, email_sender.Message{
			To:      emailNorm,
			Subject: "Регистрация Аккаунта",
			Text:    fmt.Sprintf("Перейдите по ссылке для подтверждения: %s", link),
			HTML:    body.String(),
		}); err != nil {
			return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/email").SetError(err.Error())
		}
	}
	err = s.store.Set(ctx, authConfirmEmailKey+tokenUUID, req.Email, authLinkTTL)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/create_link").SetError(err.Error())
	}
	return nil
}

func (s *service) Login(ctx context.Context, req *domain.AuthLoginInput) (*domain.AuthTokens, srverr.ServerError) {
	if req == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.Login/nil_request")
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.Login/empty_email")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.Login/empty_password")
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.LoginWithTx/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	var user *domain.User
	emailNorm, emailErr := validator.ValidateEmail(email)
	if emailErr != nil {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidEmail, "auth.Login/invalid_email").SetError(emailErr.Error())
	}
	user, sErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, emailNorm)
	if sErr != nil {
		if sErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			return nil, srverr.NewServerError(ServiceErrorAuthInvalidCredentials, "auth.LoginWithTx/not_found")
		}
		return nil, sErr
	}

	if !password.ComparePassword(user.PasswordHash, req.Password) {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidCredentials, "auth.LoginWithTx/bad_password")
	}

	if user.RoleID.IsPrivileged() {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidBrainPrivilegedUser, "auth.LoginWithTx/privileged_user")
	}

	return s.issueTokens(ctx, user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, srverr.ServerError) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.Refresh/empty_token")
	}
	secret := []byte(s.cfg.Auth.JWTSecret)
	userIDFromJWT, parseErr := jwt.ParseRefreshToken(refreshToken, secret)
	if parseErr != nil {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.RefreshWithTx/invalid_or_expired")
	}
	var userIDStr string
	err := s.store.Get(ctx, authRefreshTokenKey+refreshToken, &userIDStr)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.RefreshWithTx/invalid_or_expired")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.RefreshWithTx/token_store").SetError(err.Error())
	}
	if userIDStr != userIDFromJWT {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.RefreshWithTx/token_mismatch")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.RefreshWithTx/invalid_or_expired")
	}
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.RefreshWithTx/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	u, sErr := s.userSvc.GetUserByIDWithTx(ctx, tx, userID)
	if sErr != nil {
		if sErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			return nil, srverr.NewServerError(ServiceErrorAuthInvalidRefresh, "auth.RefreshWithTx/user_not_found")
		}
		return nil, sErr
	}
	return s.issueTokens(ctx, u)
}

func (s *service) Logout(ctx context.Context, refreshToken string) srverr.ServerError {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	var userIDStr string
	err := s.store.Get(ctx, authRefreshTokenKey+refreshToken, &userIDStr)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return nil
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.Logout/store").SetError(err.Error())
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil
	}
	if delErr := s.DeleteTokensByUserID(ctx, userID); delErr != nil {
		if se, ok := delErr.(srverr.ServerError); ok {
			return se
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.Logout/delete").SetError(delErr.Error())
	}
	return nil
}

func (s *service) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	// Удаляем access
	var accessTokenFromStore string
	err := s.store.Get(ctx, authAccessTokenByUserIDKey+userID.String(), &accessTokenFromStore)
	if err != nil && !errors.Is(err, cache.ErrorCacheValueNotFound) {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/access_store").SetError(err.Error())
	}
	err = s.store.Delete(ctx, authAccessTokenKey+accessTokenFromStore)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/access_store").SetError(err.Error())
	}
	err = s.store.Delete(ctx, authAccessTokenByUserIDKey+userID.String())
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/access_store").SetError(err.Error())
	}

	// Удаляем refresh
	var refreshTokenFromStore string
	err = s.store.Get(ctx, authRefreshTokenByUserIDKey+userID.String(), &refreshTokenFromStore)
	if err != nil && !errors.Is(err, cache.ErrorCacheValueNotFound) {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/refresh_store").SetError(err.Error())
	}
	err = s.store.Delete(ctx, authRefreshTokenKey+refreshTokenFromStore)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/refresh_store").SetError(err.Error())
	}
	err = s.store.Delete(ctx, authRefreshTokenByUserIDKey+userID.String())
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.DeleteTokensByUserID/refresh_store").SetError(err.Error())
	}
	return nil
}

func (s *service) issueTokens(ctx context.Context, user *domain.User) (*domain.AuthTokens, srverr.ServerError) {
	// Удаляем существующие токены если существуют
	err := s.DeleteTokensByUserID(ctx, user.ID)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/delete_tokens").SetError(err.Error())
	}

	accessTTL := time.Duration(s.cfg.Auth.AccessTokenTTLSec) * time.Second
	refreshTTL := time.Duration(s.cfg.Auth.RefreshTokenTTLSec) * time.Second
	accessToken, err := jwt.BuildAccessToken(user.ID, user.CompanyID, user.RoleID, []byte(s.cfg.Auth.JWTSecret), accessTTL)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/access").SetError(err.Error())
	}
	refreshToken, err := jwt.BuildRefreshToken(user.ID.String(), []byte(s.cfg.Auth.JWTSecret), refreshTTL)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/refresh_build").SetError(err.Error())
	}

	// Сохраняем по ключу с токеном
	if err := s.store.Set(ctx, authRefreshTokenKey+refreshToken, user.ID.String(), refreshTTL); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/refresh_store").SetError(err.Error())
	}
	if err := s.store.Set(ctx, authAccessTokenKey+accessToken, user.ID.String(), accessTTL); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/access_store").SetError(err.Error())
	}

	// Сохраняем по ключу с user_id
	if err := s.store.Set(ctx, authAccessTokenByUserIDKey+user.ID.String(), accessToken, accessTTL); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/access_store").SetError(err.Error())
	}
	if err := s.store.Set(ctx, authRefreshTokenByUserIDKey+user.ID.String(), refreshToken, refreshTTL); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/refresh_store").SetError(err.Error())
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

/* OAuth-related methods are temporarily disabled.

func (s *service) LoginOrRegisterWithGoogle(ctx context.Context, code, state string) (*domain.AuthTokens, srverr.ServerError) {
	if s.oauthState == nil || s.googleClient == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.LoginOrRegisterWithGoogle/not_configured")
	}
	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if state == "" || code == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.LoginOrRegisterWithGoogle/empty_code_or_state")
	}
	if err := s.oauthState.validateAndConsumeState(ctx, state); err != nil {
		if errors.Is(err, errStateNotFoundOrExpired) {
			return nil, srverr.NewServerError(ServiceErrorAuthOAuthInvalidState, "auth.LoginOrRegisterWithGoogle/invalid_state")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.LoginOrRegisterWithGoogle/state").SetError(err.Error())
	}
	email, name, _, err := s.googleClient.ExchangeAndGetUserInfo(ctx, code)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.LoginOrRegisterWithGoogle/exchange").SetError(err.Error())
	}
	if email == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNoEmail, "auth.LoginOrRegisterWithGoogle/no_email")
	}
	return s.findOrCreateOAuthUser(ctx, email, name)
}

func (s *service) LoginOrRegisterWithApple(ctx context.Context, idToken, state string) (*domain.AuthTokens, srverr.ServerError) {
	if s.oauthState == nil || s.appleClient == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.LoginOrRegisterWithApple/not_configured")
	}
	idToken = strings.TrimSpace(idToken)
	state = strings.TrimSpace(state)
	if state == "" || idToken == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.LoginOrRegisterWithApple/empty_id_token_or_state")
	}
	if err := s.oauthState.validateAndConsumeState(ctx, state); err != nil {
		if errors.Is(err, errStateNotFoundOrExpired) {
			return nil, srverr.NewServerError(ServiceErrorAuthOAuthInvalidState, "auth.LoginOrRegisterWithApple/invalid_state")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.LoginOrRegisterWithApple/state").SetError(err.Error())
	}
	email, _, err := s.appleClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.LoginOrRegisterWithApple/verify").SetError(err.Error())
	}
	if email == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNoEmail, "auth.LoginOrRegisterWithApple/no_email")
	}
	return s.findOrCreateOAuthUser(ctx, email, "")
}

// findOrCreateOAuthUser находит пользователя по email или создаёт компанию и пользователя (OAuth: пароль случайный).
func (s *service) findOrCreateOAuthUser(ctx context.Context, email, fullName string) (*domain.AuthTokens, srverr.ServerError) {
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	user, sErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, email)
	if sErr == nil && user != nil {
		return s.issueTokens(ctx, user)
	}
	if sErr != nil && sErr.GetServerError() != usermod.ServiceErrorUserNotFound {
		return nil, sErr
	}

	companyName := "Personal"
	companyEntity, serverErr := s.companySvc.CreateCompanyWithTx(ctx, tx, companyName)
	if serverErr != nil {
		return nil, serverErr
	}

	randomPass, err := randomPasswordForOAuth()
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/random_password").SetError(err.Error())
	}
	passwordHash, err := password.HashPassword(randomPass)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/hash").SetError(err.Error())
	}

	user = &domain.User{
		ID:           uuid.New(),
		CompanyID:    companyEntity.ID,
		RoleID:       domain.RoleOwner,
		Email:        &email,
		PasswordHash: passwordHash,
	}
	if fullName != "" {
		user.FullName = &fullName
	} else {
		user.FullName = &email
	}

	if err := s.userSvc.CreateUserWithTx(ctx, tx, user); err != nil {
		if err.GetServerError() == usermod.ServiceErrorUserAlreadyExists {
			return nil, srverr.NewServerError(ServiceErrorAuthUserAlreadyExists, "auth.findOrCreateOAuthUser/duplicate")
		}
		return nil, err
	}
	if err = s.transaction.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/commit").SetError(err.Error())
	}
	return s.issueTokens(ctx, user)
}

func randomPasswordForOAuth() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *service) StartGoogleOAuth(ctx context.Context) (redirectURL string, sErr srverr.ServerError) {
	if s.oauthState == nil || s.googleClient == nil {
		return "", srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.StartGoogleOAuth/not_configured")
	}
	state, err := randomState()
	if err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartGoogleOAuth/state").SetError(err.Error())
	}
	ttl := secToDur(s.cfg.Auth.OAuthStateTTLSec)
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if err := s.oauthState.setState(ctx, state, ttl); err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartGoogleOAuth/set_state").SetError(err.Error())
	}
	return s.googleClient.AuthCodeURL(state), nil
}

func (s *service) StartAppleOAuth(ctx context.Context) (redirectURL string, sErr srverr.ServerError) {
	if s.oauthState == nil || s.appleClient == nil {
		return "", srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.StartAppleOAuth/not_configured")
	}
	state, err := randomState()
	if err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartAppleOAuth/state").SetError(err.Error())
	}
	ttl := secToDur(s.cfg.Auth.OAuthStateTTLSec)
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if err := s.oauthState.setState(ctx, state, ttl); err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartAppleOAuth/set_state").SetError(err.Error())
	}
	redirectURI := strings.TrimSuffix(s.cfg.Auth.OAuthBackendBaseURL, "/") + s.cfg.Auth.OAuthAppleRedirectPath
	return s.appleClient.AuthCodeURL(redirectURI, state), nil
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

*/
