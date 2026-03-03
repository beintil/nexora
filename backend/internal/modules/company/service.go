package company

import (
	"context"
	"errors"
	"telephony/internal/domain"
	"telephony/internal/modules/call"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	callService call.Service
	repo        repository
	pool        postgres.Transaction
}

func NewService(
	callService call.Service,
	repo repository,
	pool postgres.Transaction,
) Service {
	return &service{
		callService: callService,
		repo:        repo,
		pool:        pool,
	}
}

const (
	ServiceErrorCompanyTelephonyNotFound srverr.ErrorTypeNotFound = "company_telephony_not_found"
	ServiceErrorCompanyNotFound          srverr.ErrorTypeNotFound = "company_not_found"

	ServiceErrorCallIsNotValid srverr.ErrorTypeBadRequest = "call_is_not_valid"
)

func (s *service) GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx(
	ctx context.Context,
	tx pgx.Tx,
	externalID string,
	telephonyName domain.TelephonyName,
) (*domain.CompanyTelephone, srverr.ServerError) {
	companyTelephony, err := s.repo.getCompanyTelephonyByExternalAccountIDAndTelephonyName(ctx, tx, externalID, telephonyName)
	if err != nil {
		if errors.Is(err, errRepoTelephonyNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyNotFound, "company.GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx").SetError(err.Error())
	}
	return companyTelephony, nil
}

func (s *service) GetCompanyByNameWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError) {
	c, err := s.repo.getCompanyByName(ctx, tx, name)
	if err != nil {
		if errors.Is(err, errRepoCompanyNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCompanyNotFound, "company.GetCompanyByNameWithTx")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.GetCompanyByNameWithTx").SetError(err.Error())
	}
	return c, nil
}

func (s *service) GetCompanyByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, srverr.ServerError) {
	c, err := s.repo.getCompanyByID(ctx, tx, id)
	if err != nil {
		if errors.Is(err, errRepoCompanyNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCompanyNotFound, "company.GetCompanyByIDWithTx")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.GetCompanyByIDWithTx").SetError(err.Error())
	}
	return c, nil
}

func (s *service) CreateCompanyWithTx(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, srverr.ServerError) {
	c := &domain.Company{
		ID:   uuid.New(),
		Name: name,
	}
	err := s.repo.createCompany(ctx, tx, c)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.CreateCompanyWithTx").SetError(err.Error())
	}
	return c, nil
}
