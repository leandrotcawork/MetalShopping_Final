import type { AuthSessionStatus } from "./types";

type ResolveLoginRouteModeInput = {
  status: AuthSessionStatus;
  manualMode: boolean;
};

type ResolveAuthenticatedRouteModeInput = {
  status: AuthSessionStatus;
};

type ShouldAutoRedirectInput = {
  status: AuthSessionStatus;
  errorMessage: string;
  alreadyStarted: boolean;
  enabled: boolean;
};

export function resolveLoginRouteMode(input: ResolveLoginRouteModeInput) {
  if (input.status === "authenticated") {
    return "authenticated";
  }
  if (!input.manualMode) {
    return "redirect";
  }
  return "manual";
}

export function resolveAuthenticatedRouteMode(input: ResolveAuthenticatedRouteModeInput) {
  if (input.status === "bootstrapping") {
    return "bootstrapping";
  }
  if (input.status === "unauthenticated" || input.status === "starting_login") {
    return "redirect";
  }
  return "outlet";
}

export function shouldAutoRedirect(input: ShouldAutoRedirectInput) {
  if (!input.enabled) {
    return false;
  }
  if (input.alreadyStarted) {
    return false;
  }
  if (input.status !== "unauthenticated") {
    return false;
  }
  return input.errorMessage.trim() === "";
}
