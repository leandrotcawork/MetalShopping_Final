import type {
  AuthSessionLogoutResponseV1,
  AuthSessionStateV1,
  CommonErrorV1,
  HomeSummaryV1,
  ProductsPortfolioListV1,
  ShoppingBootstrapV1,
  ShoppingCreateRunRequestV1,
  ShoppingCreateRunResponseV1,
  ShoppingManualUrlCandidateListV1,
  ShoppingManualUrlCandidateV1,
  ShoppingRunRequestV1,
  ShoppingProductLatestV1,
  ShoppingRunListV1,
  ShoppingSupplierSignalListV1,
  ShoppingSupplierSignalV1,
  ShoppingRunV1,
  ShoppingSummaryV1,
  ShoppingUpsertSupplierSignalRequestV1,
} from "@metalshopping/sdk-types";
import {
  AuthSessionApiClient,
  HomeApiClient,
  ShoppingApiClient,
} from "@metalshopping/sdk-types";
import { AuthSessionConfiguration } from "@metalshopping/sdk-types";
import { HomeConfiguration } from "@metalshopping/sdk-types";
import {
  ProductsApiClient,
  type GeneratedProductsPortfolioSortDirection,
  type GeneratedProductsPortfolioSortKey,
} from "@metalshopping/sdk-types";
import { ProductsConfiguration } from "@metalshopping/sdk-types";
import { ShoppingConfiguration } from "@metalshopping/sdk-types";

export type QueryParamValue = string | number | boolean | null | undefined;

export const defaultWebSessionCSRFCookieName = "ms_web_csrf";
export const defaultWebSessionCSRFHeaderName = "X-CSRF-Token";

export type GeneratedHttpClient = {
  baseUrl: string;
  defaultHeaders?: Record<string, string>;
  bearerToken?: string;
  csrfCookieName: string;
  csrfHeaderName: string;
};

export type BrowserGeneratedHttpClientConfig = {
  baseUrl: string;
  bearerToken?: string;
  csrfCookieName?: string;
  csrfHeaderName?: string;
  defaultHeaders?: Record<string, string>;
};

type HttpClientError = Error & {
  status?: number;
  code?: string;
  traceId?: string;
};

export type StartAuthSessionLoginQueryParams = {
  return_to?: string;
};

export type ProductsPortfolioSortKey = GeneratedProductsPortfolioSortKey;

export type ProductsPortfolioSortDirection = GeneratedProductsPortfolioSortDirection;

export type ListProductsPortfolioQueryParams = {
  search?: string;
  brand_name?: string;
  taxonomy_leaf0_name?: string;
  status?: string;
  sort_key?: ProductsPortfolioSortKey;
  sort_direction?: ProductsPortfolioSortDirection;
  limit?: number;
  offset?: number;
};

export type ShoppingRunStatus = "queued" | "running" | "completed" | "failed";

export type ListShoppingRunsQueryParams = {
  status?: ShoppingRunStatus;
  limit?: number;
  offset?: number;
};

export type ListShoppingSupplierSignalsQueryParams = {
  supplierCode?: string;
  productId?: string;
  limit?: number;
  offset?: number;
};

export type ListShoppingManualUrlCandidatesQueryParams = {
  supplierCode: string;
  search?: string;
  brandName?: string;
  taxonomyLeaf0Name?: string;
  includeExisting?: boolean;
  limit?: number;
  offset?: number;
};

export type ServerCoreSdk = {
  home: {
    getSummary(): Promise<HomeSummaryV1>;
  };
  shopping: {
    getBootstrap(): Promise<ShoppingBootstrapV1>;
    createRunRequest(payload: ShoppingCreateRunRequestV1): Promise<ShoppingCreateRunResponseV1>;
    getRunRequest(runRequestId: string): Promise<ShoppingRunRequestV1>;
    getSummary(): Promise<ShoppingSummaryV1>;
    listRuns(query?: ListShoppingRunsQueryParams): Promise<ShoppingRunListV1>;
    getRun(runId: string): Promise<ShoppingRunV1>;
    getProductLatest(productId: string): Promise<ShoppingProductLatestV1>;
    listSupplierSignals(query?: ListShoppingSupplierSignalsQueryParams): Promise<ShoppingSupplierSignalListV1>;
    listManualUrlCandidates(
      query: ListShoppingManualUrlCandidatesQueryParams,
    ): Promise<ShoppingManualUrlCandidateListV1>;
    upsertSupplierSignal(payload: ShoppingUpsertSupplierSignalRequestV1): Promise<ShoppingSupplierSignalV1>;
  };
  authSession: {
    buildStartLoginUrl(query?: StartAuthSessionLoginQueryParams): Promise<string>;
    getSessionState(): Promise<AuthSessionStateV1>;
    refreshSession(): Promise<AuthSessionStateV1>;
    logoutSession(): Promise<AuthSessionLogoutResponseV1>;
  };
  products: {
    listProductsPortfolio(query?: ListProductsPortfolioQueryParams): Promise<ProductsPortfolioListV1>;
  };
};

