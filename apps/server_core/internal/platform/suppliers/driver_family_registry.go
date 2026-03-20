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

const (
	httpStrategyMock           = "http.mock.v1"
	httpStrategyVTEX           = "http.vtex_persisted_query.v1"
	httpStrategyHTMLSearch     = "http.html_search.v1"
	httpStrategyLeroySearch    = "http.leroy_search_sellers.v1"
	httpStrategyHTMLDOMFirst   = "http.html_dom_first_card.v1"
	playwrightStrategyMock     = "playwright.mock.v1"
	playwrightStrategyPDPFirst = "playwright.pdp_first.v1"
)

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
	errors := []ValidationError{}

	strategy := normalizeStrategy(config["strategy"])
	if strategy == "" {
		errors = append(errors, ValidationError{
			Code:    "MISSING_REQUIRED_FIELD",
			Field:   "config.strategy",
			Message: "HTTP family requires strategy",
		})
		return errors
	}

	switch strategy {
	case httpStrategyMock:
		if !hasNonEmptyString(config, "endpointTemplate") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.endpointTemplate",
				Message: "http.mock.v1 requires endpointTemplate",
			})
		}
	case httpStrategyVTEX:
		if !hasNonEmptyString(config, "baseUrl") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.baseUrl",
				Message: "http.vtex_persisted_query.v1 requires baseUrl",
			})
		}
		if !hasNonEmptyString(config, "operationName") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.operationName",
				Message: "http.vtex_persisted_query.v1 requires operationName",
			})
		}
		if !hasNonEmptyString(config, "sha256Hash") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.sha256Hash",
				Message: "http.vtex_persisted_query.v1 requires sha256Hash",
			})
		}
	case httpStrategyHTMLSearch:
		if !hasNonEmptyString(config, "baseUrl") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.baseUrl",
				Message: "http.html_search.v1 requires baseUrl",
			})
		}
		if !hasNonEmptyString(config, "searchUrlTemplate") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.searchUrlTemplate",
				Message: "http.html_search.v1 requires searchUrlTemplate",
			})
		}
	case httpStrategyLeroySearch:
		if !hasNonEmptyString(config, "searchUrlTemplate") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.searchUrlTemplate",
				Message: "http.leroy_search_sellers.v1 requires searchUrlTemplate",
			})
		}
		if !hasNonEmptyString(config, "sellersUrlTemplate") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.sellersUrlTemplate",
				Message: "http.leroy_search_sellers.v1 requires sellersUrlTemplate",
			})
		}
	case httpStrategyHTMLDOMFirst:
		if !hasNonEmptyString(config, "searchUrlTemplate") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.searchUrlTemplate",
				Message: "http.html_dom_first_card.v1 requires searchUrlTemplate",
			})
		}
		if !hasNonEmptyString(config, "cardRootHint") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.cardRootHint",
				Message: "http.html_dom_first_card.v1 requires cardRootHint",
			})
		}
	default:
		errors = append(errors, ValidationError{
			Code:    "UNKNOWN_STRATEGY",
			Field:   "config.strategy",
			Message: fmt.Sprintf("Unknown HTTP strategy: %s", strategy),
		})
	}

	if !hasIntegerInRange(config, "timeoutSeconds", 1, 60) {
		if _, ok := config["timeoutSeconds"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.timeoutSeconds",
				Message: "timeoutSeconds must be integer between 1 and 60",
			})
		}
	}
	if !hasIntegerInRange(config, "maxRetries", 1, 8) {
		if _, ok := config["maxRetries"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.maxRetries",
				Message: "maxRetries must be integer between 1 and 8",
			})
		}
	}
	if !hasIntegerInRange(config, "maxConcurrency", 1, 16) {
		if _, ok := config["maxConcurrency"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.maxConcurrency",
				Message: "maxConcurrency must be integer between 1 and 16",
			})
		}
	}
	if !hasNumberInRange(config, "requestsPerSecond", 0.1, 10.0) {
		if _, ok := config["requestsPerSecond"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.requestsPerSecond",
				Message: "requestsPerSecond must be number between 0.1 and 10.0",
			})
		}
	}
	if !hasIntegerArrayInRangeWhenPresent(config, "retryHttpStatuses", 100, 599, 20) {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.retryHttpStatuses",
			Message: "retryHttpStatuses must be non-empty integer array (100..599) with max 20 items",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "lookupVariableName") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.lookupVariableName",
			Message: "lookupVariableName must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "pricePath") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.pricePath",
			Message: "pricePath must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "sellerPath") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.sellerPath",
			Message: "sellerPath must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "channelPath") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.channelPath",
			Message: "channelPath must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "priceRegex") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.priceRegex",
			Message: "priceRegex must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "sellerRegex") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.sellerRegex",
			Message: "sellerRegex must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "cardRootHint") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.cardRootHint",
			Message: "cardRootHint must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "cardItemHint") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.cardItemHint",
			Message: "cardItemHint must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "titleHint") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.titleHint",
			Message: "titleHint must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "listPriceHint") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.listPriceHint",
			Message: "listPriceHint must be non-empty string when provided",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "calculatedPriceHint") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.calculatedPriceHint",
			Message: "calculatedPriceHint must be non-empty string when provided",
		})
	}
	if !hasAllowedStringWhenPresent(config, "pricePriority", map[string]struct{}{
		"calculated_first": {},
		"sale_first":       {},
	}) {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.pricePriority",
			Message: "pricePriority must be one of: calculated_first|sale_first",
		})
	}
	if !hasNonEmptyStringWhenPresent(config, "region") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.region",
			Message: "region must be non-empty string when provided",
		})
	}
	if !hasAllowedStringWhenPresent(config, "sellerPickStrategy", map[string]struct{}{
		"selected": {},
		"min_sale": {},
	}) {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.sellerPickStrategy",
			Message: "sellerPickStrategy must be one of: selected|min_sale",
		})
	}
	if !hasObjectWhenPresent(config, "headers") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.headers",
			Message: "headers must be object when provided",
		})
	}
	if !hasObjectWhenPresent(config, "extraVariables") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.extraVariables",
			Message: "extraVariables must be object when provided",
		})
	}

	return errors
}

