package runtime_config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func HTTPAddressFromEnv() (string, error) {
	raw := strings.TrimSpace(os.Getenv("APP_PORT"))
	if raw == "" {
		return ":8080", nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid APP_PORT: %s", raw)
	}

	return ":" + strconv.Itoa(port), nil
}
