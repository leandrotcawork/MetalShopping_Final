import { useEffect, useRef } from "react";

import { LoginPage } from "./LoginPage";
import { useSession } from "./SessionProvider";
import styles from "./AuthStatusScreen.module.css";

type AuthRedirectScreenProps = {
  returnTo: string;
  logoSrc: string;
};

export function AuthRedirectScreen({ returnTo, logoSrc }: AuthRedirectScreenProps) {
  const { errorMessage, login, status } = useSession();
  const redirectStartedRef = useRef(false);

  useEffect(() => {
    if (status !== "unauthenticated" || errorMessage.trim() !== "" || redirectStartedRef.current) {
      return;
    }

    redirectStartedRef.current = true;
    login(returnTo);
  }, [errorMessage, login, returnTo, status]);

  if (errorMessage.trim() !== "") {
    return (
      <LoginPage
        apiBaseUrl=""
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
