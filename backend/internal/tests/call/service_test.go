package call

import (
	"context"
	"os"
	"strings"
	"telephony/internal/config"
	"telephony/internal/domain"
	call2 "telephony/internal/modules/call"
	"telephony/internal/modules/call_events"
	"telephony/internal/modules/company"
	"telephony/internal/modules/countries"
	"telephony/internal/modules/plan"
	"telephony/internal/modules/telephony_ingestion_pipeline"
	"telephony/internal/shared/database/postgres"
	"telephony/pkg/client/country"
	"telephony/pkg/logger"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

type testCall struct {
	Call    *domain.CallWorker
	Events  []*domain.CallEvent
	Childes []*testCall
}

func newCallTreeNode(direction domain.CallDirection, events []*domain.CallEvent) *domain.CallTree {
	return &domain.CallTree{
		Call: &domain.Call{
			FromNumber:     gofakeit.Phone(),
			ToNumber:       gofakeit.Phone(),
			Direction:      direction,
			ExternalCallID: gofakeit.UUID(),
			Details: &domain.CallDetails{
				RecordingSid:      gofakeit.UUID(),
				RecordingURL:      gofakeit.URL(),
				RecordingDuration: gofakeit.Number(5, 100),
				FromCountry:       gofakeit.Country(),
				FromCity:          gofakeit.City(),
				ToCountry:         gofakeit.Country(),
				ToCity:            gofakeit.City(),
			},
			Events: events,
		},
		Children: nil,
	}
}

func TestService_CallWorkerBaseCreateSuccess(t *testing.T) {
	telephonyTestID := "4591441d-c0a8-409f-9658-eab5459a0b81"

	var tests = map[string]struct {
		CallTree *domain.CallTree
	}{
		"create_call_success_with_childes_1_lvl": {
			CallTree: func() *domain.CallTree {
				// Родительский звонок
				root := newCallTreeNode(
					domain.CallDirectionInbound,
					[]*domain.CallEvent{
						{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
					},
				)

				root.Children = []*domain.CallTree{
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
				}
				return root
			}(),
		},
		"create_call_success_with_childes_2_lvl": {
			CallTree: func() *domain.CallTree {
				root := newCallTreeNode(
					domain.CallDirectionInbound,
					[]*domain.CallEvent{
						{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
					},
				)

				childA := newCallTreeNode(
					domain.CallDirectionInbound,
					[]*domain.CallEvent{
						{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
					},
				)
				childB := newCallTreeNode(
					domain.CallDirectionInbound,
					[]*domain.CallEvent{
						{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
					},
				)
				childC := newCallTreeNode(
					domain.CallDirectionInbound,
					[]*domain.CallEvent{
						{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
						{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
					},
				)

				childA.Children = []*domain.CallTree{
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
				}
				childB.Children = []*domain.CallTree{
					newCallTreeNode(
						domain.CallDirectionInbound,
						[]*domain.CallEvent{
							{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
							{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
						},
					),
				}
				root.Children = []*domain.CallTree{childA, childB, childC}
				return root
			}(),
		},
		"create_call_success_one_call_with_completed": {
			CallTree: newCallTreeNode(
				domain.CallDirectionInbound,
				[]*domain.CallEvent{
					{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
				},
			),
		},
		"create_call_success_one_call_with_double_event": {
			CallTree: newCallTreeNode(
				domain.CallDirectionInbound,
				[]*domain.CallEvent{
					{Status: domain.CallEventStatusInitiated, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusRinging, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusInProgress, Timestamp: time.Now().UTC()},
					{Status: domain.CallEventStatusCompleted, Timestamp: time.Now().UTC()},
				},
			),
		},
	}

	var ctx = context.Background()

	log, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	os.Setenv("TEST", "true")
	os.Setenv("CONFIG_PATH", "/Users/darkness/nexora/backend/configs/local.json")

	cfg := config.MustConfig(log)

	pool, err := postgres.New(ctx, cfg.Postgres, log)
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()

	transaction := postgres.NewTransactionsRepos(cfg.Postgres, pool)

	// Init external clients
	countriesClient := country.NewCountry()

	// Init Repositories
	countryRepos := countries.NewRepository()
	callEventsRepos := call_events.NewRepository()
	callRepos := call2.NewRepository()
	companyRepos := company.NewRepository()
	planRepos := plan.NewRepository()

	// Init Services
	countriesService := countries.NewService(countryRepos, transaction, countriesClient)
	callEventsService := call_events.NewService(callEventsRepos, transaction)
	planService := plan.NewService(planRepos, transaction)
	callService := call2.NewService(callEventsService, callRepos, transaction)
	companyService := company.NewService(callService, companyRepos, transaction)
	telephonyIngestionPipelineService := telephony_ingestion_pipeline.NewService(countriesService, callService, companyService, planService, transaction)

	for name, variableTest := range tests {
		t.Logf("running test: %s", name)

		var createTree func(node *domain.CallTree, parent *domain.CallTree)
		createTree = func(node *domain.CallTree, parent *domain.CallTree) {
			if node == nil || node.Call == nil {
				return
			}

			if parent != nil && parent.Call != nil {
				node.Call.ExternalParentCallID = parent.Call.ExternalCallID
			}

			events := node.Call.Events
			if len(events) == 0 {
				t.Errorf("test failed: empty events for call %s", node.Call.ExternalCallID)
				return
			}

			worker := &domain.CallWorker{
				Call:               node.Call,
				TelephonyAccountID: telephonyTestID,
			}

			worker.Event = events[0]
			sErr := telephonyIngestionPipelineService.CallWorker(ctx, worker, domain.Twilio)
			if sErr != nil {
				t.Errorf("test failed: %s", sErr.Error())
				return
			}

			for _, event := range events[1:] {
				worker.Event = event
				sErr = telephonyIngestionPipelineService.CallWorker(ctx, worker, domain.Twilio)
				if sErr != nil {
					t.Errorf("test failed: %s", sErr.Error())
					return
				}
			}

			for _, child := range node.Children {
				createTree(child, node)
			}
		}

		createTree(variableTest.CallTree, nil)

		checkCallTree(t, ctx, callService, transaction, variableTest.CallTree)

		t.Logf("test finished: %s \n", name)
	}
}

func checkCallTree(t *testing.T, ctx context.Context, callService call2.Service, pool postgres.Transaction, expected *domain.CallTree) {
	if expected == nil || expected.Call == nil {
		t.Errorf("check failed: expected tree is nil")
		return
	}
	rootID := expected.Call.ID
	if rootID == uuid.Nil {
		t.Errorf("check failed: expected root ID is nil")
		return
	}

	tx, err := pool.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("check failed: begin tx: %s", err.Error())
		return
	}
	defer pool.Rollback(ctx, tx)

	actual, sErr := callService.GetCallTreeByCallUUIDWithTx(ctx, tx, rootID)
	if sErr != nil {
		t.Errorf("check failed: %s", sErr.Error())
		return
	}
	compareCallTree(t, expected, actual, "root")
}

func compareCallTree(t *testing.T, expected *domain.CallTree, actual *domain.CallTree, path string) {
	if expected == nil || expected.Call == nil || actual == nil || actual.Call == nil {
		t.Errorf("check failed at %s: one of nodes is nil", path)
		return
	}

	// ID и родство
	if expected.Call.ID != uuid.Nil && expected.Call.ID != actual.Call.ID {
		t.Errorf("check failed at %s: id ожидается %s, получено %s",
			path, expected.Call.ID, actual.Call.ID)
	}
	if expected.Call.ParentCallID != uuid.Nil && expected.Call.ParentCallID != actual.Call.ParentCallID {
		t.Errorf("check failed at %s: parent_call_id ожидается %s, получено %s",
			path, expected.Call.ParentCallID, actual.Call.ParentCallID)
	}
	if expected.Call.CompanyTelephonyID != uuid.Nil && expected.Call.CompanyTelephonyID != actual.Call.CompanyTelephonyID {
		t.Errorf("check failed at %s: company_telephony_id ожидается %s, получено %s",
			path, expected.Call.CompanyTelephonyID, actual.Call.CompanyTelephonyID)
	}

	// Внешние идентификаторы и основные поля
	if expected.Call.ExternalCallID != actual.Call.ExternalCallID {
		t.Errorf("check failed at %s: external_call_id ожидается %s, получено %s",
			path, expected.Call.ExternalCallID, actual.Call.ExternalCallID)
	}
	if expected.Call.ExternalParentCallID != actual.Call.ExternalParentCallID {
		t.Errorf("check failed at %s: external_parent_call_id ожидается %s, получено %s",
			path, expected.Call.ExternalParentCallID, actual.Call.ExternalParentCallID)
	}
	if expected.Call.WaitingForParent != actual.Call.WaitingForParent {
		t.Errorf("check failed at %s: waiting_for_parent ожидается %v, получено %v",
			path, expected.Call.WaitingForParent, actual.Call.WaitingForParent)
	}
	if expected.Call.FromNumber != actual.Call.FromNumber {
		t.Errorf("check failed at %s: from_number ожидается %s, получено %s",
			path, expected.Call.FromNumber, actual.Call.FromNumber)
	}
	if expected.Call.ToNumber != actual.Call.ToNumber {
		t.Errorf("check failed at %s: to_number ожидается %s, получено %s",
			path, expected.Call.ToNumber, actual.Call.ToNumber)
	}
	if expected.Call.Direction != actual.Call.Direction {
		t.Errorf("check failed at %s: direction ожидается %s, получено %s",
			path, expected.Call.Direction, actual.Call.Direction)
	}

	// Details
	compareCallDetails(t, expected.Call.Details, actual.Call.Details, path)

	// Events
	compareCallEvents(t, expected.Call.Events, actual.Call.Events, path)

	// Дети
	expChildren := expected.Children
	actChildren := actual.Children

	if len(expChildren) != len(actChildren) {
		t.Errorf("check failed at %s: детей ожидается %d, получено %d",
			path, len(expChildren), len(actChildren))
	}

	actByExternal := make(map[string]*domain.CallTree, len(actChildren))
	for _, ch := range actChildren {
		if ch == nil || ch.Call == nil {
			continue
		}
		actByExternal[ch.Call.ExternalCallID] = ch
	}

	for _, expChild := range expChildren {
		if expChild == nil || expChild.Call == nil {
			continue
		}
		actChild := actByExternal[expChild.Call.ExternalCallID]
		if actChild == nil {
			t.Errorf("check failed at %s: не найден ребёнок external_call_id=%s",
				path, expChild.Call.ExternalCallID)
			continue
		}
		compareCallTree(t, expChild, actChild, path+" -> "+expChild.Call.ExternalCallID)
	}
}

func compareCallDetails(t *testing.T, expected *domain.CallDetails, actual *domain.CallDetails, path string) {
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Errorf("check failed at %s: details expected=%v actual=%v", path, expected != nil, actual != nil)
		return
	}

	if expected.RecordingSid != actual.RecordingSid {
		t.Errorf("check failed at %s: recording_sid ожидается %s, получено %s",
			path, expected.RecordingSid, actual.RecordingSid)
	}
	if expected.RecordingURL != actual.RecordingURL {
		t.Errorf("check failed at %s: recording_url ожидается %s, получено %s",
			path, expected.RecordingURL, actual.RecordingURL)
	}
	if expected.RecordingDuration != actual.RecordingDuration {
		t.Errorf("check failed at %s: recording_duration ожидается %d, получено %d",
			path, expected.RecordingDuration, actual.RecordingDuration)
	}
	if expected.FromCountry != actual.FromCountry {
		t.Errorf("check failed at %s: from_country ожидается %s, получено %s",
			path, expected.FromCountry, actual.FromCountry)
	}
	if expected.FromCity != actual.FromCity {
		t.Errorf("check failed at %s: from_city ожидается %s, получено %s",
			path, expected.FromCity, actual.FromCity)
	}
	if expected.ToCountry != actual.ToCountry {
		t.Errorf("check failed at %s: to_country ожидается %s, получено %s",
			path, expected.ToCountry, actual.ToCountry)
	}
	if expected.ToCity != actual.ToCity {
		t.Errorf("check failed at %s: to_city ожидается %s, получено %s",
			path, expected.ToCity, actual.ToCity)
	}
	if expected.Carrier != actual.Carrier {
		t.Errorf("check failed at %s: carrier ожидается %s, получено %s",
			path, expected.Carrier, actual.Carrier)
	}
	if expected.Trunk != actual.Trunk {
		t.Errorf("check failed at %s: trunk ожидается %s, получено %s",
			path, expected.Trunk, actual.Trunk)
	}
}

func compareCallEvents(t *testing.T, expected []*domain.CallEvent, actual []*domain.CallEvent, path string) {
	expected = removeConsecutiveDuplicateEvents(expected)

	expMap := make(map[domain.CallEventStatus]*domain.CallEvent, len(expected))
	for _, ev := range expected {
		if ev == nil {
			t.Errorf("check failed at %s: expected event nil", path)
			continue
		}
		if _, exists := expMap[ev.Status]; exists {
			t.Errorf("check failed at %s: duplicate expected status %s", path, ev.Status)
			continue
		}
		expMap[ev.Status] = ev
	}

	actMap := make(map[domain.CallEventStatus]*domain.CallEvent, len(actual))
	for _, ev := range actual {
		if ev == nil {
			t.Errorf("check failed at %s: actual event nil", path)
			continue
		}
		if _, exists := actMap[ev.Status]; exists {
			t.Errorf("check failed at %s: duplicate actual status %s", path, ev.Status)
			continue
		}
		actMap[ev.Status] = ev
	}

	if len(expMap) != len(actMap) {
		t.Errorf("check failed at %s: events ожидается %d, получено %d",
			path, len(expMap), len(actMap))
	}

	for status, exp := range expMap {
		act := actMap[status]
		if act == nil {
			t.Errorf("check failed at %s: отсутствует событие со статусом %s", path, status)
			continue
		}
		if !exp.Timestamp.Equal(act.Timestamp) {
			t.Errorf("check failed at %s: event[%s].timestamp ожидается %s, получено %s",
				path, status, exp.Timestamp, act.Timestamp)
		}
	}
}

func printCallTree(t *domain.CallTree, prefix string, isLast bool, b *strings.Builder) {
	if t == nil || t.Call == nil {
		return
	}
	branch := "├── "
	nextPrefix := prefix + "│   "
	if isLast {
		branch = "└── "
		nextPrefix = prefix + "    "
	}

	b.WriteString(prefix)
	b.WriteString(branch)
	b.WriteString(t.Call.ID.String())
	b.WriteString(" | ")
	b.WriteString(t.Call.ExternalCallID)
	b.WriteString(" | ")
	b.WriteString(string(t.Call.Direction))
	b.WriteString(" | ")
	b.WriteString(t.Call.FromNumber)
	b.WriteString(" -> ")
	b.WriteString(t.Call.ToNumber)
	b.WriteString("\n")

	for i, ch := range t.Children {
		printCallTree(ch, nextPrefix, i == len(t.Children)-1, b)
	}
}

func removeConsecutiveDuplicateEvents(events []*domain.CallEvent) []*domain.CallEvent {
	if len(events) == 0 {
		return events
	}

	out := make([]*domain.CallEvent, 0, len(events))
	var lastStatus domain.CallEventStatus
	hasLast := false

	for _, ev := range events {
		if ev == nil {
			continue
		}
		if hasLast && ev.Status == lastStatus {
			continue
		}
		out = append(out, ev)
		lastStatus = ev.Status
		hasLast = true
	}
	return out
}
