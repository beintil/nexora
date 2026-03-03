package countries

import (
	"context"
	"telephony/internal/domain"
	srverr "telephony/internal/shared/server_error"

	"github.com/jackc/pgx/v5"
)

type Service interface {
	GetCountryByFullNameWithTx(ctx context.Context, tx pgx.Tx, fullName string) (*domain.Country, srverr.ServerError)
	GetCountryByCodeWithTx(ctx context.Context, tx pgx.Tx, code string) (*domain.Country, srverr.ServerError)
	SaveUpdateCountries(ctx context.Context) srverr.ServerError
}

type repository interface {
	saveUpdateCountries(ctx context.Context, tx pgx.Tx, countries []*domain.Country) error
	getCountryByCode(ctx context.Context, tx pgx.Tx, code string) (*domain.Country, error)
	getCountryByFullName(ctx context.Context, tx pgx.Tx, fullName string) (*domain.Country, error)
}
