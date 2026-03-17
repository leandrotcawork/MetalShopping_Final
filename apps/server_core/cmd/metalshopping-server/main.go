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
	cataloghttp "metalshopping/server_core/internal/modules/catalog/transport/http"
	iampg "metalshopping/server_core/internal/modules/iam/adapters/postgres"
	iamapp "metalshopping/server_core/internal/modules/iam/application"
	iamhttp "metalshopping/server_core/internal/modules/iam/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
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

	authenticator, err := platformauth.NewStaticBearerAuthenticatorFromEnv()
	if err != nil {
		log.Fatalf("load bootstrap authenticator: %v", err)
	}
	governanceRegistry := governancebootstrap.NewRegistry()
	featureFlagResolver, err := feature_flags.NewPostgresResolver(context.Background(), db, governanceRegistry)
	if err != nil {
		log.Fatalf("load governance feature flags from database: %v", err)
	}

	iamRepo := iampg.NewRepository(db)
	catalogRepo := catalogpg.NewRepository(db)
	iamAuthorizer := iamapp.NewStaticAuthorizer()
	iamAuthorization := iamapp.NewAuthorizationService(iamRepo, iamAuthorizer)
	iamAdminService := iamapp.NewAdminService(iamRepo)
	iamAdminHandler := iamhttp.NewAdminHandler(iamAdminService, iamAuthorization)
	catalogProductCreationGuard := cataloggov.NewProductCreationGuard(featureFlagResolver, environment)
	catalogService := catalogapp.NewService(catalogRepo, catalogProductCreationGuard)
	catalogHandler := cataloghttp.NewHandler(catalogService, iamAuthorization)

	authMiddleware := platformauth.NewMiddleware(authenticator, []string{
		"/health/live",
		"/health/ready",
	})
	tenancyMiddleware := tenancy_runtime.NewMiddleware([]string{
		"/health/live",
		"/health/ready",
	})
	requestLogging := observability.NewRequestLoggingMiddleware()

	mux := http.NewServeMux()
	mux.Handle("/health/live", observability.NewLiveHandler())
	mux.Handle("/health/ready", observability.NewReadyHandler(postgresReadiness(db)))
	iamAdminHandler.RegisterRoutes(mux)
	catalogHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           requestLogging(authMiddleware.Wrap(tenancyMiddleware.Wrap(mux))),
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
