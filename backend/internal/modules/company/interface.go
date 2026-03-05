package company

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler interface {
	handleListCompanyTelephony(w http.ResponseWriter, r *http.Request)
	handleAttachCompanyTelephony(w http.ResponseWriter, r *http.Request)
	handleDetachCompanyTelephony(w http.ResponseWriter, r *http.Request)
	handleListTelephonyDictionary(w http.ResponseWriter, r *http.Request)
}

type Service interface {
	GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx(ctx context.Context, tx pgx.Tx, externalID string, telephonyName domain.TelephonyName) (*domain.CompanyTelephone, srverr.ServerError)
	GetCompanyByNameWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError)
	CreateCompanyWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError)
	GetCompanyByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, srverr.ServerError)

	ListTelephonyByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*domain.CompanyTelephone, srverr.ServerError)
	AttachTelephonyToCompany(ctx context.Context, companyID uuid.UUID, telephonyName domain.TelephonyName, externalAccountID string) (*domain.CompanyTelephone, srverr.ServerError)
	DetachTelephonyFromCompany(ctx context.Context, companyID, companyTelephonyID uuid.UUID) srverr.ServerError
	ListTelephonyDictionary(ctx context.Context) ([]*domain.Telephony, srverr.ServerError)
}