function createTraceId() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `trace-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`;
}

function readCookie(name: string) {
  if (typeof document === "undefined") {
    return null;
  }
  const prefix = `${encodeURIComponent(name)}=`;
  const entries = document.cookie.split(/;\s*/);
  for (const entry of entries) {
    if (!entry.startsWith(prefix)) {
      continue;
    }
    return decodeURIComponent(entry.slice(prefix.length));
  }
  return null;
}

function buildQueryString(query?: Record<string, QueryParamValue>) {
  const params = new URLSearchParams();
  if (query === undefined) {
    return "";
  }

  for (const [key, rawValue] of Object.entries(query)) {
    if (rawValue === undefined || rawValue === null || rawValue === "") {
      continue;
    }
    params.set(key, String(rawValue));
  }

  return params.toString();
}

function buildRequestUrl(baseUrl: string, path: string, query?: Record<string, QueryParamValue>) {
  const queryString = buildQueryString(query);
  return `${baseUrl.replace(/\/$/, "")}${path}${queryString ? `?${queryString}` : ""}`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function assertString(value: unknown, fieldName: string): asserts value is string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`[sdk-runtime] ${fieldName} must be a non-empty string`);
  }
}

function assertStringArray(value: unknown, fieldName: string): asserts value is string[] {
  if (!Array.isArray(value) || value.some((item) => typeof item !== "string")) {
    throw new Error(`[sdk-runtime] ${fieldName} must be a string array`);
  }
}

function assertNumber(value: unknown, fieldName: string): asserts value is number {
  if (typeof value !== "number" || Number.isNaN(value)) {
    throw new Error(`[sdk-runtime] ${fieldName} must be a number`);
  }
}

function normalizeDateTime(value: unknown, fieldName: string): string {
  if (value instanceof Date) {
    if (Number.isNaN(value.getTime())) {
      throw new Error(`[sdk-runtime] ${fieldName} must be a valid date`);
    }
    return value.toISOString();
  }
  if (typeof value === "string" && value.trim() !== "") {
    return value;
  }
  throw new Error(`[sdk-runtime] ${fieldName} must be a non-empty string`);
}

function parseAuthSessionState(raw: unknown): AuthSessionStateV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] AuthSessionStateV1 response must be an object");
  }
  assertString(raw.user_id, "AuthSessionStateV1.user_id");
  assertString(raw.tenant_id, "AuthSessionStateV1.tenant_id");
  assertString(raw.display_name, "AuthSessionStateV1.display_name");
  assertString(raw.email, "AuthSessionStateV1.email");
  assertStringArray(raw.roles, "AuthSessionStateV1.roles");
  assertStringArray(raw.capabilities, "AuthSessionStateV1.capabilities");

  const issuedAt = normalizeDateTime(raw.issued_at, "AuthSessionStateV1.issued_at");
  const expiresAt = normalizeDateTime(raw.expires_at, "AuthSessionStateV1.expires_at");
  const idleTimeoutExpiresAt = normalizeDateTime(
    raw.idle_timeout_expires_at,
    "AuthSessionStateV1.idle_timeout_expires_at",
  );
  const absoluteTimeoutExpiresAt = normalizeDateTime(
    raw.absolute_timeout_expires_at,
    "AuthSessionStateV1.absolute_timeout_expires_at",
  );

  if (raw.session_id !== undefined && raw.session_id !== null) {
    assertString(raw.session_id, "AuthSessionStateV1.session_id");
  }

  return {
    user_id: raw.user_id,
    tenant_id: raw.tenant_id,
    display_name: raw.display_name,
    email: raw.email,
    roles: raw.roles,
    capabilities: raw.capabilities,
    issued_at: issuedAt,
    expires_at: expiresAt,
    idle_timeout_expires_at: idleTimeoutExpiresAt,
    absolute_timeout_expires_at: absoluteTimeoutExpiresAt,
    session_id: raw.session_id ?? undefined,
  };
}

function parseLogoutResponse(raw: unknown): AuthSessionLogoutResponseV1 {
  if (!isRecord(raw) || typeof raw.logged_out !== "boolean") {
    throw new Error("[sdk-runtime] AuthSessionLogoutResponseV1 response must include boolean logged_out");
  }
  return raw as AuthSessionLogoutResponseV1;
}

function parseProductsPortfolioList(raw: unknown): ProductsPortfolioListV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ProductsPortfolioListV1 response must be an object");
  }
  if (!Array.isArray(raw.rows)) {
    throw new Error("[sdk-runtime] ProductsPortfolioListV1.rows must be an array");
  }
  if (!isRecord(raw.filters)) {
    throw new Error("[sdk-runtime] ProductsPortfolioListV1.filters must be an object");
  }
  if (!isRecord(raw.paging)) {
    throw new Error("[sdk-runtime] ProductsPortfolioListV1.paging must be an object");
  }
  const paging = raw.paging;
  assertNumber(paging.offset, "ProductsPortfolioListV1.paging.offset");
  assertNumber(paging.limit, "ProductsPortfolioListV1.paging.limit");
  assertNumber(paging.returned, "ProductsPortfolioListV1.paging.returned");
  assertNumber(paging.total, "ProductsPortfolioListV1.paging.total");
  return raw as ProductsPortfolioListV1;
}

