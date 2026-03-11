package password

import (
	"crypto/rand"
	"fmt"
	"math/big"

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

func GeneratePassword() string {
	const (
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberChars  = "0123456789"
		specialChars = "!@#$%^&*()-_+="
		allChars     = lowerChars + upperChars + numberChars + specialChars
	)

	// У нас длина должна быть больше 6, сделаем надежные 12 символов
	length := 12
	password := make([]byte, length)

	// Гарантируем обязательные символы: заглавная буква и спецсимвол (а также строчную и цифру на всякий случай)
	password[0] = upperChars[cryptoRandInt(len(upperChars))]
	password[1] = specialChars[cryptoRandInt(len(specialChars))]
	password[2] = numberChars[cryptoRandInt(len(numberChars))]
	password[3] = lowerChars[cryptoRandInt(len(lowerChars))]

	// Остальные символы случайны из всех
	for i := 4; i < length; i++ {
		password[i] = allChars[cryptoRandInt(len(allChars))]
	}

	// Перемешаем (Fisher-Yates) чтобы шаблоны не всегда были вначале
	for i := len(password) - 1; i > 0; i-- {
		j := cryptoRandInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

func cryptoRandInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
