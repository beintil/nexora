package plan

import (
	"context"
	"os"
	"testing"
	"time"

	"telephony/internal/config"
	"telephony/internal/domain"
	plan2 "telephony/internal/modules/plan"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"
	"telephony/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testEnv struct {
	ctx         context.Context
	pool        *pgxpool.Pool
	transaction postgres.Transaction
	planService plan2.Service
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctx := context.Background()

	log, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	t.Cleanup(func() {
		_ = log.Sync()
	})

	os.Setenv("TEST", "true")
	os.Setenv("CONFIG_PATH", "/Users/darkness/nexora/backend/configs/local.json")

	cfg := config.MustConfig(log)

	pool, err := postgres.New(ctx, cfg.Postgres, log)
	if err != nil {
		t.Fatalf("failed to init postgres: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	transaction := postgres.NewTransactionsRepos(cfg.Postgres, pool)

	repos := plan2.NewRepository()
	planService := plan2.NewService(repos, transaction)

	return &testEnv{
		ctx:         ctx,
		pool:        pool,
		transaction: transaction,
		planService: planService,
	}
}

func (e *testEnv) createCompany(t *testing.T, name string) uuid.UUID {
	t.Helper()

	tx, err := e.pool.Begin(e.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(e.ctx)

	var id uuid.UUID
	err = tx.QueryRow(e.ctx, `INSERT INTO company (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert company: %v", err)
	}

	if err := tx.Commit(e.ctx); err != nil {
		t.Fatalf("failed to commit company insert: %v", err)
	}
	return id
}

func (e *testEnv) getPlanIDBySlug(t *testing.T, slug string) uuid.UUID {
	t.Helper()

	tx, err := e.pool.Begin(e.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(e.ctx)

	var id uuid.UUID
	err = tx.QueryRow(e.ctx, `SELECT id FROM plan WHERE slug = $1 LIMIT 1`, slug).Scan(&id)
	if err != nil {
		t.Fatalf("failed to select plan by slug=%s: %v", slug, err)
	}
	return id
}

func (e *testEnv) assignPlanToCompany(t *testing.T, companyID, planID uuid.UUID) {
	t.Helper()

	tx, err := e.pool.Begin(e.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(e.ctx)

	if _, err := tx.Exec(e.ctx, `DELETE FROM company_plan WHERE company_id = $1`, companyID); err != nil {
		t.Fatalf("failed to cleanup company_plan: %v", err)
	}

	_, err = tx.Exec(
		e.ctx,
		`INSERT INTO company_plan (company_id, plan_id, started_at, ends_at) VALUES ($1, $2, now(), NULL)`,
		companyID,
		planID,
	)
	if err != nil {
		t.Fatalf("failed to insert company_plan: %v", err)
	}

	if err := tx.Commit(e.ctx); err != nil {
		t.Fatalf("failed to commit company_plan insert: %v", err)
	}
}

func (e *testEnv) setPlanLimitValue(t *testing.T, planID uuid.UUID, key domain.PlanLimitKey, value int64) {
	t.Helper()

	tx, err := e.pool.Begin(e.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(e.ctx)

	_, err = tx.Exec(
		e.ctx,
		`UPDATE plan_limit SET value = $3 WHERE plan_id = $1 AND limit_type = $2`,
		planID,
		string(key),
		value,
	)
	if err != nil {
		t.Fatalf("failed to update plan_limit: %v", err)
	}

	if err := tx.Commit(e.ctx); err != nil {
		t.Fatalf("failed to commit plan_limit update: %v", err)
	}
}

func (e *testEnv) getUsageCount(t *testing.T, companyID uuid.UUID, key domain.PlanLimitKey) int64 {
	t.Helper()

	tx, err := e.pool.Begin(e.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(e.ctx)

	var count int64
	err = tx.QueryRow(
		e.ctx,
		`SELECT COUNT(*) FROM plan_usage WHERE company_id = $1 AND limit_type = $2`,
		companyID,
		string(key),
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count plan_usage: %v", err)
	}
	return count
}

func TestPlanService_GetActivePlanByCompanyID_Success(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_get_active_success")
	planID := env.getPlanIDBySlug(t, "basic")
	env.assignPlanToCompany(t, companyID, planID)

	cp, p, limits, sErr := env.planService.GetActivePlanByCompanyID(env.ctx, companyID)
	if sErr != nil {
		t.Fatalf("expected no error, got: %v", sErr.Error())
	}
	if cp == nil || p == nil {
		t.Fatalf("expected non-nil company plan and plan")
	}
	if cp.CompanyID != companyID {
		t.Fatalf("unexpected company_id in company_plan: %s", cp.CompanyID)
	}
	if cp.PlanID != planID {
		t.Fatalf("unexpected plan_id in company_plan: %s", cp.PlanID)
	}
	if !p.IsActive {
		t.Fatalf("expected active plan, got inactive")
	}
	if p.Slug != "basic" && p.Slug != "pro" {
		t.Fatalf("unexpected plan slug: %s", p.Slug)
	}
	if len(limits) == 0 {
		t.Fatalf("expected non-empty plan limits")
	}

	foundBasicAccess := false
	for _, l := range limits {
		if l.LimitType == domain.PlanLimitMetricsBasicAccess {
			foundBasicAccess = true
			break
		}
	}
	if !foundBasicAccess {
		t.Fatalf("expected PlanLimitMetricsBasicAccess in limits")
	}
}

func TestPlanService_GetActivePlanByCompanyID_NoPlan(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_no_plan")

	cp, p, limits, sErr := env.planService.GetActivePlanByCompanyID(env.ctx, companyID)
	if sErr == nil {
		t.Fatalf("expected error, got nil")
	}
	if sErr.GetServerError() != plan2.ServiceErrorPlanNotAssigned {
		t.Fatalf("unexpected error type: %s", sErr.GetServerError().String())
	}
	if cp != nil || p != nil || len(limits) != 0 {
		t.Fatalf("expected nil results when no plan assigned")
	}
}

func TestPlanService_GetActivePlanByCompanyID_EmptyCompanyID(t *testing.T) {
	env := newTestEnv(t)

	_, _, _, sErr := env.planService.GetActivePlanByCompanyID(env.ctx, uuid.Nil)
	if sErr == nil {
		t.Fatalf("expected error, got nil")
	}
	if sErr.GetServerError() != srverr.ErrInternalServerError {
		t.Fatalf("unexpected error type: %s", sErr.GetServerError().String())
	}
}

func TestPlanService_CheckLimit_NoLimitRecord(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_no_limit_record")
	planID := env.getPlanIDBySlug(t, "basic")
	env.assignPlanToCompany(t, companyID, planID)

	key := domain.PlanLimitKey("test.limit.no_record")

	ok, sErr := env.planService.CheckLimit(env.ctx, companyID, key, 0, 1)
	if sErr != nil {
		t.Fatalf("expected no error, got: %v", sErr.Error())
	}
	if !ok {
		t.Fatalf("expected ok=true when no limit record exists")
	}
}

func TestPlanService_CheckLimit_EnoughLimit(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_enough_limit")
	planID := env.getPlanIDBySlug(t, "basic")
	env.assignPlanToCompany(t, companyID, planID)

	env.setPlanLimitValue(t, planID, domain.PlanLimitMetricsBasicAccess, 5)

	ok, sErr := env.planService.CheckLimit(env.ctx, companyID, domain.PlanLimitMetricsBasicAccess, 2, 3)
	if sErr != nil {
		t.Fatalf("expected no error, got: %v", sErr.Error())
	}
	if !ok {
		t.Fatalf("expected ok=true for total usage within limit")
	}
}

func TestPlanService_CheckLimit_LimitExceeded(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_limit_exceeded")
	planID := env.getPlanIDBySlug(t, "basic")
	env.assignPlanToCompany(t, companyID, planID)

	env.setPlanLimitValue(t, planID, domain.PlanLimitMetricsBasicAccess, 4)

	ok, sErr := env.planService.CheckLimit(env.ctx, companyID, domain.PlanLimitMetricsBasicAccess, 3, 2)
	if sErr == nil {
		t.Fatalf("expected error, got nil")
	}
	if ok {
		t.Fatalf("expected ok=false when limit exceeded")
	}
	if sErr.GetServerError() != plan2.ServiceErrorLimitExceeded {
		t.Fatalf("unexpected error type: %s", sErr.GetServerError().String())
	}
}

func TestPlanService_CheckLimit_NoPlanAssigned(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_no_plan_for_check_limit")

	ok, sErr := env.planService.CheckLimit(env.ctx, companyID, domain.PlanLimitMetricsBasicAccess, 0, 1)
	if sErr == nil {
		t.Fatalf("expected error, got nil")
	}
	if ok {
		t.Fatalf("expected ok=false when no plan assigned")
	}
	if sErr.GetServerError() != plan2.ServiceErrorPlanNotAssigned {
		t.Fatalf("unexpected error type: %s", sErr.GetServerError().String())
	}
}

func TestPlanService_CheckLimit_ZeroRequiredAmount(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_zero_required")
	planID := env.getPlanIDBySlug(t, "basic")
	env.assignPlanToCompany(t, companyID, planID)

	ok, sErr := env.planService.CheckLimit(env.ctx, companyID, domain.PlanLimitMetricsBasicAccess, 10, 0)
	if sErr != nil {
		t.Fatalf("expected no error, got: %v", sErr.Error())
	}
	if !ok {
		t.Fatalf("expected ok=true when requiredAmount is zero")
	}
}

func TestPlanService_IncrementUsageWithTx_CreateAndUpdate(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_usage_create_update")

	tx, err := env.transaction.BeginTransaction(env.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer env.transaction.Rollback(env.ctx, tx)

	key := domain.PlanLimitSmsPerMonth

	sErr := env.planService.IncrementUsageWithTx(env.ctx, tx, companyID, key, 5)
	if sErr != nil {
		t.Fatalf("expected no error on first increment, got: %v", sErr.Error())
	}

	var value1 int64
	var periodStart1 time.Time
	err = tx.QueryRow(
		env.ctx,
		`SELECT value, period_start FROM plan_usage WHERE company_id = $1 AND limit_type = $2`,
		companyID,
		string(key),
	).Scan(&value1, &periodStart1)
	if err != nil {
		t.Fatalf("failed to select plan_usage after first increment: %v", err)
	}
	if value1 != 5 {
		t.Fatalf("expected value=5 after first increment, got %d", value1)
	}

	now := time.Now().UTC()
	expectedPeriodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if periodStart1.Year() != expectedPeriodStart.Year() ||
		periodStart1.Month() != expectedPeriodStart.Month() ||
		periodStart1.Day() != expectedPeriodStart.Day() {
		t.Fatalf("unexpected period_start: %v, expected year=%d month=%d day=%d",
			periodStart1, expectedPeriodStart.Year(), expectedPeriodStart.Month(), expectedPeriodStart.Day())
	}

	sErr = env.planService.IncrementUsageWithTx(env.ctx, tx, companyID, key, 2)
	if sErr != nil {
		t.Fatalf("expected no error on second increment, got: %v", sErr.Error())
	}

	var value2 int64
	var periodStart2 time.Time
	err = tx.QueryRow(
		env.ctx,
		`SELECT value, period_start FROM plan_usage WHERE company_id = $1 AND limit_type = $2`,
		companyID,
		string(key),
	).Scan(&value2, &periodStart2)
	if err != nil {
		t.Fatalf("failed to select plan_usage after second increment: %v", err)
	}
	if value2 != 7 {
		t.Fatalf("expected value=7 after second increment, got %d", value2)
	}
	if !periodStart2.Equal(periodStart1) {
		t.Fatalf("expected same period_start after second increment, got %v vs %v", periodStart2, periodStart1)
	}
}

func TestPlanService_IncrementUsageWithTx_ZeroDelta(t *testing.T) {
	env := newTestEnv(t)

	companyID := env.createCompany(t, "plan_test_company_zero_delta")

	tx, err := env.transaction.BeginTransaction(env.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer env.transaction.Rollback(env.ctx, tx)

	key := domain.PlanLimitEmailPerMonth

	sErr := env.planService.IncrementUsageWithTx(env.ctx, tx, companyID, key, 0)
	if sErr != nil {
		t.Fatalf("expected no error on zero delta, got: %v", sErr.Error())
	}

	count := env.getUsageCount(t, companyID, key)
	if count != 0 {
		t.Fatalf("expected no usage records for zero delta, got %d", count)
	}
}

func TestPlanService_IncrementUsageWithTx_EmptyCompanyID(t *testing.T) {
	env := newTestEnv(t)

	tx, err := env.transaction.BeginTransaction(env.ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer env.transaction.Rollback(env.ctx, tx)

	sErr := env.planService.IncrementUsageWithTx(env.ctx, tx, uuid.Nil, domain.PlanLimitSmsPerMonth, 1)
	if sErr == nil {
		t.Fatalf("expected error, got nil")
	}
	if sErr.GetServerError() != srverr.ErrInternalServerError {
		t.Fatalf("unexpected error type: %s", sErr.GetServerError().String())
	}
}
