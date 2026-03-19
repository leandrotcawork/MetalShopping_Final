package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/authsession"
)

type authSessionComposition struct {
	requestAuthenticator platformauth.RequestAuthenticator
	publicPaths          []string
	registerHTTP         func(mux *http.ServeMux)
}

type authSessionBootstrapMode string

const (
	authSessionModeRequired authSessionBootstrapMode = "required"
	authSessionModeOptional authSessionBootstrapMode = "optional"
	authSessionModeDisabled authSessionBootstrapMode = "disabled"
)

func resolveAuthSessionBootstrapMode() (authSessionBootstrapMode, error) {
	rawMode := strings.ToLower(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_SESSION_MODE")))
	if rawMode == "" {
		if strings.ToLower(strings.TrimSpace(os.Getenv("MS_AUTH_MODE"))) == "static" {
			return authSessionModeOptional, nil
		}
		return authSessionModeRequired, nil
	}

	mode := authSessionBootstrapMode(rawMode)
	switch mode {
	case authSessionModeRequired, authSessionModeOptional, authSessionModeDisabled:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid MS_AUTH_WEB_SESSION_MODE %q (expected required, optional, or disabled)", rawMode)
	}
}

func composeAuthSession(
	runtime runtimeComposition,
	governance governanceComposition,
	modules moduleComposition,
) (authSessionComposition, error) {
	basePublicPaths := []string{
		"/health/live",
		"/health/ready",
	}
	bearerAuthenticator := platformauth.NewBearerRequestAuthenticator(runtime.authenticator)

	composition := authSessionComposition{
		requestAuthenticator: bearerAuthenticator,
		publicPaths:          append([]string(nil), basePublicPaths...),
		registerHTTP:         func(*http.ServeMux) {},
	}

	mode, err := resolveAuthSessionBootstrapMode()
	if err != nil {
		return composition, err
	}
	if mode == authSessionModeDisabled {
		log.Printf("auth/session disabled by MS_AUTH_WEB_SESSION_MODE")
		return composition, nil
	}

	authSessionConfig, err := authsession.ConfigFromEnv(runtime.environment)
	if err != nil {
		if mode == authSessionModeOptional {
			log.Printf("auth/session optional mode fallback: %v", err)
			return composition, nil
		}
		return composition, fmt.Errorf("auth/session required but config is invalid: %w", err)
	}

	tokenValidator, err := platformauth.NewJWTAuthenticatorFromEnv()
	if err != nil {
		if mode == authSessionModeOptional {
			log.Printf("auth/session optional mode fallback: %v", err)
			return composition, nil
		}
		return composition, fmt.Errorf("auth/session required but jwt authenticator init failed: %w", err)
	}

	authSessionStore := authsession.NewStore(runtime.db)
	authSessionService := authsession.NewService(
		authSessionConfig,
		authSessionStore,
		authsession.NewOIDCClient(authSessionConfig),
		tokenValidator,
		authsession.NewRuntimePolicyResolver(governance.featureFlags, governance.thresholds, runtime.environment),
		modules.iamRepo,
		modules.iamAuthorizer,
	)
	authSessionHandler := authsession.NewHandler(authSessionService, authSessionConfig, runtime.allowedOrigins)

	composition.requestAuthenticator = platformauth.NewBearerOrCookieRequestAuthenticator(
		bearerAuthenticator,
		authsession.NewRequestAuthenticator(authSessionConfig, authSessionStore),
	)
	composition.publicPaths = append(composition.publicPaths,
		"/api/v1/auth/session/login",
		"/api/v1/auth/session/callback",
	)
	composition.registerHTTP = authSessionHandler.RegisterRoutes
	return composition, nil
}
