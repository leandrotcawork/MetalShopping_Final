package runtime_config

import (
	"os"
	"strings"
)

func CORSAllowedOriginsFromEnv(environment string) []string {
	raw := strings.TrimSpace(os.Getenv("APP_CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		if strings.EqualFold(strings.TrimSpace(environment), "local") {
			return []string{
				"http://127.0.0.1:5173",
				"http://localhost:5173",
			}
		}
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	return origins
}
