package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	"telephony/internal/modules/company"
	"telephony/internal/modules/message_delivery"
	usermod "telephony/internal/modules/user"
	"telephony/internal/shared/cache"
	"telephony/internal/shared/database/postgres"
	"telephony/internal/shared/server_error"
	"telephony/internal/shared/templates"
	"telephony/pkg/client/oauth/appleoauth"
	"telephony/pkg/client/oauth/googleoauth"
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
	repo        repository
	cfg         config.Config

	msgDelivery  message_delivery.Service
	googleClient googleoauth.Client
	appleClient  appleoauth.Client
}

func NewService(
	userSvc usermod.Service,
	companySvc company.Service,
	transaction postgres.Transaction,
	store cache.Cache,
	cfg config.Config,

	msgDelivery message_delivery.Service,
	googleClient googleoauth.Client,
	appleClient appleoauth.Client,
) Service {
	return &service{
		userSvc:     userSvc,
		companySvc:  companySvc,
		transaction: transaction,
		repo:        newRepository(store),
		cfg:         cfg,

		msgDelivery:  msgDelivery,
		googleClient: googleClient,
		appleClient:  appleClient,
	}
}

const (
	ServiceErrorAuthInvalidBrainPrivilegedUser srverr.ErrorTypeUnauthorized = "Privileged accounts cannot log in via the client application. Please use the admin panel."

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

const (
	ServiceErrorAuthOAuthNotConfigured srverr.ErrorTypeBadRequest   = "oauth_not_configured"
	ServiceErrorAuthOAuthInvalidState  srverr.ErrorTypeUnauthorized = "oauth_invalid_state"
	ServiceErrorAuthOAuthNoEmail       srverr.ErrorTypeUnauthorized = "oauth_no_email"
	ServiceErrorAuthUserAlreadyExists  srverr.ErrorTypeConflict     = "user_already_exists"

	ServiceErrorAuthResetTokenInvalid srverr.ErrorTypeBadRequest = "password_reset_token_invalid"
)

var errOAuthStateNotFoundOrExpired = errors.New("oauth state not found or expired")

func (s *service) Register(ctx context.Context, req *domain.AuthRegisterInput) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.Register/nil_request")
	}
	emailNorm, err := validator.ValidateEmail(req.Email)
	if err != nil {
		return srverr.NewServerError(ServiceErrorAuthInvalidEmail, "auth.Register/email").SetError(err.Error())
	}
	err = validator.ValidatePassword(req.Password)
	if err != nil {
		return srverr.NewServerError(ServiceErrorAuthInvalidPassword, "auth.Register/password").SetError(err.Error())
	}

	companyName := strings.TrimSpace(req.CompanyName)
	if companyName == "" || len(companyName) > 100 || len(companyName) < 5 {
		return srverr.NewServerError(ServiceErrorAuthInvalidCompanyName, "auth.Register/empty_company")
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.Register/begin").SetError(err.Error())
	}
	defer s.transaction.Rollback(ctx, tx)

	{
		u, serverErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, emailNorm)
		if serverErr != nil && serverErr.GetServerError() != usermod.ServiceErrorUserNotFound {
			return serverErr
		}
		if u != nil {
			if u.VerifiedRegistration {
				return srverr.NewServerError(ServiceErrorAuthInvalidCredentials, "auth.Register/user_already_verified")
			} else {
				serverErr = s.sendCodeWithTx(ctx, tx, &domain.SendCodeInput{Email: emailNorm})
				if serverErr != nil {
					return serverErr
				}
				return srverr.NewServerError(ServiceErrorAuthConfirmEmail, "auth.Register/user_already_registered")
			}
		}
	}

	companyEntity, serverErr := s.companySvc.CreateCompanyWithTx(ctx, tx, companyName)
	if serverErr != nil {
		return serverErr
	}

	passwordHash, err := password.HashPassword(req.Password)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.RegisterWithTx/hash").SetError(err.Error())
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
		return serverErr
	}
	serverErr = s.sendCodeWithTx(ctx, tx, &domain.SendCodeInput{Email: emailNorm})
	if serverErr != nil {
		return serverErr
	}

	if err = s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.RegisterWithTx/commit").SetError(err.Error())
	}

	return nil
}

