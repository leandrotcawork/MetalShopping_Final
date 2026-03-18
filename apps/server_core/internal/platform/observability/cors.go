package observability

import (
	"net/http"
	"strings"
)

var defaultCORSAllowedHeaders = []string{
	"Authorization",
	"Accept",
	"Content-Type",
	"X-Trace-Id",
}

var defaultCORSAllowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodOptions,
}

func NewCORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	allowHeaders := strings.Join(defaultCORSAllowedHeaders, ", ")
	allowMethods := strings.Join(defaultCORSAllowedMethods, ", ")

	return func(next http.Handler) http.Handler {
		if len(allowed) == 0 {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowed[origin]; !ok {
				if r.Method == http.MethodOptions {
					http.Error(w, "cors origin not allowed", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			headers := w.Header()
			headers.Set("Vary", "Origin")
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Headers", allowHeaders)
			headers.Set("Access-Control-Allow-Methods", allowMethods)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
