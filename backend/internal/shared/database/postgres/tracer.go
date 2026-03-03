package postgres

import (
	"context"
	"telephony/pkg/logger"

	"github.com/jackc/pgx/v5"
)

type pgxTracer struct {
	log logger.Logger
}

type ctxKey string

const queryKey ctxKey = "sql"

func (t *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, queryKey, data.SQL)
	t.log.Debugf("[PGX][START] %s | args: %v", data.SQL, data.Args)
	return ctx
}

func (t *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	sql, _ := ctx.Value(queryKey).(string)
	t.log.Debugf("[PGX][END] %s | command: %s | err: %v", sql, data.CommandTag, data.Err)
}
