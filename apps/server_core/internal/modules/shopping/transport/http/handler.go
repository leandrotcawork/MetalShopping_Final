package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/shopping/adapters/postgres"
	"metalshopping/server_core/internal/modules/shopping/application"
	"metalshopping/server_core/internal/modules/shopping/ports"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/shopping/bootstrap", h.handleBootstrap)
	mux.HandleFunc("/api/v1/shopping/summary", h.handleSummary)
	mux.HandleFunc("/api/v1/shopping/runs", h.handleRunsList)
	mux.HandleFunc("/api/v1/shopping/runs/", h.handleRunByID)
	mux.HandleFunc("/api/v1/shopping/products/", h.handleProductLatest)
	mux.HandleFunc("/api/v1/shopping/run-requests/", h.handleRunRequestByID)
	mux.HandleFunc("/api/v1/shopping/supplier-signals", h.handleSupplierSignals)
	mux.HandleFunc("/api/v1/shopping/manual-url-candidates", h.handleManualURLCandidates)
}

func (h *Handler) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_bootstrap", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	bootstrap, err := h.service.GetBootstrap(r.Context(), tenantID)
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping bootstrap", traceID)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"inputModes":         bootstrap.InputModes,
		"runStatuses":        bootstrap.RunStatuses,
		"supportsManualUrls": bootstrap.SupportsManual,
		"advancedDefaults": map[string]any{
			"timeoutSeconds":    bootstrap.AdvancedDefaults.TimeoutSeconds,
			"httpWorkers":       bootstrap.AdvancedDefaults.HTTPWorkers,
			"playwrightWorkers": bootstrap.AdvancedDefaults.PlaywrightWorker,
			"topN":              bootstrap.AdvancedDefaults.TopN,
		},
		"suppliers": mapSuppliers(bootstrap.Suppliers),
	})
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_summary", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	summary, err := h.service.GetSummary(r.Context(), tenantID)
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping summary", traceID)
		return
	}

	lastRunAt := time.Now().UTC()
	if summary.LastRunAt != nil {
		lastRunAt = summary.LastRunAt.UTC()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalRuns":     summary.TotalRuns,
		"runningRuns":   summary.RunningRuns,
		"completedRuns": summary.CompletedRuns,
		"failedRuns":    summary.FailedRuns,
		"lastRunAt":     lastRunAt.Format(time.RFC3339),
	})
}

func (h *Handler) handleRunsList(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusCreated
	reqResult := "success"
	action := "shopping.create_run_request"
	if r.Method == http.MethodGet {
		statusCode = http.StatusOK
		action = "shopping.list_runs"
	}
	defer logRequest(action, traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	if r.Method == http.MethodPost {
		var requestBody shoppingCreateRunRequestBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&requestBody); err != nil {
			statusCode = http.StatusBadRequest
			reqResult = "validation_error"
			writeShoppingError(w, http.StatusBadRequest, "SHOPPING_RUN_REQUEST_INVALID", "Invalid shopping run request payload", traceID)
			return
		}

		principal, principalOK := platformauth.PrincipalFromContext(r.Context())
		if !principalOK {
			statusCode = http.StatusUnauthorized
			reqResult = "auth_or_tenant_error"
			writeShoppingError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
			return
		}
		requestedBy := strings.TrimSpace(principal.Email)
		if requestedBy == "" {
			requestedBy = strings.TrimSpace(principal.SubjectID)
		}
		if requestedBy == "" {
			requestedBy = "unknown"
		}

		created, err := h.service.CreateRunRequest(r.Context(), tenantID, traceID, ports.CreateRunRequestInput{
			InputMode:         requestBody.InputMode,
			CatalogProductIDs: requestBody.CatalogProductIDs,
			XLSXFilePath:      requestBody.XLSXFilePath,
			XLSXScopeIDs:      requestBody.XLSXScopeIdentifiers,
			SupplierCodes:     requestBody.SupplierCodes,
			Advanced: ports.AdvancedDefaults{
				TimeoutSeconds:   requestBody.Advanced.TimeoutSeconds,
				HTTPWorkers:      requestBody.Advanced.HTTPWorkers,
				PlaywrightWorker: requestBody.Advanced.PlaywrightWorkers,
				TopN:             requestBody.Advanced.TopN,
			},
			Notes:       requestBody.Notes,
			RequestedBy: requestedBy,
		})
		if err != nil {
			statusCode = http.StatusBadRequest
			reqResult = "validation_error"
			writeShoppingError(w, http.StatusBadRequest, "SHOPPING_RUN_REQUEST_INVALID", err.Error(), traceID)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"runRequestId": created.RunRequestID,
			"status":       created.Status,
			"inputMode":    created.InputMode,
			"requestedAt":  created.RequestedAt.Format(time.RFC3339),
			"requestedBy":  created.RequestedBy,
		})
		return
	}

	limit := parseQueryInt64(r, "limit", 50)
	offset := parseQueryInt64(r, "offset", 0)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	runList, err := h.service.ListRuns(r.Context(), tenantID, ports.RunListFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list shopping runs", traceID)
		return
	}

	rows := make([]map[string]any, 0, len(runList.Rows))
	for _, run := range runList.Rows {
		rows = append(rows, mapRun(run))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rows": rows,
		"paging": map[string]any{
			"offset":   runList.Offset,
			"limit":    runList.Limit,
			"returned": len(runList.Rows),
			"total":    runList.Total,
		},
	})
}

