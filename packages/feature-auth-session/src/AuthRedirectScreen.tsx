import { useEffect, useRef } from "react";

import { LoginPage } from "./LoginPage";
import { shouldAutoRedirect } from "./routePolicy";
import { useSession } from "./SessionProvider";
import styles from "./AuthStatusScreen.module.css";

type AuthRedirectScreenProps = {
  returnTo: string;
  logoSrc: string;
};

export function AuthRedirectScreen({ returnTo, logoSrc }: AuthRedirectScreenProps) {
  const { errorMessage, login, status } = useSession();
  const redirectStartedRef = useRef(false);
  const canAutoRedirect = shouldAutoRedirect({
    status,
    errorMessage,
    alreadyStarted: redirectStartedRef.current,
    enabled: true,
  });

  useEffect(() => {
    if (!canAutoRedirect) {
      return;
    }

    redirectStartedRef.current = true;
    login(returnTo);
  }, [canAutoRedirect, login, returnTo]);

  if (errorMessage.trim() !== "") {
    return (
      <LoginPage
        defaultReturnTo={returnTo}
        logoSrc={logoSrc}
        autoRedirect={false}
      />
    );
  }

  return (
    <div className={styles.screen}>
      <div className={styles.panel}>
        <strong className={styles.brand}>MetalShopping</strong>
        <span className={styles.message}>
          Encaminhando voce para a autenticacao segura do MetalShopping.
        </span>
        <button type="button" onClick={() => login(returnTo)} className={styles.action}>
          Abrir login manualmente
        </button>
      </div>
    </div>
  );
}
