package notification

import (
	"net/http"
	"telephony/internal/runner"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	transperr "telephony/internal/shared/transport_error"

	"github.com/gorilla/mux"
)

type runnerV1 struct {
	router    *mux.Router
	handler   Handler
	httpResp  response.HttpResponse
	converter transperr.ErrorConverter
}

func NewRunnerHandlerV1(
	router *mux.Router,
	service Service,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
) runner.Handler {
	h := NewHandler(service, httpResp, converter)
	return &runnerV1{
		router:    router.PathPrefix("/v1").Subrouter(),
		handler:   h,
		httpResp:  httpResp,
		converter: converter,
	}
}

func (r *runnerV1) Init() []runner.Runner {
	return []runner.Runner{r}
}

func (r *runnerV1) RouterWithVersion() *mux.Router {
	return r.router
}

func (r *runnerV1) Run(router *mux.Router, mid middleware.Middleware) {
	router.Handle("/notifications",
		mid.WithAccess(http.HandlerFunc(r.handler.handleGetNotifications)),
	).Methods(http.MethodGet)

	router.Handle("/notifications/{id}/read",
		mid.WithAccess(http.HandlerFunc(r.handler.handleMarkAsRead)),
	).Methods(http.MethodPost)
}
