package dto

import (
	"telephony/internal/domain"
	"telephony/models"

	"github.com/go-openapi/strfmt"
)

// CompanyTelephonyCreateRequestToParams извлекает из models.CompanyTelephonyCreateRequest параметры для сервиса (после Validate).
func CompanyTelephonyCreateRequestToParams(req *models.CompanyTelephonyCreateRequest) (telephonyName domain.TelephonyName, externalAccountID string) {
	if req == nil {
		return "", ""
	}
	if req.TelephonyName != nil {
		telephonyName = domain.TelephonyName(*req.TelephonyName)
	}
	if req.ExternalAccountID != nil {
		externalAccountID = *req.ExternalAccountID
	}
	return telephonyName, externalAccountID
}

// CompanyTelephonyDomainToModel конвертирует domain.CompanyTelephone в models.CompanyTelephonyItem.
func CompanyTelephonyDomainToModel(ct *domain.CompanyTelephone) *models.CompanyTelephonyItem {
	if ct == nil {
		return nil
	}
	name := ""
	if ct.Telephone != nil {
		name = string(ct.Telephone.Name)
	}
	return &models.CompanyTelephonyItem{
		ID:                strfmt.UUID(ct.ID.String()),
		TelephonyName:     name,
		ExternalAccountID: ct.ExternalAccountID,
		CreatedAt:         strfmt.DateTime(ct.CreatedAt),
	}
}

// CompanyTelephonyListDomainToModel конвертирует список domain.CompanyTelephone в models.CompanyTelephonyListResponse.
func CompanyTelephonyListDomainToModel(list []*domain.CompanyTelephone) *models.CompanyTelephonyListResponse {
	if list == nil {
		return &models.CompanyTelephonyListResponse{Items: []*models.CompanyTelephonyItem{}}
	}
	return &models.CompanyTelephonyListResponse{
		Items: MapSlice(list, CompanyTelephonyDomainToModel),
	}
}

// TelephonyDomainToModel конвертирует domain.Telephony в models.TelephonyDictionaryItem.
func TelephonyDomainToModel(t *domain.Telephony) *models.TelephonyDictionaryItem {
	if t == nil {
		return nil
	}
	return &models.TelephonyDictionaryItem{
		ID:   t.ID,
		Name: string(t.Name),
	}
}

// TelephonyDictionaryDomainToModel конвертирует список domain.Telephony в models.TelephonyDictionaryResponse.
func TelephonyDictionaryDomainToModel(list []*domain.Telephony) *models.TelephonyDictionaryResponse {
	if list == nil {
		return &models.TelephonyDictionaryResponse{Items: []*models.TelephonyDictionaryItem{}}
	}
	return &models.TelephonyDictionaryResponse{
		Items: MapSlice(list, TelephonyDomainToModel),
	}
}
