package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"telephony/internal/config"
	"telephony/internal/cron"
	"telephony/internal/modules/auth"
	"telephony/internal/modules/call"
	"telephony/internal/modules/call_events"
	"telephony/internal/modules/company"
	"telephony/internal/modules/countries"
	"telephony/internal/modules/notification"
	"telephony/internal/modules/plan"
	telephonywebhook "telephony/internal/modules/telephony/webhook"
	mango2 "telephony/internal/modules/telephony/webhook/mango"
	mts2 "telephony/internal/modules/telephony/webhook/mts"
	twilio2 "telephony/internal/modules/telephony/webhook/twilio"
	zadarma2 "telephony/internal/modules/telephony/webhook/zadarma"
	"telephony/internal/modules/telephony_ingestion_pipeline"
	"telephony/internal/modules/user"
	"telephony/internal/runner"
	http2 "telephony/internal/server/http"
	cache_redis "telephony/internal/shared/cache/redis"
	"telephony/internal/shared/database/postgres"
	redis2 "telephony/internal/shared/database/redis"
	"telephony/internal/shared/middleware"
	"telephony/internal/shared/response"
	transperr "telephony/internal/shared/transport_error"
	"telephony/pkg/client/country"
	email_stmp "telephony/pkg/client/email_sender/stmp"
	"telephony/pkg/client/oauth/appleoauth"
	"telephony/pkg/client/oauth/googleoauth"
	"telephony/pkg/client/yandexstorage"
	"telephony/pkg/logger"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func main() {
	var ctx = context.Background()

	log, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg := config.MustConfig(log)

	pool, err := postgres.New(ctx, cfg.Postgres, log)
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()

	transaction := postgres.NewTransactionsRepos(cfg.Postgres, pool)

	client, err := redis2.NewRedis(ctx, cfg.Redis, log)
	if err != nil {
		log.Panic("failed to connected redis: ", err)
	}
	defer func(client *redis.Client) {
		err := client.Close()
		if err != nil {
			log.Error("failed to close redis client: ", err)
		}
	}(client)

	var (
		httpResp         = response.NewHTTPResponse(log, true)
		convert          = transperr.NewErrorConverter()
		mid              = middleware.NewMiddleware(log, cfg, httpResp, convert)
		validationFormat = strfmt.NewFormats()

		router = mux.NewRouter()
	)

	middleware.SetupCORS(router, cfg.Handler.AllowedCORSOrigins, mid)
	router.Use(middleware.AccessLog(log))

	setupBaseRoutes(router)

	initBusinessLogic(
		router,
		mid,
		httpResp,
		convert,
		validationFormat,

		transaction,
		client,

		log,
		*cfg,
	)
	httpServer := http2.NewServer(&cfg.Server, router)

	logRegisteredEndpoints(log, cfg.Server.Port, router)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server failed: %v", err)
		}
	}()

	log.Infof("server listening on port [%d] | Env %s", cfg.Server.Port, cfg.Env)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	log.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Errorf("error shutdown: %s", err)
	}

	log.Info("server shutdown")
}

