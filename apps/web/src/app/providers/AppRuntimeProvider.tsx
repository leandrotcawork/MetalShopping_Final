import type { PropsWithChildren } from "react";
import { createContext, useContext, useMemo } from "react";

import type { CommonErrorV1 } from "@metalshopping/generated-types";

type QueryParamValue = string | number | boolean | null | undefined;

export type AppHttpClient = {
  getJson<T>(path: string, options?: { query?: Record<string, QueryParamValue> }): Promise<T>;
};

type AppRuntime = {
  apiBaseUrl: string;
  bearerToken: string;
  httpClient: AppHttpClient;
};

const defaultApiBaseUrl = "http://127.0.0.1:8080";
const defaultBearerToken = "local-dev-token";

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
    const httpClient: AppHttpClient = {
      async getJson<T>(path: string, options?: { query?: Record<string, QueryParamValue> }) {
        const queryString = buildQueryString(options?.query);
        const response = await fetch(`${apiBaseUrl}${path}${queryString ? `?${queryString}` : ""}`, {
          headers: {
            Authorization: `Bearer ${bearerToken}`,
            Accept: "application/json",
          },
        });

        if (!response.ok) {
          const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;
          const message =
            errorPayload?.error?.message ??
            `Request failed with status ${response.status}`;
          throw new Error(message);
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