function parseHomeSummary(raw: unknown): HomeSummaryV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] HomeSummaryV1 response must be an object");
  }
  assertNumber(raw.productCount, "HomeSummaryV1.productCount");
  assertNumber(raw.activeProductCount, "HomeSummaryV1.activeProductCount");
  assertNumber(raw.pricedProductCount, "HomeSummaryV1.pricedProductCount");
  assertNumber(raw.inventoryTrackedCount, "HomeSummaryV1.inventoryTrackedCount");
  const lastUpdated = normalizeDateTime(raw.lastUpdated, "HomeSummaryV1.lastUpdated");
  return {
    productCount: raw.productCount,
    activeProductCount: raw.activeProductCount,
    pricedProductCount: raw.pricedProductCount,
    inventoryTrackedCount: raw.inventoryTrackedCount,
    lastUpdated,
  };
}

function parseShoppingRun(raw: unknown): ShoppingRunV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingRunV1 response must be an object");
  }
  assertString(raw.runId, "ShoppingRunV1.runId");
  assertString(raw.status, "ShoppingRunV1.status");
  const startedAt = normalizeDateTime(raw.startedAt, "ShoppingRunV1.startedAt");
  let finishedAt: string | null | undefined;
  if (raw.finishedAt !== undefined) {
    if (raw.finishedAt === null) {
      finishedAt = null;
    } else {
      finishedAt = normalizeDateTime(raw.finishedAt, "ShoppingRunV1.finishedAt");
    }
  }
  assertNumber(raw.processedItems, "ShoppingRunV1.processedItems");
  assertNumber(raw.totalItems, "ShoppingRunV1.totalItems");

  const notes = raw.notes;
  if (notes !== undefined && notes !== null && typeof notes !== "string") {
    throw new Error("[sdk-runtime] ShoppingRunV1.notes must be a string when provided");
  }

  return {
    runId: raw.runId,
    status: raw.status as ShoppingRunV1["status"],
    startedAt,
    finishedAt,
    processedItems: raw.processedItems,
    totalItems: raw.totalItems,
    notes: notes ?? undefined,
  };
}

function parseShoppingRunList(raw: unknown): ShoppingRunListV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingRunListV1 response must be an object");
  }
  if (!Array.isArray(raw.rows)) {
    throw new Error("[sdk-runtime] ShoppingRunListV1.rows must be an array");
  }
  if (!isRecord(raw.paging)) {
    throw new Error("[sdk-runtime] ShoppingRunListV1.paging must be an object");
  }
  const paging = raw.paging;
  assertNumber(paging.offset, "ShoppingRunListV1.paging.offset");
  assertNumber(paging.limit, "ShoppingRunListV1.paging.limit");
  assertNumber(paging.returned, "ShoppingRunListV1.paging.returned");
  assertNumber(paging.total, "ShoppingRunListV1.paging.total");

  return {
    rows: raw.rows.map(parseShoppingRun),
    paging: {
      offset: paging.offset,
      limit: paging.limit,
      returned: paging.returned,
      total: paging.total,
    },
  };
}

function parseShoppingSummary(raw: unknown): ShoppingSummaryV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingSummaryV1 response must be an object");
  }
  assertNumber(raw.totalRuns, "ShoppingSummaryV1.totalRuns");
  assertNumber(raw.runningRuns, "ShoppingSummaryV1.runningRuns");
  assertNumber(raw.completedRuns, "ShoppingSummaryV1.completedRuns");
  assertNumber(raw.failedRuns, "ShoppingSummaryV1.failedRuns");
  const lastRunAt = normalizeDateTime(raw.lastRunAt, "ShoppingSummaryV1.lastRunAt");

  return {
    totalRuns: raw.totalRuns,
    runningRuns: raw.runningRuns,
    completedRuns: raw.completedRuns,
    failedRuns: raw.failedRuns,
    lastRunAt,
  };
}

