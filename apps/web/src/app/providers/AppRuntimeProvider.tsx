import type { PropsWithChildren } from "react";
import { createContext, useContext, useMemo } from "react";

import type { GeneratedHttpClient } from "@metalshopping/generated-sdk";
import type { CommonErrorV1 } from "@metalshopping/generated-types";

type QueryParamValue = string | number | boolean | null | undefined;

export type AppHttpClient = GeneratedHttpClient;

type HttpClientError = Error & {
  status?: number;
  code?: string;
  traceId?: string;
};

type AppRuntime = {
  apiBaseUrl: string;
  bearerToken: string;
  httpClient: AppHttpClient;
};

const defaultApiBaseUrl = "http://127.0.0.1:8080";
const defaultBearerToken = "";

const AppRuntimeContext = createContext<AppRuntime | null>(null);

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

export function AppRuntimeProvider({ children }: PropsWithChildren) {
  const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl).replace(/\/$/, "");
  const bearerToken = import.meta.env.VITE_API_BEARER_TOKEN ?? defaultBearerToken;

  const runtime = useMemo<AppRuntime>(() => {
    function buildHeaders() {
      const headers: Record<string, string> = {
        Accept: "application/json",
      };
      if (bearerToken.trim() !== "") {
        headers.Authorization = `Bearer ${bearerToken}`;
      }
      return headers;
    }

    const httpClient: AppHttpClient = {
      async getJson<T>(path: string, options?: { query?: Record<string, QueryParamValue> }) {
        const queryString = buildQueryString(options?.query);
        const response = await fetch(`${apiBaseUrl}${path}${queryString ? `?${queryString}` : ""}`, {
          credentials: "include",
          headers: buildHeaders(),
        });

        if (!response.ok) {
          const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;
          const message =
            errorPayload?.error?.message ??
            `Request failed with status ${response.status}`;
          const error = new Error(message) as HttpClientError;
          error.status = response.status;
          error.code = errorPayload?.error?.code;
          error.traceId = errorPayload?.error?.trace_id;
          throw error;
        }

        return (await response.json()) as T;
      },
      async postJson<T>(path: string, options?: { body?: unknown; query?: Record<string, QueryParamValue> }) {
        const queryString = buildQueryString(options?.query);
        const headers = buildHeaders();
        headers["Content-Type"] = "application/json";
        const response = await fetch(`${apiBaseUrl}${path}${queryString ? `?${queryString}` : ""}`, {
          method: "POST",
          credentials: "include",
          headers,
          body: options?.body === undefined ? undefined : JSON.stringify(options.body),
        });

        if (!response.ok) {
          const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;
          const message =
            errorPayload?.error?.message ??
            `Request failed with status ${response.status}`;
          const error = new Error(message) as HttpClientError;
          error.status = response.status;
          error.code = errorPayload?.error?.code;
          error.traceId = errorPayload?.error?.trace_id;
          throw error;
        }

        return (await response.json()) as T;
      },
    };

    return {
      apiBaseUrl,
      bearerToken,
      httpClient,
    };
  }, [apiBaseUrl, bearerToken]);

  return <AppRuntimeContext.Provider value={runtime}>{children}</AppRuntimeContext.Provider>;
}

export function useAppRuntime() {
  const runtime = useContext(AppRuntimeContext);
  if (runtime === null) {
    throw new Error("AppRuntimeProvider is required");
  }
  return runtime;
}
