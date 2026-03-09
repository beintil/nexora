package plan

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository interface {
	getActiveCompanyPlans(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) ([]*domain.CompanyPlan, error)
	getPlansByIDs(ctx context.Context, tx pgx.Tx, planIDs []uuid.UUID) ([]*domain.Plan, error)
	getPlansUsagesByCompanyPlanIDs(ctx context.Context, tx pgx.Tx, companyPlanID []uuid.UUID) ([]*domain.PlanUsage, error)
	getActivePlansUsagesOverByCompanyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) ([]*domain.PlanUsageOver, error)
}

var (
	errRepoCompanyPlanNotFound = errors.New("company plan not found")
	errRepoPlanNotFound        = errors.New("plan not found")
	errRepoPlanUsageNotFound   = errors.New("plan usage not found")
)

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) getActiveCompanyPlans(
	ctx context.Context,
	tx pgx.Tx,
	companyID uuid.UUID,
) ([]*domain.CompanyPlan, error) {
	const query = `
	SELECT 
	    id,
	    company_id,
	    plan_id,
	    started_at,
	    ends_at
	FROM company_plan
	WHERE company_id = $1 AND is_active
`
	rows, err := tx.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("plan.getCompanyPlans: %w", err)
	}
	defer rows.Close()

	var companyPlans []*domain.CompanyPlan
	for rows.Next() {
		var cp domain.CompanyPlan
		err = rows.Scan(
			&cp.ID,
			&cp.CompanyID,
			&cp.PlanID,
			&cp.StartedAt,
			&cp.EndsAt,
		)
		if err != nil {
			return nil, fmt.Errorf("plan.getCompanyPlans/scan: %w", err)
		}
		companyPlans = append(companyPlans, &cp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan.getCompanyPlans/rows: %w", err)
	}
	return companyPlans, nil
}

func (r *repositoryImpl) getPlansByIDs(ctx context.Context, tx pgx.Tx, planIDs []uuid.UUID) ([]*domain.Plan, error) {
	const query = `
	SELECT  
		p.id,
		p.name,
		p.slug,
		p.is_active,
		p.sort_order, 
		p.description,    
	
		pl.plan_id,
	    pl.limit_type,
	    pl.value,
	    pl.extra
	FROM plan p
		JOIN plan_limit pl ON pl.plan_id = p.id
	WHERE p.id = ANY($1)
`
	rows, err := tx.Query(ctx, query, planIDs)
	if err != nil {
		return nil, fmt.Errorf("plan.getPlansByIDs/Query: %w", err)
	}
	defer rows.Close()
	var plansMap = make(map[uuid.UUID]*domain.Plan)
	for rows.Next() {
		var p domain.Plan
		var pl domain.PlanLimit
		err = rows.Scan(
			&p.ID,
			&p.Name,
			&p.Slug,
			&p.IsActive,
			&p.SortOrder,
			&p.Description,

			&pl.PlanID,
			&pl.LimitType,
			&pl.Value,
			&pl.Extra,
		)
		if err != nil {
			return nil, fmt.Errorf("plan.getPlansByIDs/scan: %w", err)
		}
		if _, ok := plansMap[p.ID]; !ok {
			plansMap[p.ID] = &p
		}
		p.PlanLimits = append(p.PlanLimits, &pl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan.getPlansByIDs/rows: %w", err)
	}

	var plans []*domain.Plan
	for _, p := range plansMap {
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *repositoryImpl) getPlansUsagesByCompanyPlanIDs(ctx context.Context, tx pgx.Tx, companyPlanID []uuid.UUID) ([]*domain.PlanUsage, error) {
	const query = `
	SELECT
	 company_plan_id, limit_type, value    
	FROM plan_usage
	WHERE company_plan_id = ANY($1)
`
	rows, err := tx.Query(ctx, query, companyPlanID)
	if err != nil {
		return nil, fmt.Errorf("plan.getPlansUsagesByCompanyPlanIDs/Query: %w", err)
	}
	defer rows.Close()

	var planUsages []*domain.PlanUsage
	for rows.Next() {
		pu := domain.PlanUsage{}
		err = rows.Scan(&pu.CompanyPlanID, &pu.LimitType, &pu.Value)
		if err != nil {
			return nil, fmt.Errorf("plan.getPlansUsagesByCompanyPlanIDs/Scan: %w", err)
		}
		planUsages = append(planUsages, &pu)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan.getPlansUsagesByCompanyPlanIDs/Rows: %w", err)
	}
	return planUsages, nil
}

func (r *repositoryImpl) getActivePlansUsagesOverByCompanyID(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) ([]*domain.PlanUsageOver, error) {
	const query = `
	SELECT
	    company_id, limit_type, value, period_start, period_end, is_active
	FROM plan_usage_over
	WHERE company_id = $1 AND is_active
`
	rows, err := tx.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("plan.getActivePlansUsagesOverByCompanyID/Query: %w", err)
	}
	defer rows.Close()

	var planUsages []*domain.PlanUsageOver
	for rows.Next() {
		pu := domain.PlanUsageOver{}
		err = rows.Scan(
			&pu.CompanyID,
			&pu.LimitType,
			&pu.Value,
			&pu.StartPeriod,
			&pu.EndPeriod,
			&pu.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("plan.getActivePlansUsagesOverByCompanyID/Scan: %w", err)
		}
		planUsages = append(planUsages, &pu)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan.getActivePlansUsagesOverByCompanyID/Rows: %w", err)
	}
	return planUsages, nil
}