type shoppingCreateRunRequestBody struct {
	InputMode            string   `json:"inputMode"`
	CatalogProductIDs    []string `json:"catalogProductIds"`
	XLSXFilePath         string   `json:"xlsxFilePath"`
	XLSXScopeIdentifiers []string `json:"xlsxScopeIdentifiers"`
	SupplierCodes        []string `json:"supplierCodes"`
	Advanced             struct {
		TimeoutSeconds    int64 `json:"timeoutSeconds"`
		HTTPWorkers       int64 `json:"httpWorkers"`
		PlaywrightWorkers int64 `json:"playwrightWorkers"`
		TopN              int64 `json:"topN"`
	} `json:"advanced"`
	Notes string `json:"notes"`
}

type shoppingUpsertSupplierSignalBody struct {
	ProductID      string  `json:"productId"`
	SupplierCode   string  `json:"supplierCode"`
	ProductURL     *string `json:"productUrl"`
	URLStatus      *string `json:"urlStatus"`
	LookupMode     *string `json:"lookupMode"`
	ManualOverride *bool   `json:"manualOverride"`
}

func (h *Handler) handleRunByID(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_run", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	runID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/v1/shopping/runs/"))
	if runID == "" || strings.Contains(runID, "/") {
		statusCode = http.StatusNotFound
		reqResult = "not_found"
		writeShoppingError(w, http.StatusNotFound, "SHOPPING_RUN_NOT_FOUND", "Shopping run not found", traceID)
		return
	}

	run, err := h.service.GetRun(r.Context(), tenantID, runID)
	if err != nil {
		if errors.Is(err, postgres.ErrRunNotFound) {
			statusCode = http.StatusNotFound
			reqResult = "not_found"
			writeShoppingError(w, http.StatusNotFound, "SHOPPING_RUN_NOT_FOUND", "Shopping run not found", traceID)
			return
		}
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping run", traceID)
		return
	}

	writeJSON(w, http.StatusOK, mapRun(run))
}

