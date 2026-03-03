package main

import (
	"context"
	"telephony/internal/config"
	"telephony/internal/shared/database/postgres"
	"telephony/pkg/logger"
	"time"
)

type dictionary struct {
	MainType    string
	Name        string
	Value       string
	Description string
}

func main() {
	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg := config.MustConfig(log)

	pool, err := postgres.New(ctx, cfg.Postgres, log)
	if err != nil {
		log.Errorf("failed to connect to postgres: %v", err)
		return
	}
	defer pool.Close()

	transaction := postgres.NewTransactionsRepos(cfg.Postgres, pool)

	tx, err := transaction.BeginTransaction(ctx)
	if err != nil {
		log.Errorf("failed to begin transaction: %v", err)
		return
	}
	defer transaction.Rollback(ctx, tx)

	for _, value := range dictionaryValues {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO dictionary (main_type, key, value, comment)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (main_type, key) DO NOTHING`,
			value.MainType, value.Name, value.Value, value.Description,
		)
		if err != nil {
			log.Errorf("failed to insert dictionary value: %v", err)
			return
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return
	}
	log.Info("Dictionary values inserted successfully")
}

var dictionaryValues = []*dictionary{
	// Call Event Statuses
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusQueued",
		Value:       "call_event_status_queued",
		Description: "Звонок поставлен в очередь, не начат",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusInitiated",
		Value:       "call_event_status_initiated",
		Description: "Начат процесс дозвона",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusRinging",
		Value:       "call_event_status_ringing",
		Description: "Вызываемая сторона слышит звонок",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusInProgress",
		Value:       "call_event_status_in_progress",
		Description: "Звонок принят, разговор идёт",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusCompleted",
		Value:       "call_event_status_completed",
		Description: "Звонок завершён успешно",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusBusy",
		Value:       "call_event_status_busy",
		Description: "Линия занята",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusFailed",
		Value:       "call_event_status_failed",
		Description: "Ошибка соединения",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusNoAnswer",
		Value:       "call_event_status_no_answer",
		Description: "Не ответили",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusCanceled",
		Value:       "call_event_status_canceled",
		Description: "Звонок отменён",
	},
	{
		MainType:    "CallEventStatus",
		Name:        "CallEventStatusTimeout",
		Value:       "call_event_status_timeout",
		Description: "Время ожидания ответа истекло",
	},
	{
		MainType:    "CallDirection",
		Name:        "CallDirectionInbound",
		Value:       "call_direction_inbound",
		Description: "Входящий звонок на номер (то есть кто-то звонит на ваш Twilio или иной телефонии номер)",
	},
	{
		MainType:    "CallDirection",
		Name:        "CallDirectionOutboundApi",
		Value:       "call_direction_outbound_api",
		Description: "исходящий звонок, инициированный через REST API Twilio Или иной телефонии (например, программно с вашего сервера)",
	},
	{
		MainType:    "CallDirection",
		Name:        "CallDirectionOutboundDial",
		Value:       "call_direction_outbound_dial",
		Description: "исходящий звонок, созданный внутри Twilio с помощью XML-тега <Dial> (например, когда вы внутри звонка перенаправляете на другой номер)",
	},
}
