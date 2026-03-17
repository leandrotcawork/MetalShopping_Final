package runtime_config

import (
	"os"
	"strings"
)

func EnvironmentFromEnv() string {
	raw := strings.TrimSpace(os.Getenv("APP_ENV"))
	if raw == "" {
		return "local"
	}
	return raw
}
