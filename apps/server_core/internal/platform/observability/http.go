package observability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type ReadinessChecker func(context.Context) error

func NewRequestLoggingMiddleware() func(http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id"))
			if traceID == "" {
				traceID = "trace-local"
			}

			logger.Info("http_request",
				"trace_id", traceID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

func NewLiveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
		})
	})
}

func NewReadyHandler(checker ReadinessChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checker == nil {
			writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
			return
		}

		if err := checker(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
