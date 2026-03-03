package countries

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"

	"github.com/jackc/pgx/v5"
)

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

var (
	errRepoCountryNotFound = errors.New("country not found")
)

func (r *repositoryImpl) getCountryByFullName(ctx context.Context, tx pgx.Tx, fullName string) (*domain.Country, error) {
	const query = `SELECT id, code, name, description, created_at, updated_at FROM country WHERE  name ILIKE $1 OR description ILIKE $1 LIMIT 1`

	var country domain.Country
	row := tx.QueryRow(ctx, query, fullName)
	err := row.Scan(&country.ID, &country.Code, &country.Name, &country.Description, &country.CreatedAt, &country.UpdatedAt)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCountryNotFound
		}
		return nil, fmt.Errorf("repository.GetCountryByFullName/Scan: %w", err)
	}
	return &country, nil
}

func (r *repositoryImpl) getCountryByCode(ctx context.Context, tx pgx.Tx, code string) (*domain.Country, error) {
	const query = `SELECT id, code, name, description, created_at, updated_at FROM country WHERE code = $1`
	var country domain.Country
	row := tx.QueryRow(ctx, query, code)
	err := row.Scan(&country.ID, &country.Code, &country.Name, &country.Description, &country.CreatedAt, &country.UpdatedAt)
	if err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoCountryNotFound
		}
		return nil, fmt.Errorf("repository.GetCountryByCode/Scan: %w", err)
	}
	return &country, nil
}

func (r *repositoryImpl) saveUpdateCountries(
	ctx context.Context,
	tx pgx.Tx,
	countries []*domain.Country,
) error {

	if len(countries) == 0 {
		return nil
	}

	const query = `
		INSERT INTO country (code, name, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (code) DO UPDATE
		SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = NOW()
		WHERE
			country.name IS DISTINCT FROM EXCLUDED.name
			OR country.description IS DISTINCT FROM EXCLUDED.description;
	`

	batch := &pgx.Batch{}

	for _, c := range countries {
		batch.Queue(
			query,
			c.Code,
			c.Name,
			c.Description,
		)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range countries {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("repository.saveUpdateCountries/Exec: %w", err)
		}
	}
	return nil
}