function parseShoppingBootstrap(raw: unknown): ShoppingBootstrapV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1 response must be an object");
  }
  if (!Array.isArray(raw.inputModes) || raw.inputModes.some((item) => item !== "xlsx" && item !== "catalog")) {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1.inputModes must be an array with xlsx/catalog");
  }
  if (
    !Array.isArray(raw.runStatuses) ||
    raw.runStatuses.some((item) => !["queued", "running", "completed", "failed"].includes(String(item)))
  ) {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1.runStatuses must be a valid status array");
  }
  if (typeof raw.supportsManualUrls !== "boolean") {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1.supportsManualUrls must be boolean");
  }
  if (!isRecord(raw.advancedDefaults)) {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1.advancedDefaults must be an object");
  }
  const defaults = raw.advancedDefaults;
  assertNumber(defaults.timeoutSeconds, "ShoppingBootstrapV1.advancedDefaults.timeoutSeconds");
  assertNumber(defaults.httpWorkers, "ShoppingBootstrapV1.advancedDefaults.httpWorkers");
  assertNumber(defaults.playwrightWorkers, "ShoppingBootstrapV1.advancedDefaults.playwrightWorkers");
  assertNumber(defaults.topN, "ShoppingBootstrapV1.advancedDefaults.topN");
  if (!Array.isArray(raw.suppliers)) {
    throw new Error("[sdk-runtime] ShoppingBootstrapV1.suppliers must be an array");
  }
  return raw as ShoppingBootstrapV1;
}

function parseShoppingCreateRunResponse(raw: unknown): ShoppingCreateRunResponseV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingCreateRunResponseV1 response must be an object");
  }
  assertString(raw.runRequestId, "ShoppingCreateRunResponseV1.runRequestId");
  assertString(raw.status, "ShoppingCreateRunResponseV1.status");
  assertString(raw.inputMode, "ShoppingCreateRunResponseV1.inputMode");
  const requestedAt = normalizeDateTime(raw.requestedAt, "ShoppingCreateRunResponseV1.requestedAt");
  assertString(raw.requestedBy, "ShoppingCreateRunResponseV1.requestedBy");
  return {
    runRequestId: raw.runRequestId,
    status: raw.status as ShoppingCreateRunResponseV1["status"],
    inputMode: raw.inputMode as ShoppingCreateRunResponseV1["inputMode"],
    requestedAt,
    requestedBy: raw.requestedBy,
  };
}

function parseShoppingRunRequest(raw: unknown): ShoppingRunRequestV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingRunRequestV1 response must be an object");
  }
  assertString(raw.runRequestId, "ShoppingRunRequestV1.runRequestId");
  assertString(raw.status, "ShoppingRunRequestV1.status");
  assertString(raw.inputMode, "ShoppingRunRequestV1.inputMode");
  const requestedAt = normalizeDateTime(raw.requestedAt, "ShoppingRunRequestV1.requestedAt");
  assertString(raw.requestedBy, "ShoppingRunRequestV1.requestedBy");
  const claimedAt = raw.claimedAt === null || raw.claimedAt === undefined ? null : normalizeDateTime(raw.claimedAt, "ShoppingRunRequestV1.claimedAt");
  const startedAt = raw.startedAt === null || raw.startedAt === undefined ? null : normalizeDateTime(raw.startedAt, "ShoppingRunRequestV1.startedAt");
  const finishedAt = raw.finishedAt === null || raw.finishedAt === undefined ? null : normalizeDateTime(raw.finishedAt, "ShoppingRunRequestV1.finishedAt");

  if (raw.workerId !== undefined && raw.workerId !== null) {
    assertString(raw.workerId, "ShoppingRunRequestV1.workerId");
  }
  if (raw.runId !== undefined && raw.runId !== null) {
    assertString(raw.runId, "ShoppingRunRequestV1.runId");
  }
  if (raw.errorMessage !== undefined && raw.errorMessage !== null) {
    assertString(raw.errorMessage, "ShoppingRunRequestV1.errorMessage");
  }
  if (raw.catalogProductIds !== undefined) {
    assertStringArray(raw.catalogProductIds, "ShoppingRunRequestV1.catalogProductIds");
  }
  if (raw.xlsxScopeIdentifiers !== undefined) {
    assertStringArray(raw.xlsxScopeIdentifiers, "ShoppingRunRequestV1.xlsxScopeIdentifiers");
  }
  if (raw.resolvedCatalogProductIds !== undefined) {
    assertStringArray(raw.resolvedCatalogProductIds, "ShoppingRunRequestV1.resolvedCatalogProductIds");
  }
  if (raw.unresolvedScopeIdentifiers !== undefined) {
    assertStringArray(raw.unresolvedScopeIdentifiers, "ShoppingRunRequestV1.unresolvedScopeIdentifiers");
  }
  if (raw.ambiguousScopeIdentifiers !== undefined) {
    assertStringArray(raw.ambiguousScopeIdentifiers, "ShoppingRunRequestV1.ambiguousScopeIdentifiers");
  }
  if (raw.xlsxFilePath !== undefined && raw.xlsxFilePath !== null) {
    assertString(raw.xlsxFilePath, "ShoppingRunRequestV1.xlsxFilePath");
  }

  return {
    runRequestId: raw.runRequestId,
    status: raw.status as ShoppingRunRequestV1["status"],
    inputMode: raw.inputMode as ShoppingRunRequestV1["inputMode"],
    requestedAt,
    requestedBy: raw.requestedBy,
    claimedAt: claimedAt === null ? null : claimedAt,
    startedAt: startedAt === null ? null : startedAt,
    finishedAt: finishedAt === null ? null : finishedAt,
    workerId: raw.workerId ?? null,
    runId: raw.runId ?? null,
    errorMessage: raw.errorMessage ?? null,
    catalogProductIds: (raw.catalogProductIds as string[] | undefined) ?? [],
    xlsxFilePath: (raw.xlsxFilePath as string | null | undefined) ?? null,
    xlsxScopeIdentifiers: (raw.xlsxScopeIdentifiers as string[] | undefined) ?? [],
    resolvedCatalogProductIds: (raw.resolvedCatalogProductIds as string[] | undefined) ?? [],
    unresolvedScopeIdentifiers: (raw.unresolvedScopeIdentifiers as string[] | undefined) ?? [],
    ambiguousScopeIdentifiers: (raw.ambiguousScopeIdentifiers as string[] | undefined) ?? [],
  };
}