func (h *Handler) handleRunRequestByID(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_run_request", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	runRequestID := strings.TrimPrefix(r.URL.Path, "/api/v1/shopping/run-requests/")
	runRequestID = strings.TrimSpace(runRequestID)
	if runRequestID == "" || strings.Contains(runRequestID, "/") {
		statusCode = http.StatusBadRequest
		reqResult = "validation_error"
		writeShoppingError(w, http.StatusBadRequest, "SHOPPING_RUN_REQUEST_INVALID", "Missing run_request_id", traceID)
		return
	}

	runRequest, err := h.service.GetRunRequest(r.Context(), tenantID, runRequestID)
	if err != nil {
		if errors.Is(err, postgres.ErrRunRequestNotFound) {
			statusCode = http.StatusNotFound
			reqResult = "not_found"
			writeShoppingError(w, http.StatusNotFound, "SHOPPING_RUN_REQUEST_NOT_FOUND", "Run request not found", traceID)
			return
		}
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping run request", traceID)
		return
	}

	payload := map[string]any{
		"runRequestId":               runRequest.RunRequestID,
		"status":                     runRequest.Status,
		"inputMode":                  runRequest.InputMode,
		"requestedAt":                runRequest.RequestedAt.Format(time.RFC3339),
		"requestedBy":                runRequest.RequestedBy,
		"catalogProductIds":          runRequest.CatalogProductIDs,
		"xlsxScopeIdentifiers":       runRequest.XLSXScopeIDs,
		"resolvedCatalogProductIds":  runRequest.ResolvedCatalogProductIDs,
		"unresolvedScopeIdentifiers": runRequest.UnresolvedScopeIDs,
		"ambiguousScopeIdentifiers":  runRequest.AmbiguousScopeIDs,
		"claimedAt":                  nil,
		"startedAt":                  nil,
		"finishedAt":                 nil,
		"workerId":                   nil,
		"runId":                      nil,
		"errorMessage":               nil,
	}
	if runRequest.ClaimedAt != nil {
		payload["claimedAt"] = runRequest.ClaimedAt.UTC().Format(time.RFC3339)
	}
	if runRequest.StartedAt != nil {
		payload["startedAt"] = runRequest.StartedAt.UTC().Format(time.RFC3339)
	}
	if runRequest.FinishedAt != nil {
		payload["finishedAt"] = runRequest.FinishedAt.UTC().Format(time.RFC3339)
	}
	if runRequest.WorkerID != nil {
		payload["workerId"] = *runRequest.WorkerID
	}
	if runRequest.RunID != nil {
		payload["runId"] = *runRequest.RunID
	}
	if runRequest.ErrorMessage != nil {
		payload["errorMessage"] = *runRequest.ErrorMessage
	}
	if runRequest.XLSXFilePath != nil {
		payload["xlsxFilePath"] = *runRequest.XLSXFilePath
	} else {
		payload["xlsxFilePath"] = nil
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) handleProductLatest(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_product_latest", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	productID, routeOK := extractProductLatestPathParam(r.URL.Path)
	if !routeOK {
		statusCode = http.StatusNotFound
		reqResult = "not_found"
		writeShoppingError(w, http.StatusNotFound, "SHOPPING_PRODUCT_LATEST_NOT_FOUND", "Shopping latest snapshot not found", traceID)
		return
	}

	item, err := h.service.GetProductLatest(r.Context(), tenantID, productID)
	if err != nil {
		if errors.Is(err, postgres.ErrProductLatestNotFound) {
			statusCode = http.StatusNotFound
			reqResult = "not_found"
			writeShoppingError(w, http.StatusNotFound, "SHOPPING_PRODUCT_LATEST_NOT_FOUND", "Shopping latest snapshot not found", traceID)
			return
		}
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping latest snapshot", traceID)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"productId":     item.ProductID,
		"runId":         item.RunID,
		"observedAt":    item.ObservedAt.Format(time.RFC3339),
		"sellerName":    item.SellerName,
		"channel":       item.Channel,
		"observedPrice": item.ObservedPrice,
		"currency":      item.Currency,
	})
}

func (h *Handler) handleSupplierSignals(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	action := "shopping.list_supplier_signals"
	if r.Method == http.MethodPut {
		action = "shopping.upsert_supplier_signal"
	}
	defer logRequest(action, traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	if r.Method == http.MethodGet {
		limit := parseQueryInt64(r, "limit", 50)
		offset := parseQueryInt64(r, "offset", 0)
		supplierCode := strings.TrimSpace(r.URL.Query().Get("supplier_code"))
		productID := strings.TrimSpace(r.URL.Query().Get("product_id"))

		signals, err := h.service.ListSupplierSignals(r.Context(), tenantID, ports.SupplierSignalListFilter{
			SupplierCode: supplierCode,
			ProductID:    productID,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil {
			statusCode = http.StatusInternalServerError
			reqResult = "internal_error"
			writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list shopping supplier signals", traceID)
			return
		}

		rows := make([]map[string]any, 0, len(signals.Rows))
		for _, item := range signals.Rows {
			rows = append(rows, mapSupplierSignal(item))
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"rows": rows,
			"paging": map[string]any{
				"offset":   signals.Offset,
				"limit":    signals.Limit,
				"returned": len(rows),
				"total":    signals.Total,
			},
		})
		return
	}

	var requestBody shoppingUpsertSupplierSignalBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestBody); err != nil {
		statusCode = http.StatusBadRequest
		reqResult = "validation_error"
		writeShoppingError(w, http.StatusBadRequest, "SHOPPING_SUPPLIER_SIGNAL_INVALID", "Invalid supplier signal payload", traceID)
		return
	}

	principal, principalOK := platformauth.PrincipalFromContext(r.Context())
	if !principalOK {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		writeShoppingError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
		return
	}
	updatedBy := strings.TrimSpace(principal.Email)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(principal.SubjectID)
	}
	if updatedBy == "" {
		updatedBy = "unknown"
	}

	signal, err := h.service.UpsertSupplierSignal(r.Context(), tenantID, ports.UpsertSupplierSignalInput{
		ProductID:      requestBody.ProductID,
		SupplierCode:   requestBody.SupplierCode,
		ProductURL:     requestBody.ProductURL,
		URLStatus:      requestBody.URLStatus,
		LookupMode:     requestBody.LookupMode,
		ManualOverride: requestBody.ManualOverride,
		UpdatedBy:      updatedBy,
	})
	if err != nil {
		statusCode = http.StatusBadRequest
		reqResult = "validation_error"
		writeShoppingError(w, http.StatusBadRequest, "SHOPPING_SUPPLIER_SIGNAL_INVALID", err.Error(), traceID)
		return
	}

	writeJSON(w, http.StatusOK, mapSupplierSignal(signal))
}

