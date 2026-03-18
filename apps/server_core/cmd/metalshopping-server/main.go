package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	cataloggov "metalshopping/server_core/internal/modules/catalog/adapters/governance"
	catalogpg "metalshopping/server_core/internal/modules/catalog/adapters/postgres"
	catalogapp "metalshopping/server_core/internal/modules/catalog/application"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
	cataloghttp "metalshopping/server_core/internal/modules/catalog/transport/http"
	iamgov "metalshopping/server_core/internal/modules/iam/adapters/governance"
	iampg "metalshopping/server_core/internal/modules/iam/adapters/postgres"
	iamapp "metalshopping/server_core/internal/modules/iam/application"
	iamhttp "metalshopping/server_core/internal/modules/iam/transport/http"
	inventorypg "metalshopping/server_core/internal/modules/inventory/adapters/postgres"
	inventoryapp "metalshopping/server_core/internal/modules/inventory/application"
	inventoryhttp "metalshopping/server_core/internal/modules/inventory/transport/http"
	pricinggov "metalshopping/server_core/internal/modules/pricing/adapters/governance"
	pricingpg "metalshopping/server_core/internal/modules/pricing/adapters/postgres"
	pricingapp "metalshopping/server_core/internal/modules/pricing/application"
	pricinghttp "metalshopping/server_core/internal/modules/pricing/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/authsession"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
	"metalshopping/server_core/internal/platform/messaging/outbox"
	"metalshopping/server_core/internal/platform/observability"
	"metalshopping/server_core/internal/platform/runtime_config"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

func main() {
	if err := runtime_config.LoadDotEnvIfPresent(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	addr, err := runtime_config.HTTPAddressFromEnv()
	if err != nil {
		log.Fatalf("load server config: %v", err)
	}
	environment := runtime_config.EnvironmentFromEnv()

	db, _, err := pgdb.OpenFromEnv(context.Background())
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	authenticator, err := platformauth.NewAuthenticatorFromEnv()
	if err != nil {
		log.Fatalf("load bootstrap authenticator: %v", err)
	}
	governanceRegistry := governancebootstrap.NewRegistry()
	featureFlagResolver, err := feature_flags.NewPostgresResolver(context.Background(), db, governanceRegistry)
	if err != nil {
		log.Fatalf("load governance feature flags from database: %v", err)
	}
	thresholdResolver, err := threshold_resolver.NewPostgresResolver(context.Background(), db, governanceRegistry)
	if err != nil {
		log.Fatalf("load governance thresholds from database: %v", err)
	}
	policyResolver, err := policy_resolver.NewPostgresResolver(context.Background(), db, governanceRegistry)
	if err != nil {
		log.Fatalf("load governance policies from database: %v", err)
	}
	outboxStore := outbox.NewStore(db)
	outboxDispatcher := outbox.NewDispatcher(outboxStore, outbox.NewLoggingPublisher(log.Default()))
	go outboxDispatcher.Run(context.Background())

	iamRepo := iampg.NewRepository(db)
	catalogRepo := catalogpg.NewRepository(db, outboxStore)
	inventoryRepo := inventorypg.NewRepository(db, outboxStore)
	pricingRepo := pricingpg.NewRepository(db, outboxStore)
	iamAuthorizer := iamapp.NewStaticAuthorizer()
	iamAuthorization := iamapp.NewAuthorizationService(iamRepo, iamAuthorizer)
	iamAdminService := iamapp.NewAdminService(iamRepo, iamgov.NewAdminPolicyGuard(policyResolver, environment))
	iamAdminHandler := iamhttp.NewAdminHandler(iamAdminService, iamAuthorization)
	catalogProductCreationGuard := cataloggov.NewProductCreationGuard(featureFlagResolver, environment)
	catalogDescriptionGuard := cataloggov.NewDescriptionGuard(thresholdResolver, environment)
	catalogService := catalogapp.NewService(catalogRepo, catalogProductCreationGuard, catalogDescriptionGuard)
	catalogProductsPortfolioService := catalogreadmodel.NewProductsPortfolioService(catalogRepo)
	catalogHandler := cataloghttp.NewHandler(catalogService, catalogProductsPortfolioService, iamAuthorization)
	inventoryService := inventoryapp.NewService(inventoryRepo)
	inventoryHandler := inventoryhttp.NewHandler(inventoryService, iamAuthorization)
	pricingManualOverrideGuard := pricinggov.NewManualOverrideGuard(policyResolver, environment)
	pricingService := pricingapp.NewService(pricingRepo, pricingManualOverrideGuard)
	pricingHandler := pricinghttp.NewHandler(pricingService, iamAuthorization)

	mux := http.NewServeMux()
	requestAuthenticator := platformauth.NewBearerRequestAuthenticator(authenticator)
	publicAuthPaths := []string{
		"/health/live",
		"/health/ready",
	}
	if authSessionConfig, err := authsession.ConfigFromEnv(environment); err == nil {
		if tokenValidator, validatorErr := platformauth.NewJWTAuthenticatorFromEnv(); validatorErr == nil {
			authSessionStore := authsession.NewStore(db)
			authSessionService := authsession.NewService(
				authSessionConfig,
				authSessionStore,
				authsession.NewOIDCClient(authSessionConfig),
				tokenValidator,
				authsession.NewRuntimePolicyResolver(featureFlagResolver, thresholdResolver, environment),
				iamRepo,
				iamAuthorizer,
			)
			authSessionHandler := authsession.NewHandler(authSessionService, authSessionConfig)
			authSessionHandler.RegisterRoutes(mux)
			requestAuthenticator = platformauth.NewBearerOrCookieRequestAuthenticator(
				requestAuthenticator,
				authsession.NewRequestAuthenticator(authSessionConfig, authSessionStore),
			)
			publicAuthPaths = append(publicAuthPaths,
				"/api/v1/auth/session/login",
				"/api/v1/auth/session/callback",
			)
		} else {
			log.Printf("auth/session disabled: %v", validatorErr)
		}
	} else {
		log.Printf("auth/session disabled: %v", err)
	}

	authMiddleware := platformauth.NewMiddlewareWithRequestAuthenticator(requestAuthenticator, publicAuthPaths)
	tenancyMiddleware := tenancy_runtime.NewMiddleware([]string{
		"/health/live",
		"/health/ready",
		"/api/v1/auth/session/login",
		"/api/v1/auth/session/callback",
	})
	requestLogging := observability.NewRequestLoggingMiddleware()
	corsMiddleware := observability.NewCORSMiddleware(runtime_config.CORSAllowedOriginsFromEnv(environment))

	mux.Handle("/health/live", observability.NewLiveHandler())
	mux.Handle("/health/ready", observability.NewReadyHandler(postgresReadiness(db)))
	iamAdminHandler.RegisterRoutes(mux)
	catalogHandler.RegisterRoutes(mux)
	inventoryHandler.RegisterRoutes(mux)
	pricingHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(requestLogging(authMiddleware.Wrap(tenancyMiddleware.Wrap(mux)))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("metalshopping-server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}

func postgresReadiness(db *sql.DB) observability.ReadinessChecker {
	return func(ctx context.Context) error {
		if db == nil {
			return errors.New("postgres database is not configured")
		}

		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return db.PingContext(pingCtx)
	}
}
