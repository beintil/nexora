package telephony_ingestion_pipeline

import (
	"context"

	"telephony/internal/domain"
	"telephony/internal/modules/call"
	"telephony/internal/modules/company"
	"telephony/internal/modules/countries"
	"telephony/internal/modules/telephony/provider"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"

	"github.com/jackc/pgx/v5"
)

type service struct {
	countryService countries.Service
	callService    call.Service
	companyService company.Service

	pool postgres.Transaction
}

func NewService(
	countryService countries.Service,
	callService call.Service,
	companyService company.Service,

	pool postgres.Transaction,
) Service {
	return &service{
		countryService: countryService,
		callService:    callService,
		companyService: companyService,

		pool: pool,
	}
}

const (
	ServiceErrorCompanyTelephonyNotFound srverr.ErrorTypeNotFound = "company_telephony_not_found"

	ServiceErrorCallIsNotValid srverr.ErrorTypeBadRequest = "call_is_not_valid"
)

func (s *service) CallWorker(ctx context.Context, call *domain.CallWorker, telephony domain.TelephonyName) srverr.ServerError {
	if call == nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/nil_call")
	}
	if call.TelephonyAccountID == "" {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "telephony_ingestion_pipeline_.CallWorker/empty_telephony_account_id")
	}
	if telephony == "" {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "telephony_ingestion_pipeline_.CallWorker/empty_telephony_name")
	}
	if call.Event == nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "telephony_ingestion_pipeline_.CallWorker/nil_event")
	}
	if call.Details == nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "telephony_ingestion_pipeline_.CallWorker/nil_details")
	}

	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "telephony_ingestion_pipeline_.CallWorker/BeginTransaction").
			SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	// Проверяем, существует ли такая компания с такой телефонной системой
	companyTelephony, sErr := s.companyService.GetCompanyTelephonyByExternalAccountIDAndTelephonyNameWithTx(ctx, tx, call.TelephonyAccountID, telephony)
	if sErr != nil {
		if sErr.GetServerError() == company.ServiceErrorCompanyTelephonyNotFound {
			return srverr.NewServerError(ServiceErrorCompanyTelephonyNotFound, "telephony_ingestion_pipeline_.CallWorkerWithTx/GetCompanyTelephonyByExternalIDAndTelephonyName")
		}
		return sErr
	}
	call.CompanyTelephonyID = companyTelephony.ID

	//ok, sErr := s.planService.CheckLimit(ctx, companyTelephony.CompanyID, domain.PlanLimitCallsPerMonth, 0, 1)
	//if sErr != nil {
	//	return sErr
	//}
	//if !ok {
	//	return srverr.NewServerError(plan.ServiceErrorLimitExceeded, "telephony_ingestion_pipeline_.CallWorker/plan_limit_exceeded")
	//}

	// Нормализуем страну
	call.Details.FromCountry, sErr = s.normalizeCountryCodeOrDefaultTH(ctx, tx, call.Details.FromCountry)
	if sErr != nil {
		return sErr
	}
	call.Details.ToCountry, sErr = s.normalizeCountryCodeOrDefaultTH(ctx, tx, call.Details.ToCountry)
	if sErr != nil {
		return sErr
	}

	// Производим флоу сохранения звонка, либо его события
	sErr = s.callService.SaveUpdateCallWithTX(ctx, tx, call)
	if sErr != nil {
		return sErr
	}
	err = s.pool.Commit(ctx, tx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "telephony_ingestion_pipeline_.CallWorkerWithTx/Commit").
			SetError(err.Error())
	}
	return nil
}

// HandleWebhookEvent — На основе нормализованного события формирует CallWorker
func (s *service) HandleWebhookEvent(ctx context.Context, event *provider.CallWebhookEvent) srverr.ServerError {
	if event == nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "telephony_ingestion_pipeline_.HandleWebhookEvent/nil_event")
	}

	callWorker := &domain.CallWorker{
		Call: &domain.Call{
			ExternalParentCallID: event.ExternalParentCallID,
			ExternalCallID:       event.ExternalCallID,
			FromNumber:           event.FromNumber,
			ToNumber:             event.ToNumber,
			Direction:            event.Direction,
			Details: &domain.CallDetails{
				RecordingSid:      event.RecordingID,
				RecordingURL:      event.RecordingURL,
				RecordingDuration: event.RecordingDurationSecond,
				FromCountry:       event.FromCountry,
				FromCity:          event.FromCity,
				ToCountry:         event.ToCountry,
				ToCity:            event.ToCity,
				Carrier:           event.Carrier,
				Trunk:             event.Trunk,
			},
		},
		Event: &domain.CallEvent{
			Status:    event.Status,
			Timestamp: event.OccurredAt,
		},
		TelephonyAccountID: event.TelephonyAccountID,
	}

	return s.CallWorker(ctx, callWorker, event.TelephonyName)
}

func (s *service) normalizeCountryCodeOrDefaultTH(ctx context.Context, tx pgx.Tx, country string) (string, srverr.ServerError) {
	const defaultCountryCode = "TH"

	// Если страна пришла с фулл найм, то проверяем есть ли такая страна в базе
	if len(country) != 2 {
		c, sErr := s.countryService.GetCountryByFullNameWithTx(ctx, tx, country)
		if sErr != nil && sErr.GetServerError() != countries.ServiceErrorCountryNotFound {
			return "", sErr
		}
		if sErr != nil {
			// Если страна не найдена, то присваиваем ей Таиланд
			return defaultCountryCode, nil
		}
		// Если страна найдена, то присваиваем ей код страны
		return c.Code, nil
	}

	// Если страна пришла с кодом страны, то проверяем есть ли такая страна в базе
	_, sErr := s.countryService.GetCountryByCodeWithTx(ctx, tx, country)
	if sErr != nil && sErr.GetServerError() != countries.ServiceErrorCountryNotFound {
		return "", sErr
	}
	if sErr != nil {
		// Если страна не найдена, то присваиваем ей Таиланд
		return defaultCountryCode, nil
	}
	return country, nil
}
