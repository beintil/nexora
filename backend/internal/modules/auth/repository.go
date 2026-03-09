package auth

import (
	"context"
	"telephony/internal/shared/cache"
	"time"

	"github.com/google/uuid"
)

const (
	authRefreshTokenKey         = "auth:refresh_token:"
	authAccessTokenKey          = "auth:access_token:"
	authRefreshTokenByUserIDKey = "auth:refresh_token_by_user_id:"
	authAccessTokenByUserIDKey  = "auth:access_token_by_user_id:"

	authConfirmEmailKey  = "auth:link:"
	authOAuthStateKey    = "auth:oauth:state:"
	authPasswordResetKey = "auth:password:reset:"
)

type repoImpl struct {
	store cache.Cache
}

func newRepository(store cache.Cache) repository {
	return &repoImpl{
		store: store,
	}
}

func (r *repoImpl) setVerifyEmailLink(ctx context.Context, token, email string, ttl time.Duration) error {
	return r.store.Set(ctx, authConfirmEmailKey+token, email, ttl)
}

func (r *repoImpl) getVerifyEmailLink(ctx context.Context, token string) (string, error) {
	var email string
	if err := r.store.Get(ctx, authConfirmEmailKey+token, &email); err != nil {
		return "", err
	}
	return email, nil
}

func (r *repoImpl) deleteVerifyEmailLink(ctx context.Context, token string) error {
	return r.store.Delete(ctx, authConfirmEmailKey+token)
}

func (r *repoImpl) setTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) error {
	userIDStr := userID.String()

	if err := r.store.Set(ctx, authRefreshTokenKey+refreshToken, userIDStr, refreshTTL); err != nil {
		return err
	}
	if err := r.store.Set(ctx, authAccessTokenKey+accessToken, userIDStr, accessTTL); err != nil {
		return err
	}
	if err := r.store.Set(ctx, authAccessTokenByUserIDKey+userIDStr, accessToken, accessTTL); err != nil {
		return err
	}
	if err := r.store.Set(ctx, authRefreshTokenByUserIDKey+userIDStr, refreshToken, refreshTTL); err != nil {
		return err
	}
	return nil
}

func (r *repoImpl) getUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	var userID string
	if err := r.store.Get(ctx, authRefreshTokenKey+refreshToken, &userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (r *repoImpl) deleteTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	userIDStr := userID.String()

	var accessToken string
	if err := r.store.Get(ctx, authAccessTokenByUserIDKey+userIDStr, &accessToken); err == nil {
		_ = r.store.Delete(ctx, authAccessTokenKey+accessToken)
	}
	_ = r.store.Delete(ctx, authAccessTokenByUserIDKey+userIDStr)

	var refreshToken string
	if err := r.store.Get(ctx, authRefreshTokenByUserIDKey+userIDStr, &refreshToken); err == nil {
		_ = r.store.Delete(ctx, authRefreshTokenKey+refreshToken)
	}
	_ = r.store.Delete(ctx, authRefreshTokenByUserIDKey+userIDStr)

	return nil
}

func (r *repoImpl) deleteTokensByRefreshToken(ctx context.Context, refreshToken string) error {
	var userIDStr string
	if err := r.store.Get(ctx, authRefreshTokenKey+refreshToken, &userIDStr); err != nil {
		return err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	return r.deleteTokensByUserID(ctx, userID)
}

func (r *repoImpl) setOAuthState(ctx context.Context, state string, ttl time.Duration) error {
	return r.store.Set(ctx, authOAuthStateKey+state, "1", ttl)
}

func (r *repoImpl) getOAuthState(ctx context.Context, state string) (string, error) {
	var val string
	if err := r.store.Get(ctx, authOAuthStateKey+state, &val); err != nil {
		return "", err
	}
	return val, nil
}

func (r *repoImpl) deleteOAuthState(ctx context.Context, state string) error {
	return r.store.Delete(ctx, authOAuthStateKey+state)
}

func (r *repoImpl) setPasswordResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return r.store.Set(ctx, authPasswordResetKey+token, userID, ttl)
}

func (r *repoImpl) getPasswordResetToken(ctx context.Context, token string) (string, error) {
	var userID string
	if err := r.store.Get(ctx, authPasswordResetKey+token, &userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (r *repoImpl) deletePasswordResetToken(ctx context.Context, token string) error {
	return r.store.Delete(ctx, authPasswordResetKey+token)
}
