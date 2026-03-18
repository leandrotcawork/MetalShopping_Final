import { Navigate, useSearchParams } from "react-router-dom";

import { AuthRedirectScreen } from "./AuthRedirectScreen";
import { LoginPage } from "./LoginPage";
import { useSession } from "./SessionProvider";

type LoginRoutePageProps = {
  apiBaseUrl: string;
  defaultReturnTo?: string;
  logoSrc: string;
};

export function LoginRoutePage({
  apiBaseUrl,
  defaultReturnTo = "/products",
  logoSrc,
}: LoginRoutePageProps) {
  const { status } = useSession();
  const [searchParams] = useSearchParams();
  const manualMode = searchParams.get("manual") === "1";

  if (status === "authenticated") {
    return <Navigate replace to={defaultReturnTo} />;
  }

  if (!manualMode) {
    return <AuthRedirectScreen returnTo={defaultReturnTo} logoSrc={logoSrc} />;
  }

  return (
    <LoginPage
      apiBaseUrl={apiBaseUrl}
      defaultReturnTo={defaultReturnTo}
      logoSrc={logoSrc}
      autoRedirect={false}
    />
  );
}
