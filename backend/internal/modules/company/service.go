package company

import (
	"context"
	"errors"
	"strings"
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
	ServiceErrorCompanyTelephonyNotFound      srverr.ErrorTypeNotFound   = "company_telephony_not_found"
	ServiceErrorCompanyNotFound               srverr.ErrorTypeNotFound   = "company_not_found"
	ServiceErrorCallIsNotValid                srverr.ErrorTypeBadRequest = "call_is_not_valid"
	ServiceErrorCompanyTelephonyAlreadyExists srverr.ErrorTypeConflict   = "company_telephony_already_exists"
	ServiceErrorCompanyTelephonyBadRequest    srverr.ErrorTypeBadRequest = "company_telephony_bad_request"
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

var allowedTelephonyNames = map[domain.TelephonyName]struct{}{
	domain.Twilio: {}, domain.Mango: {}, domain.Zadarma: {}, domain.MTS: {}, domain.Beeline: {},
}

func (s *service) ListTelephonyByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*domain.CompanyTelephone, srverr.ServerError) {
	if companyID == uuid.Nil {
		return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.ListTelephonyByCompanyID/empty_company_id")
	}
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.ListTelephonyByCompanyID/begin").SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)
	list, err := s.repo.listCompanyTelephonyByCompanyID(ctx, tx, companyID)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.ListTelephonyByCompanyID").SetError(err.Error())
	}
	return list, nil
}

func (s *service) AttachTelephonyToCompany(ctx context.Context, companyID uuid.UUID, telephonyName domain.TelephonyName, externalAccountID string) (*domain.CompanyTelephone, srverr.ServerError) {
	if companyID == uuid.Nil {
		return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.AttachTelephonyToCompany/empty_company_id")
	}
	if _, ok := allowedTelephonyNames[telephonyName]; !ok {
		return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.AttachTelephonyToCompany/invalid_telephony_name")
	}
	externalAccountID = strings.TrimSpace(externalAccountID)
	if externalAccountID == "" {
		return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.AttachTelephonyToCompany/empty_external_account_id")
	}
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.AttachTelephonyToCompany/begin").SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	telephony, err := s.repo.getTelephonyByName(ctx, tx, telephonyName)
	if err != nil {
		if errors.Is(err, errRepoTelephonyNotFound) {
			return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.AttachTelephonyToCompany/telephony_not_in_db")
		}
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.AttachTelephonyToCompany/getTelephonyByName").SetError(err.Error())
	}
	existing, err := s.repo.getCompanyTelephonyByCompanyIDAndTelephonyID(ctx, tx, companyID, telephony.ID)
	if err != nil && !errors.Is(err, errRepoCompanyTelephonyNotFound) {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.AttachTelephonyToCompany/check_duplicate").SetError(err.Error())
	}
	if existing != nil {
		return nil, srverr.NewServerError(ServiceErrorCompanyTelephonyAlreadyExists, "company.AttachTelephonyToCompany/duplicate")
	}
	ct := &domain.CompanyTelephone{
		ID:                uuid.New(),
		CompanyID:         companyID,
		Telephone:         telephony,
		ExternalAccountID: externalAccountID,
	}
	if err := s.repo.createCompanyTelephony(ctx, tx, ct); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.AttachTelephonyToCompany/create").SetError(err.Error())
	}
	if err := s.pool.Commit(ctx, tx); err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.AttachTelephonyToCompany/commit").SetError(err.Error())
	}
	return ct, nil
}

func (s *service) DetachTelephonyFromCompany(ctx context.Context, companyID, companyTelephonyID uuid.UUID) srverr.ServerError {
	if companyID == uuid.Nil || companyTelephonyID == uuid.Nil {
		return srverr.NewServerError(ServiceErrorCompanyTelephonyBadRequest, "company.DetachTelephonyFromCompany/empty_ids")
	}
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "company.DetachTelephonyFromCompany/begin").SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)
	err = s.repo.deleteCompanyTelephony(ctx, tx, companyTelephonyID, companyID)
	if err != nil {
		if errors.Is(err, errRepoCompanyTelephonyNotFound) {
			return srverr.NewServerError(ServiceErrorCompanyTelephonyNotFound, "company.DetachTelephonyFromCompany")
		}
		return srverr.NewServerError(srverr.ErrInternalServerError, "company.DetachTelephonyFromCompany").SetError(err.Error())
	}
	if err := s.pool.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "company.DetachTelephonyFromCompany/commit").SetError(err.Error())
	}
	return nil
}

func (s *service) ListTelephonyDictionary(ctx context.Context) ([]*domain.Telephony, srverr.ServerError) {
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.ListTelephonyDictionary/begin").SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)
	list, err := s.repo.listTelephony(ctx, tx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "company.ListTelephonyDictionary").SetError(err.Error())
	}
	return list, nil
}
