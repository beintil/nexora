package company

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx(ctx context.Context, tx pgx.Tx, externalID string, telephonyName domain.TelephonyName) (*domain.CompanyTelephone, srverr.ServerError)
	GetCompanyByNameWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError)
	CreateCompanyWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError)
	GetCompanyByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, srverr.ServerError)
}
