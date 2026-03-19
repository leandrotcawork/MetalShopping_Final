package suppliers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ValidationError struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Registry struct {
	families map[string]func(config map[string]any) []ValidationError
}

func NewDefaultRegistry() *Registry {
	r := &Registry{
		families: map[string]func(config map[string]any) []ValidationError{},
	}
	r.Register("http", validateHTTPFamily)
	r.Register("playwright", validatePlaywrightFamily)
	return r
}

func (r *Registry) Register(family string, validator func(config map[string]any) []ValidationError) {
	key := normalizeFamily(family)
	if key == "" || validator == nil {
		return
	}
	r.families[key] = validator
}

func (r *Registry) Validate(family string, configJSON json.RawMessage) ([]ValidationError, error) {
	key := normalizeFamily(family)
	validator, ok := r.families[key]
	if !ok {
		return []ValidationError{{
			Code:    "UNKNOWN_FAMILY",
			Field:   "family",
			Message: fmt.Sprintf("Unknown driver family: %s", strings.TrimSpace(family)),
		}}, nil
	}

	payload := map[string]any{}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &payload); err != nil {
			return []ValidationError{{
				Code:    "INVALID_JSON",
				Field:   "config",
				Message: "Config must be valid JSON object",
			}}, nil
		}
	}
	return validator(payload), nil
}

func normalizeFamily(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "http", "http_basic", "http_v1":
		return "http"
	case "playwright", "playwright_basic", "playwright_v1":
		return "playwright"
	default:
		return value
	}
}

func validateHTTPFamily(config map[string]any) []ValidationError {
	if hasNonEmptyString(config, "baseUrl") || hasNonEmptyString(config, "endpointTemplate") {
		return nil
	}
	return []ValidationError{{
		Code:    "MISSING_REQUIRED_FIELD",
		Field:   "config.baseUrl",
		Message: "HTTP family requires baseUrl or endpointTemplate",
	}}
}

func validatePlaywrightFamily(config map[string]any) []ValidationError {
	if hasNonEmptyString(config, "startUrl") || hasNonEmptyString(config, "searchUrl") {
		return nil
	}
	return []ValidationError{{
		Code:    "MISSING_REQUIRED_FIELD",
		Field:   "config.startUrl",
		Message: "Playwright family requires startUrl or searchUrl",
	}}
}

func hasNonEmptyString(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}
