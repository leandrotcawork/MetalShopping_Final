package main

import (
	"context"
	"database/sql"
	"fmt"

	platformauth "metalshopping/server_core/internal/platform/auth"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/runtime_config"
)

type runtimeComposition struct {
	addr           string
	environment    string
	db             *sql.DB
	authenticator  platformauth.Authenticator
	allowedOrigins []string
}

func composeRuntime(ctx context.Context) (runtimeComposition, error) {
	addr, err := runtime_config.HTTPAddressFromEnv()
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("load server config: %w", err)
	}

	environment := runtime_config.EnvironmentFromEnv()
	db, _, err := pgdb.OpenFromEnv(ctx)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("open postgres: %w", err)
	}

	authenticator, err := platformauth.NewAuthenticatorFromEnv()
	if err != nil {
		_ = db.Close()
		return runtimeComposition{}, fmt.Errorf("load bootstrap authenticator: %w", err)
	}

	return runtimeComposition{
		addr:           addr,
		environment:    environment,
		db:             db,
		authenticator:  authenticator,
		allowedOrigins: runtime_config.CORSAllowedOriginsFromEnv(environment),
	}, nil
}
