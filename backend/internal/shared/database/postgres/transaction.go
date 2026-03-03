package postgres

import (
	"context"
	"telephony/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transaction struct {
	cfg  config.PostgresConfig
	pool *pgxpool.Pool
}

func NewTransactionsRepos(cfg config.PostgresConfig, pool *pgxpool.Pool) Transaction {
	return &transaction{cfg: cfg, pool: pool}
}

func (r *transaction) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *transaction) Commit(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

func (r *transaction) Rollback(ctx context.Context, tx pgx.Tx) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.RequestTimeout)
	defer cancel()

	return tx.Rollback(ctx)
}

type testTransaction struct {
	cfg  config.PostgresConfig
	pool *pgxpool.Pool
}

func NewTestTransactionsRepos(cfg config.PostgresConfig, pool *pgxpool.Pool) Transaction {
	return &testTransaction{cfg: cfg, pool: pool}
}

func (r *testTransaction) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *testTransaction) Rollback(ctx context.Context, tx pgx.Tx) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.RequestTimeout)
	defer cancel()

	return tx.Rollback(ctx)
}

func (r *testTransaction) Commit(ctx context.Context, tx pgx.Tx) error {
	return r.Rollback(ctx, tx)
}
