package dto

import (
	"telephony/internal/domain"
	"telephony/models"
)

// RegisterRequestToDomain конвертирует models.RegisterRequest в domain.AuthRegisterInput.
func RegisterRequestToDomain(req *models.RegisterRequest) *domain.AuthRegisterInput {
	if req == nil {
		return nil
	}
	out := &domain.AuthRegisterInput{
		Email:    req.Email.String(),
		FullName: req.FullName,
	}
	if req.CompanyName != nil {
		out.CompanyName = *req.CompanyName
	}
	if req.Password != nil {
		out.Password = req.Password.String()
	}
	return out
}

// LoginRequestToDomain конвертирует models.LoginRequest в domain.AuthLoginInput.
func LoginRequestToDomain(req *models.LoginRequest) *domain.AuthLoginInput {
	if req == nil {
		return nil
	}
	out := &domain.AuthLoginInput{}
	if req.Login != nil {
		out.Email = *req.Login
	}
	if req.Password != nil {
		out.Password = req.Password.String()
	}
	return out
}

// RefreshRequestToToken возвращает refresh token из models.RefreshRequest.
func RefreshRequestToToken(req *models.RefreshRequest) string {
	if req == nil || req.RefreshToken == nil {
		return ""
	}
	return *req.RefreshToken
}