func initBusinessLogic(
	router *mux.Router,
	mid middleware.Middleware,
	httpResp response.HttpResponse,
	convert transperr.ErrorConverter,
	validationFormat strfmt.Registry,

	transaction postgres.Transaction,
	redisClient *redis.Client,

	log logger.Logger,
	cfg config.Config,
) {
	redisCache := cache_redis.NewCacheRedis(redisClient)

	// Init external clients
	countriesClient := country.NewCountry()

	// Init Repositories
	countryRepos := countries.NewRepository()
	callEventsRepos := call_events.NewRepository()
	callRepos := call.NewRepository()
	companyRepos := company.NewRepository()
	planRepos := plan.NewRepository()
	userRepos := user.NewRepository()
	notificationRepos := notification.NewRepository()

	//phoneSender := newPhoneSender(cfg)
	emailSender := email_stmp.NewSMTPSender(
		cfg.Sms.Smtp.Host,
		cfg.Sms.Smtp.Port,
		cfg.Sms.Smtp.Username,
		cfg.Sms.Smtp.Password,
		cfg.Sms.Smtp.From,
	)

	// Init Services
	countriesService := countries.NewService(countryRepos, transaction, countriesClient)
	callEventsService := call_events.NewService(callEventsRepos, transaction)
	planService := plan.NewService(planRepos, transaction)
	callService := call.NewService(callEventsService, callRepos, transaction)
	companyService := company.NewService(callService, companyRepos, transaction)
	telephonyIngestionPipelineService := telephony_ingestion_pipeline.NewService(countriesService, callService, companyService, planService, transaction)
	telephonyCallService := telephonywebhook.NewTelephonyCall(telephonyIngestionPipelineService)
	notificationService := notification.NewService(notificationRepos, transaction)

	s3Storage, err := yandexstorage.NewClient(
		cfg.S3.AccessKeyID,
		cfg.S3.SecretAccessKey,
		cfg.S3.Bucket,
		cfg.S3.Region,
		cfg.S3.Endpoint,
		cfg.S3.PublicBaseURL,
	)
	if err != nil {
		log.Panic("storage init: ", err)
	}
	err = s3Storage.Ping(context.Background())
	if err != nil {
		log.Panic("storage ping: ", err)
	}

	userService := user.NewService(userRepos, companyService, transaction, s3Storage, cfg.Storage.AvatarPrefix)

	// OAuth Clients
	googleOAuthClient, err := googleoauth.NewClient(
		cfg.Auth.OAuth.OAuthGoogle.ClientID,
		cfg.Auth.OAuth.OAuthGoogle.ClientSecret,
		cfg.Auth.OAuth.OAuthBackendBaseURL+cfg.Auth.OAuth.OAuthGoogle.RedirectPath,
		cfg.Auth.OAuth.OAuthGoogle.UserinfoURL,
	)
	if err != nil {
		log.Errorf("google oauth init: %v", err)
	}

	appleOAuthClient, err := appleoauth.NewClient(
		cfg.Auth.OAuth.OAuthApple.ClientID,
		cfg.Auth.OAuth.OAuthApple.AuthURL,
		cfg.Auth.OAuth.OAuthApple.JWKSURL,
		cfg.Auth.OAuth.OAuthApple.Issuer,
	)
	if err != nil {
		log.Errorf("apple oauth init: %v", err)
	}

	authService := auth.NewService(userService, companyService, transaction, redisCache, cfg, emailSender, googleOAuthClient, appleOAuthClient)

	runner.InitHandlers(router, mid,
		auth.NewRunnerHandlerV1(router, authService, httpResp, convert, cfg, log, redisClient),
		user.NewRunnerHandlerV1(router, userService, httpResp, convert, cfg.Auth.JWTSecret),
		call.NewRunnerHandlerV1(router, callService, httpResp, convert),
		company.NewRunnerHandlerV1(router, companyService, httpResp, convert),
		twilio2.NewRunnerHandlerV1(router, telephonyCallService, httpResp, convert, validationFormat),
		mango2.NewRunnerHandlerV1(router, telephonyCallService, httpResp, convert, validationFormat),
		zadarma2.NewRunnerHandlerV1(router, telephonyCallService, httpResp, convert, validationFormat),
		mts2.NewRunnerHandlerV1(router, telephonyCallService, httpResp, convert, validationFormat),
		notification.NewRunnerHandlerV1(router, notificationService, httpResp, convert),
	)

	runner.InitCronTasks(log,
		cron.NewUpdateCountriesCron(countriesService, log),
	)
}

func setupBaseRoutes(router *mux.Router) {
	router.HandleFunc("/health", http2.HealthHandler).Methods(http.MethodGet)
	router.HandleFunc("/swagger", http2.SwaggerUIHandler).Methods(http.MethodGet)
	router.HandleFunc("/swagger.yaml", http2.SwaggerYamlHandler).Methods(http.MethodGet)
}

func logRegisteredEndpoints(log logger.Logger, port int, router *mux.Router) {
	endpoints := http2.RegisteredEndpoints(router)
	if len(endpoints) == 0 {
		return
	}
	const methodWidth = 8
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	log.Infof("HTTP endpoints (base: %s):", baseURL)
	for _, ep := range endpoints {
		log.Infof("  %-*s %s", methodWidth, ep.Method, ep.Path)
	}
}
