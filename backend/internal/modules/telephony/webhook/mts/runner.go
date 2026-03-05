package mts

import (
	"telephony/internal/modules/telephony/webhook"
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

	call webhook.TelephonyCall,

	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,

	validationFormat strfmt.Registry,
) runner.Handler {
	// validationFormat пока не используется, но оставляем для единообразия.
	_ = validationFormat

	return &handlerV1{
		router: router.PathPrefix("/v1").Subrouter(),

		httpResp: httpResp,

		Handler: NewHandler(call, httpResp, converter),
	}
}

func (h *handlerV1) Init() []runner.Runner {
	return []runner.Runner{
		h.Handler,
	}
}

func (h *handlerV1) RouterWithVersion() *mux.Router {
	return h.router
}
