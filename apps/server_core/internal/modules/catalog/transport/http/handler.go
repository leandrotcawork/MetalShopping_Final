package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/catalog/ports"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service                  *application.Service
	productsPortfolioService *catalogreadmodel.ProductsPortfolioService
	permissionChecker        ports.PermissionChecker
}

type CreateProductRequest struct {
	SKU                   string                           `json:"sku"`
	Name                  string                           `json:"name"`
	Description           string                           `json:"description,omitempty"`
	BrandName             string                           `json:"brand_name,omitempty"`
	StockProfileCode      string                           `json:"stock_profile_code,omitempty"`
	PrimaryTaxonomyNodeID string                           `json:"primary_taxonomy_node_id,omitempty"`
	Status                string                           `json:"status,omitempty"`
	Identifiers           []CreateProductIdentifierRequest `json:"identifiers,omitempty"`
}

type CreateProductIdentifierRequest struct {
	IdentifierType  string `json:"identifier_type"`
	IdentifierValue string `json:"identifier_value"`
	SourceSystem    string `json:"source_system,omitempty"`
	IsPrimary       bool   `json:"is_primary,omitempty"`
}

func NewHandler(service *application.Service, productsPortfolioService *catalogreadmodel.ProductsPortfolioService, permissionChecker ports.PermissionChecker) *Handler {
	return &Handler{
		service:                  service,
		productsPortfolioService: productsPortfolioService,
		permissionChecker:        permissionChecker,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/catalog/products", h.handleProducts)
	mux.HandleFunc("/api/v1/catalog/products/", h.handleProductSubresources)
	mux.HandleFunc("/api/v1/catalog/taxonomy/nodes", h.handleTaxonomyNodes)
	mux.HandleFunc("/api/v1/catalog/taxonomy/levels", h.handleTaxonomyLevels)
	mux.HandleFunc("/api/v1/products/portfolio", h.handleProductsPortfolio)
}

func (h *Handler) handleProducts(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handleCreateProduct(w, r, principal.SubjectID, tenant.ID)
	case http.MethodGet:
		h.handleListProducts(w, r, principal.SubjectID, tenant.ID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleProductSubresources(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/catalog/products/")
	productID, subresource, found := strings.Cut(path, "/")
	if !found || strings.TrimSpace(productID) == "" || subresource != "identifiers" {
		http.NotFound(w, r)
		return
	}

	h.handleListProductIdentifiers(w, r, principal.SubjectID, tenant.ID, strings.TrimSpace(productID))
}

func (h *Handler) handleTaxonomyNodes(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	allowed, err := h.permissionChecker.HasPermission(r.Context(), principal.SubjectID, iamdomain.PermCatalogRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	filter, err := taxonomyNodeFilterFromRequest(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		return
	}

	nodes, err := h.service.ListTaxonomyNodes(r.Context(), tenant.ID, filter)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list taxonomy nodes", requestTraceID(r))
		return
	}

	items := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, mapTaxonomyNode(node))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleTaxonomyLevels(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	allowed, err := h.permissionChecker.HasPermission(r.Context(), principal.SubjectID, iamdomain.PermCatalogRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	defs, err := h.service.ListTaxonomyLevelDefs(r.Context(), tenant.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list taxonomy levels", requestTraceID(r))
		return
	}

	items := make([]map[string]any, 0, len(defs))
	for _, def := range defs {
		items = append(items, mapTaxonomyLevelDef(def))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleProductsPortfolio(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.productsPortfolioService == nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Products portfolio read surface is not configured", requestTraceID(r))
		return
	}

	allowed, err := h.hasAllPermissions(r, principal.SubjectID, iamdomain.PermCatalogRead, iamdomain.PermPricingRead, iamdomain.PermInventoryRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	filter, err := productsPortfolioFilterFromRequest(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		return
	}

	result, err := h.productsPortfolioService.ListProductsPortfolio(r.Context(), tenant.ID, filter)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list products portfolio", requestTraceID(r))
		return
	}

	writeJSON(w, http.StatusOK, mapProductsPortfolio(result))
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request, userID, tenantID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermCatalogWrite)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	product, err := h.service.CreateProduct(r.Context(), application.CreateProductCommand{
		TenantID:              tenantID,
		TraceID:               requestTraceID(r),
		SKU:                   req.SKU,
		Name:                  req.Name,
		Description:           req.Description,
		BrandName:             req.BrandName,
		StockProfileCode:      req.StockProfileCode,
		PrimaryTaxonomyNodeID: req.PrimaryTaxonomyNodeID,
		Status:                req.Status,
		Identifiers:           mapCreateProductIdentifierInputs(req.Identifiers),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTenantIDRequired), errors.Is(err, domain.ErrSKURequired), errors.Is(err, domain.ErrProductNameRequired), errors.Is(err, domain.ErrProductDescriptionTooLong), errors.Is(err, domain.ErrInvalidProductStatus), errors.Is(err, domain.ErrProductIDRequired), errors.Is(err, domain.ErrProductIdentifierTypeRequired), errors.Is(err, domain.ErrProductIdentifierValueRequired):
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		case errors.Is(err, domain.ErrProductCreationDisabled):
			writeAPIError(w, http.StatusForbidden, "GOVERNANCE_DISABLED", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product", requestTraceID(r))
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapProduct(product))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request, userID, tenantID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermCatalogRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	products, err := h.service.ListProducts(r.Context(), tenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list products", requestTraceID(r))
		return
	}

	items := make([]map[string]any, 0, len(products))
	for _, product := range products {
		items = append(items, mapProduct(product))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleListProductIdentifiers(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermCatalogRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	identifiers, err := h.service.ListProductIdentifiers(r.Context(), tenantID, productID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list product identifiers", requestTraceID(r))
		return
	}

	items := make([]map[string]any, 0, len(identifiers))
	for _, identifier := range identifiers {
		items = append(items, mapProductIdentifier(identifier))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) requirePrincipalAndTenant(w http.ResponseWriter, r *http.Request) (platformauth.Principal, tenancy_runtime.Tenant, bool) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return platformauth.Principal{}, tenancy_runtime.Tenant{}, false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return platformauth.Principal{}, tenancy_runtime.Tenant{}, false
	}
	return principal, tenant, true
}

func (h *Handler) hasAllPermissions(r *http.Request, subjectID string, permissions ...iamdomain.Permission) (bool, error) {
	for _, permission := range permissions {
		allowed, err := h.permissionChecker.HasPermission(r.Context(), subjectID, permission)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

func taxonomyNodeFilterFromRequest(r *http.Request) (ports.TaxonomyNodeFilter, error) {
	filter := ports.TaxonomyNodeFilter{
		ParentTaxonomyNodeID: strings.TrimSpace(r.URL.Query().Get("parent_taxonomy_node_id")),
	}
	levelRaw := strings.TrimSpace(r.URL.Query().Get("level"))
	if levelRaw == "" {
		return filter, nil
	}

	level, err := strconv.Atoi(levelRaw)
	if err != nil || level < 0 {
		return ports.TaxonomyNodeFilter{}, domain.ErrInvalidTaxonomyLevel
	}
	filter.Level = &level
	return filter, nil
}

func productsPortfolioFilterFromRequest(r *http.Request) (catalogreadmodel.ProductsPortfolioFilter, error) {
	filter := catalogreadmodel.ProductsPortfolioFilter{
		Search:             r.URL.Query().Get("search"),
		BrandNames:         parseMultiQuery(r, "brand_name"),
		TaxonomyLeaf0Names: parseMultiQuery(r, "taxonomy_leaf0_name"),
		Statuses:           parseMultiQuery(r, "status"),
		SortKey:            r.URL.Query().Get("sort_key"),
		SortDirection:      r.URL.Query().Get("sort_direction"),
	}

	limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitRaw != "" {
		limit, err := strconv.Atoi(limitRaw)
		if err != nil || limit < 1 {
			return catalogreadmodel.ProductsPortfolioFilter{}, errors.New("products portfolio limit must be a positive integer")
		}
		filter.Limit = limit
	}

	offsetRaw := strings.TrimSpace(r.URL.Query().Get("offset"))
	if offsetRaw != "" {
		offset, err := strconv.Atoi(offsetRaw)
		if err != nil || offset < 0 {
			return catalogreadmodel.ProductsPortfolioFilter{}, errors.New("products portfolio offset must be zero or greater")
		}
		filter.Offset = offset
	}
	if sortKey := strings.TrimSpace(strings.ToLower(filter.SortKey)); sortKey != "" {
		switch sortKey {
		case "pn_interno", "name", "brand_name", "taxonomy_leaf0_name", "product_status", "current_price_amount", "replacement_cost_amount", "on_hand_quantity":
		default:
			return catalogreadmodel.ProductsPortfolioFilter{}, errors.New("products portfolio sort_key is invalid")
		}
	}
	if sortDirection := strings.TrimSpace(strings.ToLower(filter.SortDirection)); sortDirection != "" {
		switch sortDirection {
		case "asc", "desc":
		default:
			return catalogreadmodel.ProductsPortfolioFilter{}, errors.New("products portfolio sort_direction must be asc or desc")
		}
	}

	return filter, nil
}

func parseMultiQuery(r *http.Request, key string) []string {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		raw := strings.TrimSpace(r.URL.Query().Get(key))
		if raw == "" {
			return nil
		}
		values = []string{raw}
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
	}
	return out
}

func mapProduct(product domain.Product) map[string]any {
	return map[string]any{
		"product_id":               product.ProductID,
		"tenant_id":                product.TenantID,
		"sku":                      product.SKU,
		"name":                     product.Name,
		"description":              product.Description,
		"brand_name":               product.BrandName,
		"stock_profile_code":       product.StockProfileCode,
		"primary_taxonomy_node_id": product.PrimaryTaxonomyNodeID,
		"identifiers":              mapProductIdentifiers(product.Identifiers),
		"status":                   string(product.Status),
	}
}

func mapProductIdentifier(identifier domain.ProductIdentifier) map[string]any {
	return map[string]any{
		"product_identifier_id": identifier.ProductIdentifierID,
		"product_id":            identifier.ProductID,
		"tenant_id":             identifier.TenantID,
		"identifier_type":       identifier.IdentifierType,
		"identifier_value":      identifier.IdentifierValue,
		"source_system":         identifier.SourceSystem,
		"is_primary":            identifier.IsPrimary,
	}
}

func mapProductIdentifiers(identifiers []domain.ProductIdentifier) []map[string]any {
	items := make([]map[string]any, 0, len(identifiers))
	for _, identifier := range identifiers {
		items = append(items, mapProductIdentifier(identifier))
	}
	return items
}

func mapCreateProductIdentifierInputs(inputs []CreateProductIdentifierRequest) []application.CreateProductIdentifierInput {
	identifiers := make([]application.CreateProductIdentifierInput, 0, len(inputs))
	for _, input := range inputs {
		identifiers = append(identifiers, application.CreateProductIdentifierInput{
			IdentifierType:  input.IdentifierType,
			IdentifierValue: input.IdentifierValue,
			SourceSystem:    input.SourceSystem,
			IsPrimary:       input.IsPrimary,
		})
	}
	return identifiers
}

func mapTaxonomyNode(node domain.TaxonomyNode) map[string]any {
	return map[string]any{
		"taxonomy_node_id":        node.TaxonomyNodeID,
		"tenant_id":               node.TenantID,
		"name":                    node.Name,
		"name_norm":               node.NameNorm,
		"code":                    node.Code,
		"parent_taxonomy_node_id": node.ParentTaxonomyNodeID,
		"level":                   node.Level,
		"path":                    node.Path,
		"is_active":               node.IsActive,
	}
}

func mapTaxonomyLevelDef(def domain.TaxonomyLevelDef) map[string]any {
	return map[string]any{
		"tenant_id":   def.TenantID,
		"level":       def.Level,
		"label":       def.Label,
		"short_label": def.ShortLabel,
		"is_enabled":  def.IsEnabled,
	}
}

func mapProductsPortfolio(result catalogreadmodel.ProductsPortfolioResult) map[string]any {
	rows := make([]map[string]any, 0, len(result.Rows))
	for _, item := range result.Rows {
		rows = append(rows, mapProductsPortfolioItem(item))
	}

	return map[string]any{
		"rows": rows,
		"filters": map[string]any{
			"brands":               result.Filters.Brands,
			"taxonomy_leaf0_names": result.Filters.TaxonomyLeaf0Names,
			"taxonomy_leaf0_label": result.Filters.TaxonomyLeaf0Label,
			"status":               result.Filters.Status,
		},
		"paging": map[string]any{
			"offset":   result.Paging.Offset,
			"limit":    result.Paging.Limit,
			"returned": result.Paging.Returned,
			"total":    result.Paging.Total,
		},
	}
}

func mapProductsPortfolioItem(item catalogreadmodel.ProductsPortfolioItem) map[string]any {
	return map[string]any{
		"product_id":                item.ProductID,
		"sku":                       item.SKU,
		"name":                      item.Name,
		"description":               nullableStringValue(item.Description),
		"brand_name":                nullableStringValue(item.BrandName),
		"pn_interno":                nullableStringValue(item.PNInterno),
		"reference":                 nullableStringValue(item.Reference),
		"ean":                       nullableStringValue(item.EAN),
		"taxonomy_leaf_name":        nullableStringValue(item.TaxonomyLeafName),
		"taxonomy_leaf0_name":       nullableStringValue(item.TaxonomyLeaf0Name),
		"stock_profile_code":        nullableStringValue(item.StockProfileCode),
		"product_status":            item.ProductStatus,
		"current_price_amount":      nullableFloat64Value(item.CurrentPriceAmount),
		"replacement_cost_amount":   nullableFloat64Value(item.ReplacementCostAmount),
		"average_cost_amount":       nullableFloat64Value(item.AverageCostAmount),
		"currency_code":             nullableStringValue(item.CurrencyCode),
		"on_hand_quantity":          nullableFloat64Value(item.OnHandQuantity),
		"inventory_position_status": nullableStringValue(item.InventoryPositionStatus),
		"updated_at":                item.UpdatedAt.UTC(),
	}
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableFloat64Value(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
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

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, apiErrorEnvelope{
		Error: apiError{
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
