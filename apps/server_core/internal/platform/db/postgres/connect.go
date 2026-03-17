package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	if err := validateDSN(cfg.DSN); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func OpenFromEnv(ctx context.Context) (*sql.DB, Config, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, Config{}, err
	}

	db, err := Open(ctx, cfg)
	if err != nil {
		return nil, Config{}, err
	}

	return db, cfg, nil
}
