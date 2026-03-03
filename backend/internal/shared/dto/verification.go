package dto

import (
	"telephony/internal/domain"
	"telephony/models"
)

// SendCodeRequestToDomain конвертирует models.SendCodeRequest в domain.SendCodeInput.
func SendCodeRequestToDomain(req *models.SendCodeRequest) *domain.SendCodeInput {
	if req == nil {
		return nil
	}
	out := &domain.SendCodeInput{}
	if req.Email.String() != "" {
		out.Email = req.Email.String()
	}
	return out
}

// VerifyLinkRequestToDomain конвертирует models.VerifyLinkRequest в domain.VerifyLinkInput.
func VerifyLinkRequestToDomain(req *models.VerifyLinkRequest) *domain.VerifyLinkInput {
	if req == nil {
		return &domain.VerifyLinkInput{}
	}
	out := &domain.VerifyLinkInput{}
	if req.Token != nil {
		out.Token = req.Token.String()
	}
	return out
}
