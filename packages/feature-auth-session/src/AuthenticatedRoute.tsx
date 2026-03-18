import { Outlet, useLocation } from "react-router-dom";

import { AuthBootstrapScreen } from "./AuthBootstrapScreen";
import { AuthRedirectScreen } from "./AuthRedirectScreen";
import { useSession } from "./SessionProvider";

type AuthenticatedRouteProps = {
  defaultReturnTo?: string;
  logoSrc: string;
};

export function AuthenticatedRoute({
  defaultReturnTo = "/products",
  logoSrc,
}: AuthenticatedRouteProps) {
  const { status } = useSession();
  const location = useLocation();
  const returnTo = `${location.pathname}${location.search}${location.hash}`;

  if (status === "bootstrapping") {
    return <AuthBootstrapScreen />;
  }

  if (status === "unauthenticated" || status === "starting_login") {
    return (
      <AuthRedirectScreen
        returnTo={returnTo === "/" ? defaultReturnTo : returnTo}
        logoSrc={logoSrc}
      />
    );
  }

  return <Outlet />;
}
