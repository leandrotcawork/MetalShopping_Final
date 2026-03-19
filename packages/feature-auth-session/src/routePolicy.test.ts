import { describe, expect, it } from "vitest";

import {
  resolveAuthenticatedRouteMode,
  resolveLoginRouteMode,
  shouldAutoRedirect,
} from "./routePolicy";

describe("auth session route policy", () => {
  it("forces authenticated users out of /login", () => {
    const mode = resolveLoginRouteMode({
      status: "authenticated",
      manualMode: false,
    });
    expect(mode).toBe("authenticated");
  });

  it("keeps manual login route when manual mode is enabled", () => {
    const mode = resolveLoginRouteMode({
      status: "unauthenticated",
      manualMode: true,
    });
    expect(mode).toBe("manual");
  });

  it("auto-redirects unauthenticated users from /login when manual mode is disabled", () => {
    const mode = resolveLoginRouteMode({
      status: "unauthenticated",
      manualMode: false,
    });
    expect(mode).toBe("redirect");
  });

  it("renders bootstrap gate while session status is bootstrapping", () => {
    const mode = resolveAuthenticatedRouteMode({ status: "bootstrapping" });
    expect(mode).toBe("bootstrapping");
  });

  it("redirects protected route when session becomes unauthenticated", () => {
    const mode = resolveAuthenticatedRouteMode({ status: "unauthenticated" });
    expect(mode).toBe("redirect");
  });

  it("does not trigger a second auto-redirect when one was already started", () => {
    const shouldRedirect = shouldAutoRedirect({
      enabled: true,
      status: "unauthenticated",
      errorMessage: "",
      alreadyStarted: true,
    });
    expect(shouldRedirect).toBe(false);
  });

  it("suppresses auto-redirect when there is an auth error message", () => {
    const shouldRedirect = shouldAutoRedirect({
      enabled: true,
      status: "unauthenticated",
      errorMessage: "Authentication failed",
      alreadyStarted: false,
    });
    expect(shouldRedirect).toBe(false);
  });
});
