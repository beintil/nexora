package appleoauth

import "errors"

var (
	ErrAppleInvalidToken = errors.New("apple id_token invalid or verification failed")
	ErrAppleJWKS         = errors.New("apple jwks fetch or key failed")
)