function parseShoppingProductLatest(raw: unknown): ShoppingProductLatestV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingProductLatestV1 response must be an object");
  }
  assertString(raw.productId, "ShoppingProductLatestV1.productId");
  assertString(raw.runId, "ShoppingProductLatestV1.runId");
  const observedAt = normalizeDateTime(raw.observedAt, "ShoppingProductLatestV1.observedAt");
  assertString(raw.sellerName, "ShoppingProductLatestV1.sellerName");
  assertString(raw.channel, "ShoppingProductLatestV1.channel");
  assertNumber(raw.observedPrice, "ShoppingProductLatestV1.observedPrice");
  assertString(raw.currency, "ShoppingProductLatestV1.currency");

  return {
    productId: raw.productId,
    runId: raw.runId,
    observedAt,
    sellerName: raw.sellerName,
    channel: raw.channel,
    observedPrice: raw.observedPrice,
    currency: raw.currency,
  };
}

function parseShoppingSupplierSignal(raw: unknown): ShoppingSupplierSignalV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingSupplierSignalV1 response must be an object");
  }
  assertString(raw.productId, "ShoppingSupplierSignalV1.productId");
  assertString(raw.supplierCode, "ShoppingSupplierSignalV1.supplierCode");
  assertString(raw.urlStatus, "ShoppingSupplierSignalV1.urlStatus");
  assertString(raw.lookupMode, "ShoppingSupplierSignalV1.lookupMode");
  assertString(raw.lookupModeSource, "ShoppingSupplierSignalV1.lookupModeSource");
  if (typeof raw.manualOverride !== "boolean") {
    throw new Error("[sdk-runtime] ShoppingSupplierSignalV1.manualOverride must be boolean");
  }
  const updatedAt = normalizeDateTime(raw.updatedAt, "ShoppingSupplierSignalV1.updatedAt");
  if (raw.productUrl !== undefined && raw.productUrl !== null) {
    assertString(raw.productUrl, "ShoppingSupplierSignalV1.productUrl");
  }
  const productUrl = (raw.productUrl as string | null | undefined) ?? null;
  const lastCheckedAt = raw.lastCheckedAt === null || raw.lastCheckedAt === undefined ? null : normalizeDateTime(raw.lastCheckedAt, "ShoppingSupplierSignalV1.lastCheckedAt");
  const lastSuccessAt = raw.lastSuccessAt === null || raw.lastSuccessAt === undefined ? null : normalizeDateTime(raw.lastSuccessAt, "ShoppingSupplierSignalV1.lastSuccessAt");
  if (raw.lastHttpStatus !== undefined && raw.lastHttpStatus !== null) {
    assertNumber(raw.lastHttpStatus, "ShoppingSupplierSignalV1.lastHttpStatus");
  }
  if (raw.lastErrorMessage !== undefined && raw.lastErrorMessage !== null) {
    assertString(raw.lastErrorMessage, "ShoppingSupplierSignalV1.lastErrorMessage");
  }

  return {
    productId: raw.productId,
    supplierCode: raw.supplierCode,
    productUrl: productUrl === null ? null : productUrl,
    urlStatus: raw.urlStatus as ShoppingSupplierSignalV1["urlStatus"],
    lookupMode: raw.lookupMode as ShoppingSupplierSignalV1["lookupMode"],
    lookupModeSource: raw.lookupModeSource as ShoppingSupplierSignalV1["lookupModeSource"],
    manualOverride: raw.manualOverride,
    lastCheckedAt: lastCheckedAt === null ? null : lastCheckedAt,
    lastSuccessAt: lastSuccessAt === null ? null : lastSuccessAt,
    lastHttpStatus: (raw.lastHttpStatus as number | null | undefined) ?? null,
    lastErrorMessage: (raw.lastErrorMessage as string | null | undefined) ?? null,
    updatedAt,
  };
}

