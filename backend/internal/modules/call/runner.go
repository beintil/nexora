package call

import (
	"telephony/internal/runner"
	"telephony/internal/shared/response"
	transperr "telephony/internal/shared/transport_error"

	"github.com/gorilla/mux"
)

type runnerV1 struct {
	router  *mux.Router
	handler Handler
}

func NewRunnerHandlerV1(
	router *mux.Router,
	service Service,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
) runner.Handler {
	return &runnerV1{
		router:  router.PathPrefix("/v1").Subrouter(),
		handler: NewHandler(service, httpResp, converter),
	}
}

func (r *runnerV1) Init() []runner.Runner {
	return []runner.Runner{r.handler}
}

func (r *runnerV1) RouterWithVersion() *mux.Router {
	return r.router
}