func (s *service) VerifyLink(ctx context.Context, req *domain.VerifyLinkInput) srverr.ServerError {
	if req == nil {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLinkWithTx/nil")
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLinkWithTx/empty_token")
	}
	if len(token) != 6 || !isAllDigits(token) {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/invalid_code")
	}
	if strings.TrimSpace(req.Email) == "" {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/empty_email")
	}
	emailNorm, err := validator.ValidateEmail(req.Email)
	if err != nil {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/invalid_email").SetError(err.Error())
	}

	storedEmail, err := s.repo.getVerifyEmailLink(ctx, token)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/link_expired_or_invalid")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/link_store").SetError(err.Error())
	}
	if storedEmail != emailNorm {
		return srverr.NewServerError(ServiceErrorVerifyLink, "auth.VerifyLink/email_mismatch")
	}
	err = s.repo.deleteVerifyEmailLink(ctx, token)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/link_store").SetError(err.Error())
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.VerifyLinkWithTx/begin").SetError(err.Error())
	}
	defer s.transaction.Rollback(ctx, tx)

	u, serverErr := s.userSvc.GetUserByEmailWithTx(ctx, tx, emailNorm)
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

	code, err := generateVerificationCode()
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendCode/generate_code").SetError(err.Error())
	}
	authLinkTTL := time.Duration(s.cfg.Auth.AuthLinkTTLSec) * time.Second
	if err := s.repo.setVerifyEmailLink(ctx, code, emailNorm, authLinkTTL); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendCode/create_link").SetError(err.Error())
	}

	link := s.cfg.Auth.AuthLinkBaseURL + code + "&email=" + url.QueryEscape(emailNorm)
	render, err := templates.NewRenderer()
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/renderer").SetError(err.Error())
	}
	body, err := render.Render(templates.HTMLFileRegister, map[string]interface{}{
		"Link": link,
		"Code": code,
	})
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/render").SetError(err.Error())
	}

	if emailNorm != "" {
		if err := s.msgDelivery.Send(ctx, &domain.OutgoingMessage{
			To:      emailNorm,
			Subject: "Регистрация Аккаунта",
			Body:    fmt.Sprintf("Ваш код подтверждения: %s\n\nПерейдите по ссылке для подтверждения: %s", code, link),
			HTML:    body.String(),
		}, []domain.DeliveryChannel{domain.DeliveryChannelEmail}); err != nil {
			return srverr.NewServerError(srverr.ErrInternalServerError, "verification.SendCode/email").SetError(err.Error())
		}
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

	if !user.VerifiedRegistration {
		serverErr := s.sendCodeWithTx(ctx, tx, &domain.SendCodeInput{Email: emailNorm})
		if serverErr != nil {
			return nil, serverErr
		}
		return nil, srverr.NewServerError(ServiceErrorAuthConfirmEmail, "auth.Login/not_verified")
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
	userIDStr, err := s.repo.getUserIDByRefreshToken(ctx, refreshToken)
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
	userIDStr, err := s.repo.getUserIDByRefreshToken(ctx, refreshToken)
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
	if delErr := s.repo.deleteTokensByUserID(ctx, userID); delErr != nil {
		if se, ok := delErr.(srverr.ServerError); ok {
			return se
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.Logout/delete").SetError(delErr.Error())
	}
	return nil
}

// DeleteTokensByUserID is handled by repo now

func (s *service) issueTokens(ctx context.Context, user *domain.User) (*domain.AuthTokens, srverr.ServerError) {
	// Удаляем существующие токены если существуют
	err := s.repo.deleteTokensByUserID(ctx, user.ID)
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

	if err := s.repo.setTokens(ctx, user.ID, accessToken, refreshToken, accessTTL, refreshTTL); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.issueTokens/store").SetError(err.Error())
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// --- OAuth ---

func (s *service) StartGoogleOAuth(ctx context.Context) (string, srverr.ServerError) {
	if s.googleClient == nil {
		return "", srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.StartGoogleOAuth/not_configured")
	}
	state, err := generateRandomString(32)
	if err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartGoogleOAuth/state").SetError(err.Error())
	}
	ttl := secToDur(s.cfg.Auth.OAuth.OAuthStateTTLSec)
	if err := s.setOAuthState(ctx, state, ttl); err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartGoogleOAuth/set_state").SetError(err.Error())
	}
	return s.googleClient.AuthCodeURL(state), nil
}

func (s *service) LoginOrRegisterWithGoogle(ctx context.Context, code, state string) (*domain.AuthTokens, srverr.ServerError) {
	if s.googleClient == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.LoginOrRegisterWithGoogle/not_configured")
	}
	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.LoginOrRegisterWithGoogle/empty_code_or_state")
	}
	if err := s.consumeOAuthState(ctx, state); err != nil {
		if errors.Is(err, errOAuthStateNotFoundOrExpired) {
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

func (s *service) StartAppleOAuth(ctx context.Context) (string, srverr.ServerError) {
	if s.appleClient == nil {
		return "", srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.StartAppleOAuth/not_configured")
	}
	state, err := generateRandomString(32)
	if err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartAppleOAuth/state").SetError(err.Error())
	}
	ttl := secToDur(s.cfg.Auth.OAuth.OAuthStateTTLSec)
	if err := s.setOAuthState(ctx, state, ttl); err != nil {
		return "", srverr.NewServerError(srverr.ErrInternalServerError, "auth.StartAppleOAuth/set_state").SetError(err.Error())
	}
	redirectURI := strings.TrimSuffix(s.cfg.Auth.OAuth.OAuthBackendBaseURL, "/") + s.cfg.Auth.OAuth.OAuthApple.RedirectPath
	return s.appleClient.AuthCodeURL(redirectURI, state), nil
}

func (s *service) LoginOrRegisterWithApple(ctx context.Context, idToken, state string) (*domain.AuthTokens, srverr.ServerError) {
	if s.appleClient == nil {
		return nil, srverr.NewServerError(ServiceErrorAuthOAuthNotConfigured, "auth.LoginOrRegisterWithApple/not_configured")
	}
	idToken = strings.TrimSpace(idToken)
	state = strings.TrimSpace(state)
	if idToken == "" || state == "" {
		return nil, srverr.NewServerError(ServiceErrorAuthRequestIsNotValid, "auth.LoginOrRegisterWithApple/empty_id_token_or_state")
	}
	if err := s.consumeOAuthState(ctx, state); err != nil {
		if errors.Is(err, errOAuthStateNotFoundOrExpired) {
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

// findOrCreateOAuthUser finds user by email or creates a new company + user (OAuth: random password, verified by default).
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

	randomPass, err := generateRandomString(32)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/random_password").SetError(err.Error())
	}
	passwordHash, err := password.HashPassword(randomPass)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/hash").SetError(err.Error())
	}

	user = &domain.User{
		ID:                   uuid.New(),
		CompanyID:            companyEntity.ID,
		RoleID:               domain.RoleOwner,
		Email:                &email,
		PasswordHash:         passwordHash,
		VerifiedRegistration: true,
	}
	if fullName != "" {
		user.FullName = &fullName
	} else {
		user.FullName = &email
	}

	if createErr := s.userSvc.CreateUserWithTx(ctx, tx, user); createErr != nil {
		if createErr.GetServerError() == usermod.ServiceErrorUserAlreadyExists {
			return nil, srverr.NewServerError(ServiceErrorAuthUserAlreadyExists, "auth.findOrCreateOAuthUser/duplicate")
		}
		return nil, createErr
	}
	if err = s.transaction.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "auth.findOrCreateOAuthUser/commit").SetError(err.Error())
	}
	return s.issueTokens(ctx, user)
}