func (h *Handler) handleManualURLCandidates(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.list_manual_url_candidates", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	supplierCode := strings.TrimSpace(r.URL.Query().Get("supplier_code"))
	if supplierCode == "" {
		statusCode = http.StatusBadRequest
		reqResult = "validation_error"
		writeShoppingError(w, http.StatusBadRequest, "SHOPPING_MANUAL_URL_CANDIDATES_INVALID", "supplier_code is required", traceID)
		return
	}

	limit := parseQueryInt64(r, "limit", 50)
	offset := parseQueryInt64(r, "offset", 0)
	includeExisting := r.URL.Query().Get("include_existing")

	parsedIncludeExisting := true
	if strings.TrimSpace(includeExisting) != "" {
		value, err := strconv.ParseBool(includeExisting)
		if err != nil {
			statusCode = http.StatusBadRequest
			reqResult = "validation_error"
			writeShoppingError(w, http.StatusBadRequest, "SHOPPING_MANUAL_URL_CANDIDATES_INVALID", "include_existing must be boolean", traceID)
			return
		}
		parsedIncludeExisting = value
	}

	candidates, err := h.service.ListManualURLCandidates(r.Context(), tenantID, ports.ManualURLCandidateFilter{
		SupplierCode:      supplierCode,
		Search:            strings.TrimSpace(r.URL.Query().Get("search")),
		BrandName:         strings.TrimSpace(r.URL.Query().Get("brand_name")),
		TaxonomyLeaf0Name: strings.TrimSpace(r.URL.Query().Get("taxonomy_leaf0_name")),
		IncludeExisting:   parsedIncludeExisting,
		Limit:             limit,
		Offset:            offset,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list shopping manual URL candidates", traceID)
		return
	}

	rows := make([]map[string]any, 0, len(candidates.Rows))
	for _, item := range candidates.Rows {
		rows = append(rows, mapManualURLCandidate(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rows": rows,
		"paging": map[string]any{
			"offset":   candidates.Offset,
			"limit":    candidates.Limit,
			"returned": len(rows),
			"total":    candidates.Total,
		},
	})
}

func mapRun(run ports.Run) map[string]any {
	var finishedAt any
	if run.FinishedAt != nil {
		finishedAt = run.FinishedAt.Format(time.RFC3339)
	}
	return map[string]any{
		"runId":          run.RunID,
		"status":         run.Status,
		"startedAt":      run.StartedAt.Format(time.RFC3339),
		"finishedAt":     finishedAt,
		"processedItems": run.ProcessedItems,
		"totalItems":     run.TotalItems,
		"notes":          run.Notes,
	}
}

func mapSuppliers(items []ports.BootstrapSupplier) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"supplierCode":  item.SupplierCode,
			"supplierLabel": item.SupplierLabel,
			"executionKind": item.ExecutionKind,
			"lookupPolicy":  item.LookupPolicy,
			"enabled":       item.Enabled,
		})
	}
	return result
}

