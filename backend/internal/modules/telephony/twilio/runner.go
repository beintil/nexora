package twilio

import (
	"telephony/internal/runner"
	"telephony/internal/shared/response"
	transperr "telephony/internal/shared/transport_error"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

type handlerV1 struct {
	router *mux.Router

	httpResp response.HttpResponse

	Handler
}

func NewRunnerHandlerV1(
	router *mux.Router,

	service Service,

	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,

	validationFormat strfmt.Registry,
) runner.Handler {
	return &handlerV1{
		router: router.PathPrefix("/v1").Subrouter(),

		httpResp: httpResp,

		Handler: NewHandler(service, httpResp, converter, validationFormat),
	}
}

func (m *handlerV1) Init() []runner.Runner {
	return []runner.Runner{
		m.Handler,
	}
}

func (m *handlerV1) RouterWithVersion() *mux.Router {
	return m.router
}