func validatePlaywrightFamily(config map[string]any) []ValidationError {
	errors := []ValidationError{}

	strategy := normalizeStrategy(config["strategy"])
	if strategy == "" {
		errors = append(errors, ValidationError{
			Code:    "MISSING_REQUIRED_FIELD",
			Field:   "config.strategy",
			Message: "Playwright family requires strategy",
		})
		return errors
	}

	switch strategy {
	case playwrightStrategyMock:
		if !hasNonEmptyString(config, "startUrl") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.startUrl",
				Message: "playwright.mock.v1 requires startUrl",
			})
		}
	case playwrightStrategyPDPFirst:
		if !hasNonEmptyObject(config, "pdpSelectors") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.pdpSelectors",
				Message: "playwright.pdp_first.v1 requires pdpSelectors",
			})
		}
		if !hasNonEmptyString(config, "startUrl") && !hasNonEmptyString(config, "searchUrl") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.startUrl",
				Message: "playwright.pdp_first.v1 requires startUrl or searchUrl",
			})
		}
		if !hasNonEmptyStringInObject(config, "pdpSelectors", "price") {
			errors = append(errors, ValidationError{
				Code:    "MISSING_REQUIRED_FIELD",
				Field:   "config.pdpSelectors.price",
				Message: "playwright.pdp_first.v1 requires pdpSelectors.price",
			})
		}
	default:
		errors = append(errors, ValidationError{
			Code:    "UNKNOWN_STRATEGY",
			Field:   "config.strategy",
			Message: fmt.Sprintf("Unknown PLAYWRIGHT strategy: %s", strategy),
		})
	}

	if !hasIntegerInRange(config, "timeoutSeconds", 1, 120) {
		if _, ok := config["timeoutSeconds"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.timeoutSeconds",
				Message: "timeoutSeconds must be integer between 1 and 120",
			})
		}
	}
	if !hasIntegerInRange(config, "maxRetries", 1, 8) {
		if _, ok := config["maxRetries"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.maxRetries",
				Message: "maxRetries must be integer between 1 and 8",
			})
		}
	}
	if !hasIntegerInRange(config, "tabs", 1, 10) {
		if _, ok := config["tabs"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.tabs",
				Message: "tabs must be integer between 1 and 10",
			})
		}
	}
	if !hasIntegerInRange(config, "circuitBreakerThreshold", 1, 10) {
		if _, ok := config["circuitBreakerThreshold"]; ok {
			errors = append(errors, ValidationError{
				Code:    "INVALID_FIELD",
				Field:   "config.circuitBreakerThreshold",
				Message: "circuitBreakerThreshold must be integer between 1 and 10",
			})
		}
	}
	if !hasNonEmptyStringWhenPresent(config, "searchUrlTemplate") {
		errors = append(errors, ValidationError{
			Code:    "INVALID_FIELD",
			Field:   "config.searchUrlTemplate",
			Message: "searchUrlTemplate must be non-empty string when provided",
		})
	}

	return errors
}

func hasNonEmptyString(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}

func hasNonEmptyObject(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	obj, ok := value.(map[string]any)
	return ok && len(obj) > 0
}

func hasObjectWhenPresent(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	_, ok = value.(map[string]any)
	return ok
}

func hasIntegerInRange(payload map[string]any, key string, minValue, maxValue int) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	switch typed := value.(type) {
	case float64:
		integer := int(typed)
		return float64(integer) == typed && integer >= minValue && integer <= maxValue
	case int:
		return typed >= minValue && typed <= maxValue
	default:
		return false
	}
}

func hasNonEmptyStringWhenPresent(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}

func hasAllowedStringWhenPresent(payload map[string]any, key string, allowed map[string]struct{}) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	text, ok := value.(string)
	if !ok {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}
	_, exists := allowed[normalized]
	return exists
}

func hasNumberInRange(payload map[string]any, key string, minValue, maxValue float64) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	switch typed := value.(type) {
	case float64:
		return typed >= minValue && typed <= maxValue
	case float32:
		number := float64(typed)
		return number >= minValue && number <= maxValue
	case int:
		number := float64(typed)
		return number >= minValue && number <= maxValue
	default:
		return false
	}
}

func hasIntegerArrayInRangeWhenPresent(payload map[string]any, key string, minValue, maxValue, maxItems int) bool {
	value, ok := payload[key]
	if !ok {
		return true
	}
	list, ok := value.([]any)
	if !ok || len(list) == 0 || len(list) > maxItems {
		return false
	}
	for _, item := range list {
		switch typed := item.(type) {
		case float64:
			integer := int(typed)
			if float64(integer) != typed || integer < minValue || integer > maxValue {
				return false
			}
		case int:
			if typed < minValue || typed > maxValue {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func normalizeStrategy(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(text))
}

func hasNonEmptyStringInObject(payload map[string]any, objectKey, key string) bool {
	value, ok := payload[objectKey]
	if !ok {
		return false
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return false
	}
	entry, ok := obj[key]
	if !ok {
		return false
	}
	text, ok := entry.(string)
	return ok && strings.TrimSpace(text) != ""
}
