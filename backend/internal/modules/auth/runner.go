package auth

import (
	"telephony/internal/config"
	"telephony/internal/runner"
	"telephony/internal/shared/response"
	transperr "telephony/internal/shared/transport_error"
	"telephony/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type handlerV1 struct {
	router *mux.Router

	Handler
}

func NewRunnerHandlerV1(
	router *mux.Router,
	service Service,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
	cfg config.Config,
	log logger.Logger,
	redisClient *redis.Client,
) runner.Handler {
	return &handlerV1{
		router: router.PathPrefix("/v1").Subrouter(),

		Handler: NewHandler(service, httpResp, converter, cfg, log, redisClient),
	}
}

func (m *handlerV1) Init() []runner.Runner {
	return []runner.Runner{m.Handler}
}

func (m *handlerV1) RouterWithVersion() *mux.Router {
	return m.router
}
