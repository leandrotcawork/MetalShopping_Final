package postgres

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultPingTimeout     = 5 * time.Second
	defaultSSLMode         = "require"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

func LoadConfigFromEnv() (Config, error) {
	dsn, err := loadDSNFromEnv()
	if err != nil {
		return Config{}, err
	}

	maxOpenConns, err := loadIntFromEnv("MS_PG_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := loadIntFromEnv("MS_PG_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	connMaxLifetime, err := loadDurationSecondsFromEnv("MS_PG_CONN_MAX_LIFETIME_SECONDS", defaultConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}
	connMaxIdleTime, err := loadDurationSecondsFromEnv("MS_PG_CONN_MAX_IDLE_TIME_SECONDS", defaultConnMaxIdleTime)
	if err != nil {
		return Config{}, err
	}
	pingTimeout, err := loadDurationSecondsFromEnv("MS_PG_PING_TIMEOUT_SECONDS", defaultPingTimeout)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DSN:             dsn,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		PingTimeout:     pingTimeout,
	}, nil
}

func loadDSNFromEnv() (string, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn != "" {
		if err := validateDSN(dsn); err != nil {
			return "", err
		}
		return dsn, nil
	}

	host := strings.TrimSpace(os.Getenv("PGHOST"))
	port := strings.TrimSpace(os.Getenv("PGPORT"))
	database := strings.TrimSpace(os.Getenv("PGDATABASE"))
	user := strings.TrimSpace(os.Getenv("PGUSER"))
	password := os.Getenv("PGPASSWORD")
	sslMode := strings.TrimSpace(os.Getenv("PGSSLMODE"))
	if sslMode == "" {
		sslMode = defaultSSLMode
	}

	if host == "" || database == "" || user == "" || password == "" {
		return "", fmt.Errorf("postgres config missing: set DATABASE_URL or PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD")
	}
	if port == "" {
		port = "5432"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   database,
	}
	query := u.Query()
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	dsn = u.String()
	if err := validateDSN(dsn); err != nil {
		return "", err
	}
	return dsn, nil
}

func validateDSN(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid postgres dsn: %w", err)
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("invalid postgres dsn scheme: %s", u.Scheme)
	}
	if strings.TrimSpace(u.Host) == "" {
		return fmt.Errorf("invalid postgres dsn host")
	}
	if strings.TrimSpace(strings.TrimPrefix(u.Path, "/")) == "" {
		return fmt.Errorf("invalid postgres dsn database")
	}
	return nil
}

func loadIntFromEnv(key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s: %s", key, raw)
	}
	return value, nil
}

func loadDurationSecondsFromEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("invalid %s: %s", key, raw)
	}
	return time.Duration(seconds) * time.Second, nil
}
