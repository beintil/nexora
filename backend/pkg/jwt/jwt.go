package jwt

import (
	"fmt"
	"telephony/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Role      domain.Role `json:"role"`
	CompanyID string      `json:"company_id"`
}

// BuildAccessToken создаёт JWT access-токен для пользователя.
func BuildAccessToken(userID uuid.UUID, companyID uuid.UUID, role domain.Role, secret []byte, ttlSec time.Duration) (string, error) {
	now := time.Now()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttlSec)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
		Role:      role,
		CompanyID: companyID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseAccessToken проверяет и извлекает user_id и company_id из access-токена.
func ParseAccessToken(tokenString string, secret []byte) (user *domain.UserFromAccess, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("ParseAccessToken: unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("ParseAccessToken: invalid token")
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("ParseAccessToken: invalid user id")
	}
	var companyID uuid.UUID
	if claims.CompanyID != "" {
		companyID, err = uuid.Parse(claims.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("ParseAccessToken: invalid company id")
		}
	}
	return &domain.UserFromAccess{
		ID:        id,
		CompanyID: companyID,
		Role:      claims.Role,
	}, nil
}

type refreshClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// BuildRefreshToken создаёт подписанный JWT refresh-токен (user_id + exp). Хранить в Redis для отзыва.
func BuildRefreshToken(userID string, secret []byte, ttlSec time.Duration) (string, error) {
	now := time.Now()
	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttlSec)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseRefreshToken проверяет подпись и срок, возвращает user_id
func ParseRefreshToken(tokenString string, secret []byte) (userID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}
	return claims.UserID, nil
}
