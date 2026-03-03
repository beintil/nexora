package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword возвращает bcrypt-хеш пароля (соль встроена в хеш).
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("HashPassword/GenerateFromPassword: %w", err)
	}
	return string(hash), nil
}

// ComparePassword сравнивает пароль с хешем. Возвращает true, если пароль совпадает.
func ComparePassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