// --- OAuth state helpers ---

func (s *service) setOAuthState(ctx context.Context, state string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return s.repo.setOAuthState(ctx, state, ttl)
}

func (s *service) consumeOAuthState(ctx context.Context, state string) error {
	_, err := s.repo.getOAuthState(ctx, state)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return errOAuthStateNotFoundOrExpired
		}
		return err
	}
	_ = s.repo.deleteOAuthState(ctx, state)
	return nil
}

func secToDur(sec int) time.Duration {
	return time.Duration(sec) * time.Second
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateVerificationCode() (string, error) {
	max := big.NewInt(1000000) // 0..999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// --- Password Management ---

func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) srverr.ServerError {
	if err := validator.ValidatePassword(newPassword); err != nil {
		return srverr.NewServerError(ServiceErrorAuthInvalidPassword, "auth.ChangePassword/new_password_validate").SetError(err.Error())
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ChangePassword/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	user, sErr := s.userSvc.GetUserByIDWithTx(ctx, tx, userID)
	if sErr != nil {
		return sErr
	}

	if !password.ComparePassword(user.PasswordHash, oldPassword) {
		return srverr.NewServerError(ServiceErrorAuthInvalidCredentials, "auth.ChangePassword/old_password_compare")
	}

	newHash, err := password.HashPassword(newPassword)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ChangePassword/hash").SetError(err.Error())
	}

	if sErr := s.userSvc.UpdateUserPasswordWithTx(ctx, tx, user.ID, newHash); sErr != nil {
		return sErr
	}

	if err := s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ChangePassword/commit").SetError(err.Error())
	}

	return nil
}

