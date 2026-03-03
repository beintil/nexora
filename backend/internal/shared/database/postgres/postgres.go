package postgres

import (
	"context"
	"fmt"
	"telephony/internal/config"
	"telephony/pkg/logger"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cfg config.PostgresConfig, log logger.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s pool_max_conns=10",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

	if cfg.IsDebug {
		conf.ConnConfig.Tracer = &pgxTracer{
			log: log,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(ctxTimeout); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	log.Info("Connected to PostgreSQL Successfully")

	return pool, nil
}