function parseShoppingSupplierSignalList(raw: unknown): ShoppingSupplierSignalListV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingSupplierSignalListV1 response must be an object");
  }
  if (!Array.isArray(raw.rows)) {
    throw new Error("[sdk-runtime] ShoppingSupplierSignalListV1.rows must be an array");
  }
  const rows = raw.rows.map((item) => parseShoppingSupplierSignal(item));
  if (!isRecord(raw.paging)) {
    throw new Error("[sdk-runtime] ShoppingSupplierSignalListV1.paging must be an object");
  }
  assertNumber(raw.paging.offset, "ShoppingSupplierSignalListV1.paging.offset");
  assertNumber(raw.paging.limit, "ShoppingSupplierSignalListV1.paging.limit");
  assertNumber(raw.paging.returned, "ShoppingSupplierSignalListV1.paging.returned");
  assertNumber(raw.paging.total, "ShoppingSupplierSignalListV1.paging.total");

  return {
    rows,
    paging: {
      offset: raw.paging.offset,
      limit: raw.paging.limit,
      returned: raw.paging.returned,
      total: raw.paging.total,
    },
  };
}

function parseShoppingManualUrlCandidate(raw: unknown): ShoppingManualUrlCandidateV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingManualUrlCandidateV1 response must be an object");
  }
  assertString(raw.productId, "ShoppingManualUrlCandidateV1.productId");
  assertString(raw.supplierCode, "ShoppingManualUrlCandidateV1.supplierCode");
  assertString(raw.sku, "ShoppingManualUrlCandidateV1.sku");
  assertString(raw.name, "ShoppingManualUrlCandidateV1.name");
  assertString(raw.urlStatus, "ShoppingManualUrlCandidateV1.urlStatus");
  assertString(raw.lookupMode, "ShoppingManualUrlCandidateV1.lookupMode");
  assertString(raw.lookupModeSource, "ShoppingManualUrlCandidateV1.lookupModeSource");
  if (typeof raw.manualOverride !== "boolean") {
    throw new Error("[sdk-runtime] ShoppingManualUrlCandidateV1.manualOverride must be boolean");
  }
  assertNumber(raw.notFoundCount, "ShoppingManualUrlCandidateV1.notFoundCount");
  const updatedAt = normalizeDateTime(raw.updatedAt, "ShoppingManualUrlCandidateV1.updatedAt");

  if (raw.productUrl !== undefined && raw.productUrl !== null) {
    assertString(raw.productUrl, "ShoppingManualUrlCandidateV1.productUrl");
  }
  if (raw.brandName !== undefined && raw.brandName !== null) {
    assertString(raw.brandName, "ShoppingManualUrlCandidateV1.brandName");
  }
  if (raw.taxonomyLeaf0Name !== undefined && raw.taxonomyLeaf0Name !== null) {
    assertString(raw.taxonomyLeaf0Name, "ShoppingManualUrlCandidateV1.taxonomyLeaf0Name");
  }

  const lastCheckedAt =
    raw.lastCheckedAt === null || raw.lastCheckedAt === undefined
      ? null
      : normalizeDateTime(raw.lastCheckedAt, "ShoppingManualUrlCandidateV1.lastCheckedAt");
  const lastSuccessAt =
    raw.lastSuccessAt === null || raw.lastSuccessAt === undefined
      ? null
      : normalizeDateTime(raw.lastSuccessAt, "ShoppingManualUrlCandidateV1.lastSuccessAt");
  if (raw.lastHttpStatus !== undefined && raw.lastHttpStatus !== null) {
    assertNumber(raw.lastHttpStatus, "ShoppingManualUrlCandidateV1.lastHttpStatus");
  }
  if (raw.lastErrorMessage !== undefined && raw.lastErrorMessage !== null) {
    assertString(raw.lastErrorMessage, "ShoppingManualUrlCandidateV1.lastErrorMessage");
  }
  const nextDiscoveryAt =
    raw.nextDiscoveryAt === null || raw.nextDiscoveryAt === undefined
      ? null
      : normalizeDateTime(raw.nextDiscoveryAt, "ShoppingManualUrlCandidateV1.nextDiscoveryAt");

  return {
    productId: raw.productId,
    supplierCode: raw.supplierCode,
    sku: raw.sku,
    name: raw.name,
    brandName: (raw.brandName as string | null | undefined) ?? null,
    taxonomyLeaf0Name: (raw.taxonomyLeaf0Name as string | null | undefined) ?? null,
    productUrl: (raw.productUrl as string | null | undefined) ?? null,
    urlStatus: raw.urlStatus as ShoppingManualUrlCandidateV1["urlStatus"],
    lookupMode: raw.lookupMode as ShoppingManualUrlCandidateV1["lookupMode"],
    lookupModeSource: raw.lookupModeSource as ShoppingManualUrlCandidateV1["lookupModeSource"],
    manualOverride: raw.manualOverride,
    lastCheckedAt: lastCheckedAt === null ? null : lastCheckedAt,
    lastSuccessAt: lastSuccessAt === null ? null : lastSuccessAt,
    lastHttpStatus: (raw.lastHttpStatus as number | null | undefined) ?? null,
    lastErrorMessage: (raw.lastErrorMessage as string | null | undefined) ?? null,
    nextDiscoveryAt: nextDiscoveryAt === null ? null : nextDiscoveryAt,
    notFoundCount: raw.notFoundCount,
    updatedAt,
  };
}

