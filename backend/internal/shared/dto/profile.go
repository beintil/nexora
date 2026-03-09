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

// UsersDomainToModel конвертирует []*domain.User в *models.UsersListResponse.
func UsersDomainToModel(users []*domain.User) *models.UsersListResponse {
	res := &models.UsersListResponse{
		Users: make([]*models.ProfileResponse, 0, len(users)),
	}
	for _, u := range users {
		res.Users = append(res.Users, UserToProfileModel(u))
	}
	return res
}

// UserToProfileModel конвертирует domain.User в models.ProfileResponse.
func UserToProfileModel(u *domain.User) *models.ProfileResponse {
	if u == nil {
		return nil
	}
	res := &models.ProfileResponse{
		ID:        strfmt.UUID(u.ID.String()),
		CompanyID: strfmt.UUID(u.CompanyID.String()),
		RoleID:    int16(u.RoleID),
		CreatedAt: strfmt.DateTime(u.CreatedAt),
		UpdatedAt: strfmt.DateTime(u.UpdatedAt),
	}
	if u.Email != nil {
		res.Email = strfmt.Email(*u.Email)
	}
	if u.FullName != nil {
		res.FullName = *u.FullName
	}
	if u.AvatarURL != nil {
		res.AvatarURL = *u.AvatarURL
	}
	if u.AvatarID != nil {
		res.AvatarID = *u.AvatarID
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
