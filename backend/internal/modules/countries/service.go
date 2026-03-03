package countries

import (
	"context"
	"errors"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"
	"telephony/pkg/client/country"

	"github.com/jackc/pgx/v5"
)

type service struct {
	repository repository

	countriesClient country.Country

	pool postgres.Transaction
}

func NewService(
	repository repository,
	pool postgres.Transaction,
	countriesClient country.Country,
) Service {
	return &service{
		repository:      repository,
		pool:            pool,
		countriesClient: countriesClient,
	}
}

var (
	ServiceErrorCountryNotFound     srverr.ErrorTypeNotFound   = "country_not_found"
	ServiceErrorCountryCodeNotValid srverr.ErrorTypeBadRequest = "country_code_not_valid"
)

func (m *service) GetCountryByFullNameWithTx(ctx context.Context, tx pgx.Tx, fullName string) (*domain.Country, srverr.ServerError) {
	c, err := m.repository.getCountryByFullName(ctx, tx, fullName)
	if err != nil {
		if errors.Is(err, errRepoCountryNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCountryNotFound, "countries.GetCountryByFullName/getCountryByFullName").
				SetDetails("country not found")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "countries.GetCountryByFullName/getCountryByFullName").
			SetError(err.Error())
	}
	return c, nil
}

func (m *service) GetCountryByCodeWithTx(ctx context.Context, tx pgx.Tx, code string) (*domain.Country, srverr.ServerError) {
	if len(code) != 2 {
		return nil, srverr.NewServerError(ServiceErrorCountryCodeNotValid, "countries.GetCountryByCode/invalid_code_length").
			SetDetails("code must be 2 characters long")
	}
	c, err := m.repository.getCountryByCode(ctx, tx, code)
	if err != nil {
		if errors.Is(err, errRepoCountryNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCountryNotFound, "countries.GetCountryByCode/getCountryByCode").
				SetDetails("country not found")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "countries.GetCountryByCode/getCountryByCode").
			SetError(err.Error())
	}
	return c, nil
}

func (m *service) SaveUpdateCountries(ctx context.Context) srverr.ServerError {
	apiCountries, err := m.countriesClient.GetAllCountries(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "countries.SaveUpdateCountries/GetAllCountries").
			SetError(err.Error())
	}
	var countries = make([]*domain.Country, 0, len(apiCountries))

	for _, apiCountry := range apiCountries {
		countries = append(countries, &domain.Country{
			Code:        apiCountry.Cca2,
			Name:        apiCountry.Name.Official,
			Description: apiCountry.Name.Common,
		})
	}
	tx, err := m.pool.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "countries.SaveUpdateCountries/BeginTransaction").
			SetError(err.Error())
	}
	defer m.pool.Rollback(ctx, tx)

	err = m.repository.saveUpdateCountries(ctx, tx, countries)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "countries.SaveUpdateCountries/saveUpdateCountries").
			SetError(err.Error())
	}
	err = m.pool.Commit(ctx, tx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "countries.SaveUpdateCountries/Commit").
			SetError(err.Error())
	}
	return nil
}
