package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Transaction interface {
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
	Rollback(ctx context.Context, tx pgx.Tx) error
	Commit(ctx context.Context, tx pgx.Tx) error
}
