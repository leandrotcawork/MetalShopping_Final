import type { PropsWithChildren } from "react";
import { createContext, useContext, useMemo } from "react";

import {
  createBrowserGeneratedHttpClient,
  type GeneratedHttpClient,
} from "@metalshopping/generated-sdk";

export type AppHttpClient = GeneratedHttpClient;

type AppRuntime = {
  apiBaseUrl: string;
  httpClient: AppHttpClient;
};

const defaultApiBaseUrl = "http://127.0.0.1:8080";
const defaultBearerToken = "";

const AppRuntimeContext = createContext<AppRuntime | null>(null);

export function AppRuntimeProvider({ children }: PropsWithChildren) {
  const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl).replace(/\/$/, "");
  const bearerToken = import.meta.env.VITE_API_BEARER_TOKEN ?? defaultBearerToken;

  const runtime = useMemo<AppRuntime>(() => {
    const httpClient = createBrowserGeneratedHttpClient({
      baseUrl: apiBaseUrl,
      bearerToken,
      csrfCookieName: "ms_web_csrf",
      csrfHeaderName: "X-CSRF-Token",
    });

    return {
      apiBaseUrl,
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
