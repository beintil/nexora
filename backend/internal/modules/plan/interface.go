package plan

type Service interface {
	//// GetActivePlanByCompanyID возвращает активный тариф компании и его лимиты.
	//GetActivePlanByCompanyID(ctx context.Context, companyID uuid.UUID) (*domain.CompanyPlan, *domain.Plan, []*domain.PlanLimit, srverr.ServerError)
	//
	//// CheckLimit проверяет, можно ли выполнить операцию с требуемым количеством единиц.
	//CheckLimit(ctx context.Context, companyID uuid.UUID, key domain.PlanLimitKey, currentUsage, requiredAmount int64) (bool, srverr.ServerError)
	//
	//// IncrementUsageWithTx увеличивает usage по лимиту в рамках уже открытой транзакции.
	//IncrementUsageWithTx(ctx context.Context, tx pgx.Tx, companyID uuid.UUID, key domain.PlanLimitKey, delta int64) srverr.ServerError
}
