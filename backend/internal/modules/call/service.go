package call

import (
	"context"
	"errors"
	"strings"
	"telephony/internal/domain"
	"telephony/internal/modules/call_events"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	callEventsService call_events.Service

	repos repository

	pool postgres.Transaction
}

func NewService(
	callEventsService call_events.Service,

	repos repository,

	pool postgres.Transaction,
) Service {
	return &service{
		callEventsService: callEventsService,

		repos: repos,

		pool: pool,
	}
}

const (
	ServiceErrorCallIsNotValid srverr.ErrorTypeBadRequest = "call_is_not_valid"
)

func (s *service) GetCallTreeByCallUUIDWithTx(ctx context.Context, tx pgx.Tx, callUUID uuid.UUID) (*domain.CallTree, srverr.ServerError) {
	if callUUID == uuid.Nil {
		return nil, srverr.NewServerError(ServiceErrorCallIsNotValid, "call.GetCallTreeByCallUUIDWithTx/empty_call_uuid")
	}

	tree, err := s.repos.getCallTreeByCallUUID(ctx, tx, callUUID)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call.GetCallTreeByCallUUIDWithTx/getCallTreeByCallUUID").
			SetError(err.Error())
	}
	callIDs := tree.CallIDs()

	details, err := s.repos.getCallsDetails(ctx, tx, callIDs)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "call.GetCallTreeByCallUUIDWithTx/getCallsDetails").
			SetError(err.Error())
	}
	tree.ApplyDetails(details)

	events, sErr := s.callEventsService.GetEventsByCallIDsWithTx(ctx, tx, callIDs)
	if sErr != nil {
		return nil, sErr
	}
	tree.ApplyEvents(events)

	return tree, nil
}

func (s *service) SaveUpdateCall(ctx context.Context, call *domain.CallWorker) srverr.ServerError {
	tx, err := s.pool.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.SaveCall/BeginTransaction").
			SetError(err.Error())
	}
	defer s.pool.Rollback(ctx, tx)

	sErr := s.SaveUpdateCallWithTX(ctx, tx, call)
	if sErr != nil {
		return sErr
	}

	err = s.pool.Commit(ctx, tx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.SaveCall/Commit").
			SetError(err.Error())
	}
	return nil
}

func (s *service) SaveUpdateCallWithTX(ctx context.Context, tx pgx.Tx, callWorker *domain.CallWorker) srverr.ServerError {
	if callWorker == nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/nil_call")
	}
	if callWorker.ExternalCallID == "" {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/empty_external_call_id")
	}
	if callWorker.FromNumber == "" {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/empty_from_number")
	}
	if callWorker.ToNumber == "" {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/empty_to_number")
	}
	if callWorker.CompanyTelephonyID == uuid.Nil {
		return srverr.NewServerError(ServiceErrorCallIsNotValid, "call.CallWorker/empty_company_telephony_id")
	}

	call, err := s.repos.getCallByCompanyTelephonyIDAndExternalCallID(ctx, tx, callWorker.CompanyTelephonyID, callWorker.ExternalCallID)
	if err != nil && !errors.Is(err, errRepoCallNotFound) {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.SaveCall/getCall").
			SetError(err.Error())
	}
	var callIsExit = err == nil

	// Если звнка еще не существует, то сохраняем его
	if !callIsExit {
		sErr := s.workerCallNotExists(ctx, tx, callWorker)
		if sErr != nil {
			return sErr
		}
		call = callWorker.Call
	}

	// Сохраняем или обновляем детали звонка
	callWorker.Details.CallID = call.ID
	err = s.repos.saveOrUpdateCallDetails(ctx, tx, callWorker.Details)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.SaveCall/saveOrUpdateCallDetails").
			SetError(err.Error())
	}

	// Создаем событие
	var callEvent = callWorker.Event
	callEvent.CallID = call.ID
	callEvent.ID = uuid.New()
	sErr := s.callEventsService.SaveCallEventWithTx(ctx, tx, callEvent)
	if sErr != nil {
		return sErr
	}
	return nil
}

func (s *service) workerCallNotExists(ctx context.Context, tx pgx.Tx, callWorker *domain.CallWorker) srverr.ServerError {
	// Если звонок имеет родителя, то проверяем, существует ли он
	if strings.TrimSpace(callWorker.ExternalParentCallID) != "" {
		callParent, err := s.repos.getCallByCompanyTelephonyIDAndExternalCallID(ctx, tx, callWorker.CompanyTelephonyID, callWorker.ExternalParentCallID)
		if err != nil && !errors.Is(err, errRepoCallNotFound) {
			return srverr.NewServerError(srverr.ErrInternalServerError, "call.workerCallNotExists/getCall").
				SetError(err.Error())
		}
		// Если рд звонок не существует, то проходимся по обычному флоу, но еще дополнительно сохраняем звонок в кеш, чтобы искать родителя
		if err != nil {
			callWorker.WaitingForParent = true
		} else {
			// Если родительский звонок существует, то сохраняем его в звонке
			callWorker.ParentCallID = callParent.ID
		}
	}

	callWorker.Call.ID = uuid.New()
	err := s.repos.saveCall(ctx, tx, callWorker.Call)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.workerCallNotExists/saveCall").
			SetError(err.Error())
	}
	if callWorker.Call.WaitingForParent {
		// TODO: Сохранить звонок ждущий родителя в кеш после сохранения звонка
	}
	// Возможно звонок имеет дочерние звонки, которые ждут родителя, поэтому обновляем их
	childsCall, err := s.repos.getChillCallsByCompanyTelephonyIDAndExternalCallID(ctx, tx, callWorker.CompanyTelephonyID, callWorker.Call.ExternalCallID)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.workerCallNotExists/getChillCallsByCompanyTelephonyIDAndExternalCallID").
			SetError(err.Error())
	}
	if len(childsCall) > 0 {
		for _, childCall := range childsCall {
			childCall.WaitingForParent = false
			childCall.ParentCallID = callWorker.Call.ID
		}
	}
	err = s.repos.updateCallsWaitingForParentAndParentID(ctx, tx, childsCall)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "call.workerCallNotExists/updateCallsWaitingForParent").
			SetError(err.Error())
	}
	// TODO: Удалить из кеша звонки, которые ждут родителя
	return nil
}
