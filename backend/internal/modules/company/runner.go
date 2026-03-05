package company

import (
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/runner"
	"telephony/internal/shared/middleware"
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
	h := NewHandler(service, httpResp, converter)
	return &runnerV1{
		router:  router.PathPrefix("/v1").Subrouter(),
		handler: h,
	}
}

func (r *runnerV1) Init() []runner.Runner {
	return []runner.Runner{r}
}

func (r *runnerV1) RouterWithVersion() *mux.Router {
	return r.router
}

func (r *runnerV1) Run(router *mux.Router, mid middleware.Middleware) {
	companyRouter := router.PathPrefix("/company").Subrouter()
	companyRouter.Use(mid.WithAccess)
	companyRouter.Use(mid.PermissionMiddleware(domain.RoleOwner))

	companyRouter.HandleFunc("/telephony", r.handler.handleListCompanyTelephony).Methods(http.MethodGet)
	companyRouter.HandleFunc("/telephony", r.handler.handleAttachCompanyTelephony).Methods(http.MethodPost)
	companyRouter.HandleFunc("/telephony/dictionary", r.handler.handleListTelephonyDictionary).Methods(http.MethodGet)
	companyRouter.HandleFunc("/telephony/{id}", r.handler.handleDetachCompanyTelephony).Methods(http.MethodDelete)
}
