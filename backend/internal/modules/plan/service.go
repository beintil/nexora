package plan

import (
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"
)

const (
	ServiceErrorPlanNotAssigned srverr.ErrorTypeForbidden = "plan_not_assigned"
	ServiceErrorLimitExceeded   srverr.ErrorTypeForbidden = "plan_limit_exceeded"
)

type service struct {
	repos repository
	pool  postgres.Transaction
}

func NewService(
	repos repository,
	pool postgres.Transaction,
) Service {
	return &service{
		repos: repos,
		pool:  pool,
	}
}

//
//func (s *service) GetActivePlansByCompanyID(
//	ctx context.Context,
//	companyID uuid.UUID,
//) (*domain.CompanyPlans, srverr.ServerError) {
//	if companyID == uuid.Nil {
//		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlansByCompanyID/empty_company_id")
//	}
//	now := time.Now().UTC()
//
//	tx, err := s.pool.BeginTransaction(ctx)
//	if err != nil {
//		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlansByCompanyID/begin").
//			SetError(err.Error())
//	}
//	defer s.pool.Rollback(ctx, tx)
//
//}
//
//func (s *service) GetActivePlanByCompanyID(
//	ctx context.Context,
//	companyID uuid.UUID,
//) (*domain.CompanyPlan, *domain.Plan, []*domain.PlanLimit, srverr.ServerError) {
//	if companyID == uuid.Nil {
//		return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/empty_company_id")
//	}
//
//	tx, err := s.pool.BeginTransaction(ctx)
//	if err != nil {
//		return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/begin").
//			SetError(err.Error())
//	}
//	defer s.pool.Rollback(ctx, tx)
//
//	cp, err := s.repos.getActiveCompanyPlan(ctx, tx, companyID)
//	if err != nil {
//		if errors.Is(err, errRepoCompanyPlanNotFound) {
//			return nil, nil, nil, srverr.NewServerError(ServiceErrorPlanNotAssigned, "plan.GetActivePlanByCompanyID/getActiveCompanyPlan")
//		}
//		return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/getActiveCompanyPlan").
//			SetError(err.Error())
//	}
//
//	p, err := s.repos.getPlanByID(ctx, tx, cp.PlanID)
//	if err != nil {
//		if errors.Is(err, errRepoPlanNotFound) {
//			return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/getPlanByID_not_found")
//		}
//		return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/getPlanByID").
//			SetError(err.Error())
//	}
//
//	limits, err := s.repos.getPlanLimitsByPlanID(ctx, tx, cp.PlanID)
//	if err != nil {
//		return nil, nil, nil, srverr.NewServerError(srverr.ErrInternalServerError, "plan.GetActivePlanByCompanyID/getPlanLimitsByPlanID").
//			SetError(err.Error())
//	}
//
//	return cp, p, limits, nil
//}
//
//func (s *service) CheckLimit(
//	ctx context.Context,
//	companyID uuid.UUID,
//	key domain.PlanLimitKey,
//	currentUsage,
//	requiredAmount int64,
//) (bool, srverr.ServerError) {
//	if companyID == uuid.Nil {
//		return false, srverr.NewServerError(srverr.ErrInternalServerError, "plan.CheckLimit/empty_company_id")
//	}
//	if requiredAmount <= 0 {
//		return true, nil
//	}
//
//	tx, err := s.pool.BeginTransaction(ctx)
//	if err != nil {
//		return false, srverr.NewServerError(srverr.ErrInternalServerError, "plan.CheckLimit/begin").
//			SetError(err.Error())
//	}
//	defer s.pool.Rollback(ctx, tx)
//
//	cp, err := s.repos.getActiveCompanyPlan(ctx, tx, companyID)
//	if err != nil {
//		if errors.Is(err, errRepoCompanyPlanNotFound) {
//			return false, srverr.NewServerError(ServiceErrorPlanNotAssigned, "plan.CheckLimit/getActiveCompanyPlan")
//		}
//		return false, srverr.NewServerError(srverr.ErrInternalServerError, "plan.CheckLimit/getActiveCompanyPlan").
//			SetError(err.Error())
//	}
//
//	limits, err := s.repos.getPlanLimitsByPlanID(ctx, tx, cp.PlanID)
//	if err != nil {
//		return false, srverr.NewServerError(srverr.ErrInternalServerError, "plan.CheckLimit/getPlanLimitsByPlanID").
//			SetError(err.Error())
//	}
//
//	var limitValue *int64
//	for _, l := range limits {
//		if l.LimitType == key {
//			v := l.Value
//			limitValue = &v
//			break
//		}
//	}
//
//	// Если лимит не найден — считаем, что ограничения нет (поведение можно изменить при необходимости).
//	if limitValue == nil {
//		return true, nil
//	}
//
//	total := currentUsage + requiredAmount
//	if total <= *limitValue {
//		return true, nil
//	}
//
//	return false, srverr.NewServerError(ServiceErrorLimitExceeded, "plan.CheckLimit/limit_exceeded")
//}
//
//func (s *service) IncrementUsageWithTx(
//	ctx context.Context,
//	tx pgx.Tx,
//	companyID uuid.UUID,
//	key domain.PlanLimitKey,
//	delta int64,
//) srverr.ServerError {
//	if companyID == uuid.Nil {
//		return srverr.NewServerError(srverr.ErrInternalServerError, "plan.IncrementUsageWithTx/empty_company_id")
//	}
//	if delta == 0 {
//		return nil
//	}
//
//	now := time.Now().UTC()
//	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
//	periodStartStr := periodStart.Format(time.DateOnly)
//
//	existing, err := s.repos.getUsageByPeriod(ctx, tx, companyID, key, periodStartStr)
//	if err != nil && !errors.Is(err, errRepoPlanUsageNotFound) {
//		return srverr.NewServerError(srverr.ErrInternalServerError, "plan.IncrementUsageWithTx/getUsageByPeriod").
//			SetError(err.Error())
//	}
//
//	var newValue int64
//	if existing != nil {
//		newValue = existing.Value + delta
//	} else {
//		newValue = delta
//	}
//
//	usage := &domain.PlanUsage{
//		CompanyID:   companyID,
//		PeriodStart: periodStart,
//		LimitType:   key,
//		Value:       newValue,
//	}
//
//	if err := s.repos.saveUpdateUsageBulk(ctx, tx, []*domain.PlanUsage{usage}); err != nil {
//		return srverr.NewServerError(srverr.ErrInternalServerError, "plan.IncrementUsageWithTx/saveUpdateUsageBulk").
//			SetError(err.Error())
//	}
//	return nil
//}