func (s *service) SendResetPasswordLink(ctx context.Context, email string) srverr.ServerError {
	emailNorm, err := validator.ValidateEmail(email)
	if err != nil {
		return srverr.NewServerError(ServiceErrorAuthInvalidEmail, "auth.SendResetPasswordLink/email_validate").SetError(err.Error())
	}

	user, sErr := s.userSvc.GetUserByEmail(ctx, emailNorm)
	if sErr != nil {
		if sErr.GetServerError() == usermod.ServiceErrorUserNotFound {
			// Don't reveal if user exists for security
			return nil
		}
		return sErr
	}

	token, err := generateRandomString(32)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendResetPasswordLink/token_gen").SetError(err.Error())
	}

	// Token valid for 1 hour
	if err := s.repo.setPasswordResetToken(ctx, token, user.ID.String(), 1*time.Hour); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendResetPasswordLink/store").SetError(err.Error())
	}

	if err := s.msgDelivery.Send(ctx, &domain.OutgoingMessage{
		To:      emailNorm,
		Subject: "Password Reset",
		HTML: fmt.Sprintf("Reset your password by clicking here: %s",
			fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(s.cfg.Auth.AuthLinkBaseURL, "/"), token),
		),
	},
		[]domain.DeliveryChannel{domain.DeliveryChannelEmail}); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.SendResetPasswordLink/email_send").SetError(err.Error())
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, token, newPassword string) srverr.ServerError {
	if err := validator.ValidatePassword(newPassword); err != nil {
		return srverr.NewServerError(ServiceErrorAuthInvalidPassword, "auth.ResetPassword/password_validate").SetError(err.Error())
	}

	userIDStr, err := s.repo.getPasswordResetToken(ctx, token)
	if err != nil {
		if errors.Is(err, cache.ErrorCacheValueNotFound) {
			return srverr.NewServerError(ServiceErrorAuthResetTokenInvalid, "auth.ResetPassword/token_not_found")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ResetPassword/get_token").SetError(err.Error())
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ResetPassword/parse_id").SetError(err.Error())
	}

	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ResetPassword/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	user, sErr := s.userSvc.GetUserByIDWithTx(ctx, tx, userID)
	if sErr != nil {
		return sErr
	}

	newHash, err := password.HashPassword(newPassword)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ResetPassword/hash").SetError(err.Error())
	}

	if sErr := s.userSvc.UpdateUserPasswordWithTx(ctx, tx, user.ID, newHash); sErr != nil {
		return sErr
	}

	if err := s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "auth.ResetPassword/commit").SetError(err.Error())
	}

	_ = s.repo.deletePasswordResetToken(ctx, token)

	return nil
}
