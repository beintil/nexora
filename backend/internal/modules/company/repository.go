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
	errRepoTelephonyNotFound        = errors.New("telephony not found")
	errRepoCompanyNotFound          = errors.New("company not found")
	errRepoCompanyTelephonyNotFound = errors.New("company telephony not found")
)

// repository — неэкспортируемый интерфейс (методы с маленькой буквы), только внутри модуля.
type repository interface {
	getTelephonyByName(ctx context.Context, tx pgx.Tx, name domain.TelephonyName) (*domain.Telephony, error)
	getCompanyTelephonyByExternalAccountIDAndTelephonyName(ctx context.Context, tx pgx.Tx, externalID string, telephonyName domain.TelephonyName) (*domain.CompanyTelephone, error)
	getCompanyByName(ctx context.Context, tx pgx.Tx, name string) (*domain.Company, error)
	getCompanyByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Company, error)
	createCompany(ctx context.Context, tx pgx.Tx, company *domain.Company) error

	listCompanyTelephonyByCompanyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) ([]*domain.CompanyTelephone, error)
	createCompanyTelephony(ctx context.Context, tx pgx.Tx, ct *domain.CompanyTelephone) error
	deleteCompanyTelephony(ctx context.Context, tx pgx.Tx, id, companyID uuid.UUID) error
	getCompanyTelephonyByCompanyIDAndTelephonyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, telephonyID int64) (*domain.CompanyTelephone, error)
	listTelephony(ctx context.Context, tx pgx.Tx) ([]*domain.Telephony, error)
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
	var query = `SELECT id, name FROM telephony WHERE name = $1`
	row := tx.QueryRow(ctx, query, name)

	var telephony domain.Telephony

	err := row.Scan(&telephony.ID, &telephony.Name)
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

func (r *repositoryImpl) listCompanyTelephonyByCompanyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) ([]*domain.CompanyTelephone, error) {
	const query = `
	SELECT ct.id, ct.company_id, ct.telephony_id, ct.external_account_id, t.name, ct.created_at, ct.updated_at
	FROM company_telephony ct
	JOIN telephony t ON t.id = ct.telephony_id
	WHERE ct.company_id = $1
	ORDER BY ct.created_at
	`
	rows, err := tx.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("company.listCompanyTelephonyByCompanyID: %w", err)
	}
	defer rows.Close()
	var list []*domain.CompanyTelephone
	for rows.Next() {
		var ct domain.CompanyTelephone
		ct.Telephone = &domain.Telephony{}
		if err := rows.Scan(
			&ct.ID,
			&ct.CompanyID,
			&ct.Telephone.ID,
			&ct.ExternalAccountID,
			&ct.Telephone.Name,
			&ct.CreatedAt,
			&ct.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("company.listCompanyTelephonyByCompanyID/scan: %w", err)
		}
		list = append(list, &ct)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("company.listCompanyTelephonyByCompanyID/rows: %w", err)
	}
	if list == nil {
		list = []*domain.CompanyTelephone{}
	}
	return list, nil
}

func (r *repositoryImpl) createCompanyTelephony(ctx context.Context, tx pgx.Tx, ct *domain.CompanyTelephone) error {
	const query = `INSERT INTO company_telephony (id, company_id, telephony_id, external_account_id) VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	err := tx.QueryRow(ctx, query, ct.ID, ct.CompanyID, ct.Telephone.ID, ct.ExternalAccountID).Scan(&ct.CreatedAt, &ct.UpdatedAt)
	if err != nil {
		return fmt.Errorf("company.createCompanyTelephony: %w", err)
	}
	return nil
}

func (r *repositoryImpl) deleteCompanyTelephony(ctx context.Context, tx pgx.Tx, id, companyID uuid.UUID) error {
	const query = `DELETE FROM company_telephony WHERE id = $1 AND company_id = $2`
	tag, err := tx.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("company.deleteCompanyTelephony: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errRepoCompanyTelephonyNotFound
	}
	return nil
}

func (r *repositoryImpl) getCompanyTelephonyByCompanyIDAndTelephonyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, telephonyID int64) (*domain.CompanyTelephone, error) {
	const query = `
	SELECT ct.id, ct.company_id, ct.telephony_id, ct.external_account_id, t.name, ct.created_at, ct.updated_at
	FROM company_telephony ct
	JOIN telephony t ON t.id = ct.telephony_id
	WHERE ct.company_id = $1 AND ct.telephony_id = $2
	`
	row := tx.QueryRow(ctx, query, companyID, telephonyID)
	var ct domain.CompanyTelephone
	ct.Telephone = &domain.Telephony{}
	err := row.Scan(
		&ct.ID,
		&ct.CompanyID,
		&ct.Telephone.ID,
		&ct.ExternalAccountID,
		&ct.Telephone.Name,
		&ct.CreatedAt,
		&ct.UpdatedAt,
	)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCompanyTelephonyNotFound
		}
		return nil, fmt.Errorf("company.getCompanyTelephonyByCompanyIDAndTelephonyID: %w", err)
	}
	return &ct, nil
}

func (r *repositoryImpl) listTelephony(ctx context.Context, tx pgx.Tx) ([]*domain.Telephony, error) {
	const query = `SELECT id, name FROM telephony ORDER BY name`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("company.listTelephony: %w", err)
	}
	defer rows.Close()
	var list []*domain.Telephony
	for rows.Next() {
		var t domain.Telephony
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("company.listTelephony/scan: %w", err)
		}
		list = append(list, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("company.listTelephony/rows: %w", err)
	}
	if list == nil {
		list = []*domain.Telephony{}
	}
	return list, nil
}