function parseShoppingManualUrlCandidateList(raw: unknown): ShoppingManualUrlCandidateListV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] ShoppingManualUrlCandidateListV1 response must be an object");
  }
  if (!Array.isArray(raw.rows)) {
    throw new Error("[sdk-runtime] ShoppingManualUrlCandidateListV1.rows must be an array");
  }
  const rows = raw.rows.map((item) => parseShoppingManualUrlCandidate(item));
  if (!isRecord(raw.paging)) {
    throw new Error("[sdk-runtime] ShoppingManualUrlCandidateListV1.paging must be an object");
  }
  assertNumber(raw.paging.offset, "ShoppingManualUrlCandidateListV1.paging.offset");
  assertNumber(raw.paging.limit, "ShoppingManualUrlCandidateListV1.paging.limit");
  assertNumber(raw.paging.returned, "ShoppingManualUrlCandidateListV1.paging.returned");
  assertNumber(raw.paging.total, "ShoppingManualUrlCandidateListV1.paging.total");

  return {
    rows,
    paging: {
      offset: raw.paging.offset,
      limit: raw.paging.limit,
      returned: raw.paging.returned,
      total: raw.paging.total,
    },
  };
}

async function parseCommonError(response: Response) {
  const raw = (await response.clone().json().catch(() => null)) as unknown;
  if (!isRecord(raw) || !isRecord(raw.error)) {
    return null;
  }
  const { error } = raw;
  if (typeof error.code !== "string" || typeof error.message !== "string") {
    return null;
  }
  return raw as CommonErrorV1;
}

function mapCommonError(response: Response, errorPayload: CommonErrorV1 | null) {
  const message = errorPayload?.error?.message ?? `Request failed with status ${response.status}`;
  const error = new Error(message) as HttpClientError;
  error.status = response.status;
  error.code = errorPayload?.error?.code;
  error.traceId = errorPayload?.error?.trace_id;
  return error;
}

function extractResponseFromError(error: unknown): Response | null {
  if (!isRecord(error) || !("response" in error)) {
    return null;
  }
  const response = (error as { response?: unknown }).response;
  if (response instanceof Response) {
    return response;
  }
  return null;
}

async function runGeneratedCall<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation();
  } catch (error) {
    const response = extractResponseFromError(error);
    if (response === null) {
      throw error;
    }
    const errorPayload = await parseCommonError(response);
    throw mapCommonError(response, errorPayload);
  }
}

function readCsrfTokenOrThrow(client: GeneratedHttpClient): string {
  const token = readCookie(client.csrfCookieName);
  if (token === null || token.trim() === "") {
    const error = new Error("CSRF token ausente para operacao autenticada.") as HttpClientError;
    error.status = 403;
    error.code = "AUTH_CSRF_TOKEN_MISSING";
    throw error;
  }
  return token;
}

type AuthSessionConfigurationParameters = NonNullable<ConstructorParameters<typeof AuthSessionConfiguration>[0]>;

function createBrowserMiddleware(client: GeneratedHttpClient): NonNullable<AuthSessionConfigurationParameters["middleware"]> {
  const defaultHeaders = client.defaultHeaders ?? {};
  return [{
    pre: async (context: { init: RequestInit; url: string }) => {
      const { init, url } = context;
      const headers = new Headers(init.headers ?? undefined);
      headers.set("Accept", headers.get("Accept") ?? "application/json");
      headers.set("X-Trace-Id", headers.get("X-Trace-Id") ?? createTraceId());

      if ((client.bearerToken ?? "").trim() !== "" && !headers.has("Authorization")) {
        headers.set("Authorization", `Bearer ${client.bearerToken}`);
      }

      const method = (init.method ?? "GET").toUpperCase();
      if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
        const csrfToken = readCookie(client.csrfCookieName);
        if (csrfToken && !headers.has(client.csrfHeaderName)) {
          headers.set(client.csrfHeaderName, csrfToken);
        }
      }

      for (const [name, value] of Object.entries(defaultHeaders)) {
        if (!headers.has(name)) {
          headers.set(name, value);
        }
      }

      return {
        url,
        init: {
          ...init,
          credentials: init.credentials ?? "include",
          headers,
        },
      };
    },
  }];
}

export function createBrowserGeneratedHttpClient(config: BrowserGeneratedHttpClientConfig): GeneratedHttpClient {
  return {
    baseUrl: config.baseUrl.replace(/\/$/, ""),
    defaultHeaders: config.defaultHeaders,
    bearerToken: config.bearerToken,
    csrfCookieName: config.csrfCookieName ?? defaultWebSessionCSRFCookieName,
    csrfHeaderName: config.csrfHeaderName ?? defaultWebSessionCSRFHeaderName,
  };
}

