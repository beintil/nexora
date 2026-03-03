package validator

import (
	"fmt"
	"strings"
	"unicode"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

var emailValidator = govalidator.New()

// ValidateEmail проверяет формат email. Возвращает нормализованный email или ошибку.
func ValidateEmail(email string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(email))
	if s == "" {
		return "", fmt.Errorf("ValidateEmail: email is empty")
	}
	if err := emailValidator.Var(s, "required,email,max=255"); err != nil {
		return "", fmt.Errorf("ValidateEmail/Var: %w", err)
	}
	return s, nil
}

// ValidatePhone проверяет и нормализует номер телефона в E.164.
// Номер должен быть в международном формате с кодом страны (начинаться с +), иначе возвращается ошибка.
func ValidatePhone(phone string, regionCode string) (string, error) {
	s := strings.TrimSpace(phone)
	if s == "" {
		return "", fmt.Errorf("ValidatePhone: phone is empty")
	}
	if !strings.HasPrefix(s, "+") {
		return "", fmt.Errorf("ValidatePhone: phone must include country code (international format, e.g. +79001234567)")
	}
	if regionCode == "" {
		regionCode = "ZZ" // для номера с + код страны берётся из номера
	}
	parsed, err := phonenumbers.Parse(s, regionCode)
	if err != nil {
		return "", fmt.Errorf("ValidatePhone/Parse: %w", err)
	}
	if !phonenumbers.IsValidNumber(parsed) {
		return "", fmt.Errorf("ValidatePhone/IsValidNumber: phone number is not valid")
	}
	return phonenumbers.Format(parsed, phonenumbers.E164), nil
}

// ValidatePassword проверяет пароль: не менее 6 символов, минимум одна заглавная буква и один спецсимвол.
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("ValidatePassword: password must be at least 6 characters")
	}
	var hasUpper, hasSpecial bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r) {
			hasSpecial = true
		}
		if hasUpper && hasSpecial {
			return nil
		}
	}
	if !hasUpper {
		return fmt.Errorf("ValidatePassword: password must contain at least one uppercase letter")
	}
	return fmt.Errorf("ValidatePassword: password must contain at least one special character")
}
