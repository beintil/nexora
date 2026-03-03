package dto

import (
	"telephony/internal/domain"
	"telephony/models"

	"github.com/go-openapi/strfmt"
)

// ProfileDomainToModel конвертирует domain.Profile в models.ProfileResponse.
func ProfileDomainToModel(p *domain.Profile) *models.ProfileResponse {
	if p == nil {
		return nil
	}
	res := &models.ProfileResponse{
		ID:          strfmt.UUID(p.ID.String()),
		CompanyID:   strfmt.UUID(p.CompanyID.String()),
		CompanyName: p.CompanyName,
		RoleID:      int16(p.RoleID),
		CreatedAt:   strfmt.DateTime(p.CreatedAt),
		UpdatedAt:   strfmt.DateTime(p.UpdatedAt),
	}
	if p.Email != nil {
		res.Email = strfmt.Email(*p.Email)
	}
	if p.FullName != nil {
		res.FullName = *p.FullName
	}
	if p.AvatarURL != nil {
		res.AvatarURL = *p.AvatarURL
	}
	if p.AvatarID != nil {
		res.AvatarID = *p.AvatarID
	}
	return res
}

// UpdateProfileRequestToDomain конвертирует models.UpdateProfileRequest в domain.UpdateProfileInput.
func UpdateProfileRequestToDomain(req *models.UpdateProfileRequest) *domain.UpdateProfileInput {
	if req == nil {
		return nil
	}
	out := &domain.UpdateProfileInput{}
	if req.FullName != "" {
		out.FullName = &req.FullName
	}
	return out
}
