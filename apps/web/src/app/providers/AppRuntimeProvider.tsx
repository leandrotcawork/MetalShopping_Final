import type { PropsWithChildren } from "react";
import { createContext, useContext, useMemo } from "react";

import {
  createServerCoreSdk,
  createBrowserGeneratedHttpClient,
  defaultWebSessionCSRFCookieName,
  defaultWebSessionCSRFHeaderName,
  type GeneratedHttpClient,
  type ServerCoreSdk,
} from "@metalshopping/sdk-runtime";

export type AppHttpClient = GeneratedHttpClient;

type AppRuntime = {
  apiBaseUrl: string;
  httpClient: AppHttpClient;
  sdk: ServerCoreSdk;
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
      csrfCookieName: defaultWebSessionCSRFCookieName,
      csrfHeaderName: defaultWebSessionCSRFHeaderName,
    });
    const sdk = createServerCoreSdk(httpClient);

    return {
      apiBaseUrl,
      httpClient,
      sdk,
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
