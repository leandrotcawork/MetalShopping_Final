package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/observability"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

func composeHTTPServer(
	runtime runtimeComposition,
	modules moduleComposition,
	authSession authSessionComposition,
) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/health/live", observability.NewLiveHandler())
	mux.Handle("/health/ready", observability.NewReadyHandler(postgresReadiness(runtime.db)))

	authSession.registerHTTP(mux)
	modules.registerHTTP(mux)

	authMiddleware := platformauth.NewMiddlewareWithRequestAuthenticator(
		authSession.requestAuthenticator,
		authSession.publicPaths,
	)
	tenancyMiddleware := tenancy_runtime.NewMiddleware(uniqueStrings(authSession.publicPaths))
	requestLogging := observability.NewRequestLoggingMiddleware()
	corsMiddleware := observability.NewCORSMiddleware(runtime.allowedOrigins)

	return &http.Server{
		Addr: runtime.addr,
		Handler: corsMiddleware(
			requestLogging(
				authMiddleware.Wrap(
					tenancyMiddleware.Wrap(mux),
				),
			),
		),
		ReadHeaderTimeout: 5 * time.Second,
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

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
