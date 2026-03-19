import { Navigate, useSearchParams } from "react-router-dom";

import { AuthRedirectScreen } from "./AuthRedirectScreen";
import { LoginPage } from "./LoginPage";
import { resolveLoginRouteMode } from "./routePolicy";
import { useSession } from "./SessionProvider";

type LoginRoutePageProps = {
  defaultReturnTo?: string;
  logoSrc: string;
};

export function LoginRoutePage({
  defaultReturnTo = "/products",
  logoSrc,
}: LoginRoutePageProps) {
  const { status } = useSession();
  const [searchParams] = useSearchParams();
  const manualMode = searchParams.get("manual") === "1";
  const mode = resolveLoginRouteMode({ status, manualMode });

  if (mode === "authenticated") {
    return <Navigate replace to={defaultReturnTo} />;
  }

  if (mode === "redirect") {
    return <AuthRedirectScreen returnTo={defaultReturnTo} logoSrc={logoSrc} />;
  }

  return (
    <LoginPage
      defaultReturnTo={defaultReturnTo}
      logoSrc={logoSrc}
      autoRedirect={false}
    />
  );
}
