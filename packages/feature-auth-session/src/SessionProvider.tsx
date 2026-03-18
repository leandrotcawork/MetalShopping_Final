import type { PropsWithChildren } from "react";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import type { AuthSessionStateV1 } from "@metalshopping/generated-types";

import type { AuthSessionApi, AuthSessionContextValue, AuthSessionStatus } from "./types";

type SessionProviderProps = PropsWithChildren<{
  api: AuthSessionApi;
  apiBaseUrl: string;
  defaultReturnTo?: string;
}>;

type HttpLikeError = Error & {
  status?: number;
  code?: string;
};

const AuthSessionContext = createContext<AuthSessionContextValue | null>(null);

function errorStatus(error: unknown): number | null {
  if (typeof error === "object" && error !== null && "status" in error) {
    const value = (error as { status?: unknown }).status;
    if (typeof value === "number") {
      return value;
    }
  }
  return null;
}

function errorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim() !== "") {
    return error.message;
  }
  return fallback;
}

export function SessionProvider({
  api,
  apiBaseUrl,
  defaultReturnTo = "/products",
  children,
}: SessionProviderProps) {
  const [status, setStatus] = useState<AuthSessionStatus>("bootstrapping");
  const [session, setSession] = useState<AuthSessionStateV1 | null>(null);
  const [message, setMessage] = useState("");

  const bootstrap = useCallback(async () => {
    try {
      const nextSession = await api.getSessionState();
      setSession(nextSession);
      setStatus("authenticated");
      setMessage("");
    } catch (error) {
      if (errorStatus(error) === 401) {
        setSession(null);
        setStatus("unauthenticated");
        setMessage("");
        return;
      }

      setSession(null);
      setStatus("unauthenticated");
      setMessage(errorMessage(error, "Nao foi possivel validar a sessao atual."));
    }
  }, [api]);

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  const login = useCallback(
    (returnTo?: string) => {
      setStatus("starting_login");
      const target = api.buildStartLoginUrl(apiBaseUrl, {
        return_to: returnTo ?? defaultReturnTo,
      });
      window.location.assign(target);
    },
    [api, apiBaseUrl, defaultReturnTo],
  );

  const logout = useCallback(async () => {
    try {
      await api.logoutSession();
    } finally {
      setSession(null);
      setStatus("unauthenticated");
      setMessage("");
    }
  }, [api]);

  const refresh = useCallback(async () => {
    try {
      const nextSession = await api.refreshSession();
      setSession(nextSession);
      setStatus("authenticated");
      setMessage("");
    } catch (error) {
      if (errorStatus(error) === 401) {
        setSession(null);
        setStatus("unauthenticated");
        setMessage("");
        return;
      }
      throw error as HttpLikeError;
    }
  }, [api]);

  const value = useMemo<AuthSessionContextValue>(
    () => ({
      status,
      session,
      errorMessage: message,
      login,
      logout,
      refresh,
    }),
    [status, session, message, login, logout, refresh],
  );

  return <AuthSessionContext.Provider value={value}>{children}</AuthSessionContext.Provider>;
}

export function useSession() {
  const value = useContext(AuthSessionContext);
  if (value === null) {
    throw new Error("SessionProvider is required");
  }
  return value;
}