export function createServerCoreSdk(client: GeneratedHttpClient): ServerCoreSdk {
  const middleware = createBrowserMiddleware(client);
  const authApi = new AuthSessionApiClient(
    new AuthSessionConfiguration({
      basePath: client.baseUrl,
      middleware,
      headers: client.defaultHeaders,
      credentials: "include",
    }),
  );
  const productsApi = new ProductsApiClient(
    new ProductsConfiguration({
      basePath: client.baseUrl,
      middleware,
      headers: client.defaultHeaders,
      credentials: "include",
    }),
  );
  const homeApi = new HomeApiClient(
    new HomeConfiguration({
      basePath: client.baseUrl,
      middleware,
      headers: client.defaultHeaders,
      credentials: "include",
    }),
  );
  const shoppingApi = new ShoppingApiClient(
    new ShoppingConfiguration({
      basePath: client.baseUrl,
      middleware,
      headers: client.defaultHeaders,
      credentials: "include",
    }),
  );

  return {
    home: {
      async getSummary() {
        const raw = await runGeneratedCall(() => homeApi.getHomeSummary());
        return parseHomeSummary(raw);
      },
    },
    shopping: {
      async getBootstrap() {
        const raw = await runGeneratedCall(() => shoppingApi.getShoppingBootstrap());
        return parseShoppingBootstrap(raw);
      },
      async createRunRequest(payload) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.createShoppingRunRequest({
            shoppingCreateRunRequestV1: payload,
          }),
        );
        return parseShoppingCreateRunResponse(raw);
      },
      async getRunRequest(runRequestId) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.getShoppingRunRequest({
            runRequestId,
          }),
        );
        return parseShoppingRunRequest(raw);
      },
      async getSummary() {
        const raw = await runGeneratedCall(() => shoppingApi.getShoppingSummary());
        return parseShoppingSummary(raw);
      },
      async listRuns(query = {}) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.listShoppingRuns({
            status: query.status,
            limit: query.limit,
            offset: query.offset,
          }),
        );
        return parseShoppingRunList(raw);
      },
      async getRun(runId) {
        const raw = await runGeneratedCall(() => shoppingApi.getShoppingRun({ runId }));
        return parseShoppingRun(raw);
      },
      async getProductLatest(productId) {
        const raw = await runGeneratedCall(() => shoppingApi.getShoppingProductLatest({ productId }));
        return parseShoppingProductLatest(raw);
      },
      async listSupplierSignals(query = {}) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.listShoppingSupplierSignals({
            supplierCode: query.supplierCode,
            productId: query.productId,
            limit: query.limit,
            offset: query.offset,
          }),
        );
        return parseShoppingSupplierSignalList(raw);
      },
      async listManualUrlCandidates(query) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.listShoppingManualUrlCandidates({
            supplierCode: query.supplierCode,
            search: query.search,
            brandName: query.brandName,
            taxonomyLeaf0Name: query.taxonomyLeaf0Name,
            includeExisting: query.includeExisting,
            limit: query.limit,
            offset: query.offset,
          }),
        );
        return parseShoppingManualUrlCandidateList(raw);
      },
      async upsertSupplierSignal(payload) {
        const raw = await runGeneratedCall(() =>
          shoppingApi.upsertShoppingSupplierSignal({
            shoppingUpsertSupplierSignalRequestV1: payload,
          }),
        );
        return parseShoppingSupplierSignal(raw);
      },
    },
    authSession: {
      async buildStartLoginUrl(query) {
        const requestOptions = await authApi.startAuthSessionLoginRequestOpts({
          returnTo: query?.return_to,
        });
        return buildRequestUrl(
          client.baseUrl,
          requestOptions.path,
          requestOptions.query as Record<string, QueryParamValue> | undefined,
        );
      },
      async getSessionState() {
        const raw = await runGeneratedCall(() => authApi.getAuthSessionState());
        return parseAuthSessionState(raw);
      },
      async refreshSession() {
        const csrfToken = readCsrfTokenOrThrow(client);
        const raw = await runGeneratedCall(() =>
          authApi.refreshAuthSession({
            xCSRFToken: csrfToken,
          }),
        );
        return parseAuthSessionState(raw);
      },
      async logoutSession() {
        const csrfToken = readCsrfTokenOrThrow(client);
        const raw = await runGeneratedCall(() =>
          authApi.logoutAuthSession({
            xCSRFToken: csrfToken,
          }),
        );
        return parseLogoutResponse(raw);
      },
    },
    products: {
      async listProductsPortfolio(query = {}) {
        const raw = await runGeneratedCall(() =>
          productsApi.listProductsPortfolio({
            search: query.search,
            brandName: query.brand_name,
            taxonomyLeaf0Name: query.taxonomy_leaf0_name,
            status: query.status,
            sortKey: query.sort_key,
            sortDirection: query.sort_direction,
            limit: query.limit,
            offset: query.offset,
          }),
        );
        return parseProductsPortfolioList(raw);
      },
    },
  };
}
