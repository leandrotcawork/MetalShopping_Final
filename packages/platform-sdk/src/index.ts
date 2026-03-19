import type {
  AuthSessionLogoutResponseV1,
  AuthSessionStateV1,
  CommonErrorV1,
  ProductsPortfolioListV1,
} from "@metalshopping/sdk-types";
import {
  AuthSessionApiClient,
} from "@metalshopping/sdk-types";
import { AuthSessionConfiguration } from "@metalshopping/sdk-types";
import {
  ProductsApiClient,
  type GeneratedProductsPortfolioSortDirection,
  type GeneratedProductsPortfolioSortKey,
} from "@metalshopping/sdk-types";
import { ProductsConfiguration } from "@metalshopping/sdk-types";

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

export type ServerCoreSdk = {
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

function parseAuthSessionState(raw: unknown): AuthSessionStateV1 {
  if (!isRecord(raw)) {
    throw new Error("[sdk-runtime] AuthSessionStateV1 response must be an object");
  }
  assertString(raw.user_id, "AuthSessionStateV1.user_id");
  assertString(raw.tenant_id, "AuthSessionStateV1.tenant_id");
  assertString(raw.display_name, "AuthSessionStateV1.display_name");
  assertString(raw.email, "AuthSessionStateV1.email");
  assertString(raw.issued_at, "AuthSessionStateV1.issued_at");
  assertString(raw.expires_at, "AuthSessionStateV1.expires_at");
  assertString(raw.idle_timeout_expires_at, "AuthSessionStateV1.idle_timeout_expires_at");
  assertString(raw.absolute_timeout_expires_at, "AuthSessionStateV1.absolute_timeout_expires_at");
  assertStringArray(raw.roles, "AuthSessionStateV1.roles");
  assertStringArray(raw.capabilities, "AuthSessionStateV1.capabilities");
  return raw as AuthSessionStateV1;
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

  return {
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
