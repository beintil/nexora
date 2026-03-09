package user

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
	router    *mux.Router
	handler   Handler
	httpResp  response.HttpResponse
	converter transperr.ErrorConverter
	jwtSecret string
}

func NewRunnerHandlerV1(
	router *mux.Router,
	service Service,
	httpResp response.HttpResponse,
	converter transperr.ErrorConverter,
	jwtSecret string,
) runner.Handler {
	h := NewHandler(service, httpResp, converter)
	return &runnerV1{
		router:    router.PathPrefix("/v1").Subrouter(),
		handler:   h,
		httpResp:  httpResp,
		converter: converter,
		jwtSecret: jwtSecret,
	}
}

func (r *runnerV1) Init() []runner.Runner {
	return []runner.Runner{r}
}

func (r *runnerV1) RouterWithVersion() *mux.Router {
	return r.router
}

func (r *runnerV1) Run(router *mux.Router, mid middleware.Middleware) {
	router.Handle("/profile",
		mid.WithAccess(http.HandlerFunc(r.handler.handleGetProfile)),
	).Methods(http.MethodGet)

	router.Handle("/profile",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner)(http.HandlerFunc(r.handler.handleUpdateProfile)),
		),
	).Methods(http.MethodPatch)

	router.Handle("/profile/avatar",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner)(http.HandlerFunc(r.handler.handleUploadAvatar)),
		),
	).Methods(http.MethodPost)

	// Staff management
	router.Handle("/users",
		mid.WithAccess(http.HandlerFunc(r.handler.handleListUsers)),
	).Methods(http.MethodGet)

	router.Handle("/users",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner)(http.HandlerFunc(r.handler.handleCreateStaff)),
		),
	).Methods(http.MethodPost)

	router.Handle("/users/{id}",
		mid.WithAccess(
			mid.PermissionMiddleware(domain.RoleOwner)(http.HandlerFunc(r.handler.handleDeleteUser)),
		),
	).Methods(http.MethodDelete)
}
