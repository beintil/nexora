package company

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	errRepoTelephonyNotFound = errors.New("telephony not found")
	errRepoCompanyNotFound   = errors.New("company not found")
)

// repository — неэкспортируемый интерфейс (методы с маленькой буквы), только внутри модуля.
type repository interface {
	getTelephonyByName(ctx context.Context, tx pgx.Tx, name domain.TelephonyName) (*domain.Telephony, error)
	getCompanyTelephonyByExternalAccountIDAndTelephonyName(ctx context.Context, tx pgx.Tx, externalID string, telephonyName domain.TelephonyName) (*domain.CompanyTelephone, error)
	getCompanyByName(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, error)
	getCompanyByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, error)
	createCompany(ctx context.Context, tx pgx.Tx, company *domain.Company) error
}

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) getTelephonyByName(
	ctx context.Context,
	tx pgx.Tx,
	name domain.TelephonyName,
) (*domain.Telephony, error) {
	var query = `SELECT id FROM telephony WHERE name = $1`
	row := tx.QueryRow(ctx, query, name)

	var telephony domain.Telephony

	err := row.Scan(&telephony.ID)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoTelephonyNotFound
		}
		return nil, fmt.Errorf("company.GetTelephonyByName: %w", err)
	}
	return &telephony, nil
}

func (r *repositoryImpl) getCompanyTelephonyByExternalAccountIDAndTelephonyName(
	ctx context.Context,
	tx pgx.Tx,
	externalID string,
	telephonyName domain.TelephonyName,
) (*domain.CompanyTelephone, error) {
	var query = `
	SELECT 
	 ct.id, 
	 ct.company_id, 
	 ct.telephony_id,
	 ct.external_account_id,
	 t.name,
	 ct.created_at,
	 ct.updated_at
	FROM company_telephony ct
		JOIN telephony t ON t.name = $2
	WHERE ct.external_account_id = $1 AND ct.telephony_id = t.id 
`
	row := tx.QueryRow(ctx, query, externalID, telephonyName)
	var companyTelephony = domain.CompanyTelephone{
		Telephone: &domain.Telephony{},
	}
	err := row.Scan(
		&companyTelephony.ID,
		&companyTelephony.CompanyID,
		&companyTelephony.Telephone.ID,
		&companyTelephony.ExternalAccountID,
		&companyTelephony.Telephone.Name,
		&companyTelephony.CreatedAt,
		&companyTelephony.UpdatedAt,
	)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoTelephonyNotFound
		}
		return nil, fmt.Errorf("company.GetCompanyTelephonyByExternalIDAndTelephonyName: %w", err)
	}

	return &companyTelephony, nil
}

func (r *repositoryImpl) getCompanyByName(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, error) {
	const query = `SELECT id, name, created_at, updated_at FROM company WHERE name = $1`
	row := tx.QueryRow(ctx, query, name)
	var c domain.Company
	err := row.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCompanyNotFound
		}
		return nil, fmt.Errorf("company.GetCompanyByName: %w", err)
	}
	return &c, nil
}

func (r *repositoryImpl) getCompanyByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, error) {
	const query = `SELECT id, name, created_at, updated_at FROM company WHERE id = $1`
	row := tx.QueryRow(ctx, query, id)
	var c domain.Company
	err := row.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCompanyNotFound
		}
		return nil, fmt.Errorf("company.GetCompanyByID: %w", err)
	}
	return &c, nil
}

func (r *repositoryImpl) createCompany(ctx context.Context, tx pgx.Tx, company *domain.Company) error {
	const query = `INSERT INTO company (id, name) VALUES ($1, $2) RETURNING created_at, updated_at`
	err := tx.QueryRow(ctx, query, company.ID, company.Name).Scan(&company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("company.CreateCompany: %w", err)
	}
	return nil
}