func mapSupplierSignal(item ports.SupplierSignal) map[string]any {
	payload := map[string]any{
		"productId":        item.ProductID,
		"supplierCode":     item.SupplierCode,
		"productUrl":       nil,
		"urlStatus":        item.URLStatus,
		"lookupMode":       item.LookupMode,
		"lookupModeSource": item.LookupModeSource,
		"manualOverride":   item.ManualOverride,
		"lastCheckedAt":    nil,
		"lastSuccessAt":    nil,
		"lastHttpStatus":   nil,
		"lastErrorMessage": nil,
		"nextDiscoveryAt":  nil,
		"notFoundCount":    item.NotFoundCount,
		"updatedAt":        item.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if item.ProductURL != nil {
		payload["productUrl"] = *item.ProductURL
	}
	if item.LastCheckedAt != nil {
		payload["lastCheckedAt"] = item.LastCheckedAt.UTC().Format(time.RFC3339)
	}
	if item.LastSuccessAt != nil {
		payload["lastSuccessAt"] = item.LastSuccessAt.UTC().Format(time.RFC3339)
	}
	if item.LastHTTPStatus != nil {
		payload["lastHttpStatus"] = *item.LastHTTPStatus
	}
	if item.LastErrorMessage != nil {
		payload["lastErrorMessage"] = *item.LastErrorMessage
	}
	if item.NextDiscoveryAt != nil {
		payload["nextDiscoveryAt"] = item.NextDiscoveryAt.UTC().Format(time.RFC3339)
	}
	return payload
}

func mapManualURLCandidate(item ports.ManualURLCandidate) map[string]any {
	payload := map[string]any{
		"productId":         item.ProductID,
		"supplierCode":      item.SupplierCode,
		"sku":               item.SKU,
		"name":              item.Name,
		"brandName":         nil,
		"taxonomyLeaf0Name": nil,
		"productUrl":        nil,
		"urlStatus":         item.URLStatus,
		"lookupMode":        item.LookupMode,
		"lookupModeSource":  item.LookupModeSource,
		"manualOverride":    item.ManualOverride,
		"lastCheckedAt":     nil,
		"lastSuccessAt":     nil,
		"lastHttpStatus":    nil,
		"lastErrorMessage":  nil,
		"nextDiscoveryAt":   nil,
		"notFoundCount":     item.NotFoundCount,
		"updatedAt":         item.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if item.BrandName != nil {
		payload["brandName"] = *item.BrandName
	}
	if item.TaxonomyLeaf0Name != nil {
		payload["taxonomyLeaf0Name"] = *item.TaxonomyLeaf0Name
	}
	if item.ProductURL != nil {
		payload["productUrl"] = *item.ProductURL
	}
	if item.LastCheckedAt != nil {
		payload["lastCheckedAt"] = item.LastCheckedAt.UTC().Format(time.RFC3339)
	}
	if item.LastSuccessAt != nil {
		payload["lastSuccessAt"] = item.LastSuccessAt.UTC().Format(time.RFC3339)
	}
	if item.LastHTTPStatus != nil {
		payload["lastHttpStatus"] = *item.LastHTTPStatus
	}
	if item.LastErrorMessage != nil {
		payload["lastErrorMessage"] = *item.LastErrorMessage
	}
	if item.NextDiscoveryAt != nil {
		payload["nextDiscoveryAt"] = item.NextDiscoveryAt.UTC().Format(time.RFC3339)
	}
	return payload
}

func extractProductLatestPathParam(path string) (string, bool) {
	prefix := "/api/v1/shopping/products/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	suffix := strings.TrimPrefix(path, prefix)
	if !strings.HasSuffix(suffix, "/latest") {
		return "", false
	}
	productID := strings.TrimSuffix(suffix, "/latest")
	productID = strings.TrimSpace(strings.TrimSuffix(productID, "/"))
	if productID == "" || strings.Contains(productID, "/") {
		return "", false
	}
	return productID, true
}

func parseQueryInt64(r *http.Request, key string, fallback int64) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func authenticatedTenantID(w http.ResponseWriter, r *http.Request) (string, bool) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeShoppingError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeShoppingError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return "", false
	}
	return tenant.ID, true
}

type shoppingErrorEnvelope struct {
	Error shoppingError `json:"error"`
}

type shoppingError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func requestTraceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id")); traceID != "" {
		return traceID
	}
	return "trace-local"
}

func writeShoppingError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, shoppingErrorEnvelope{
		Error: shoppingError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func logRequest(action, traceID string, statusCode *int, result *string, startedAt time.Time) {
	durationMs := time.Since(startedAt).Milliseconds()
	slog.Info("shopping_request",
		"action", action,
		"trace_id", traceID,
		"result", *result,
		"status", *statusCode,
		"duration_ms", durationMs,
	)
}
