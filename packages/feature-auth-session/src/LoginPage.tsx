import { useMemo } from "react";

import { Button } from "@metalshopping/ui";

import { useSession } from "./SessionProvider";
import styles from "./LoginPage.module.css";

type LoginPageProps = {
  apiBaseUrl: string;
  defaultReturnTo?: string;
  logoSrc: string;
};

const loginHighlights = [
  { value: "OIDC", label: "Login corporativo" },
  { value: "RLS", label: "Tenant isolado" },
  { value: "IAM", label: "Permissoes no core" },
];

export function LoginPage({ apiBaseUrl, defaultReturnTo = "/products", logoSrc }: LoginPageProps) {
  const { errorMessage, login, status } = useSession();

  const helperCopy = useMemo(() => {
    switch (status) {
      case "bootstrapping":
        return "Validando se ja existe uma sessao autenticada no navegador.";
      case "starting_login":
        return "Redirecionando para o provedor de identidade configurado.";
      default:
        return "Entre com sua conta segura para acessar as superficies operacionais do MetalShopping.";
    }
  }, [status]);

  return (
    <section className={styles.page}>
      <aside className={styles.brandPanel}>
        <div className={styles.brandContent}>
          <p className={styles.eyebrow}>Metal Nobre Acabamentos</p>
          <h1 className={styles.title}>
            Precificacao
            <br />
            <span>inteligente.</span>
          </h1>
          <p className={styles.subtitle}>
            Compare precos de mercado, acompanhe tendencias e opere com uma sessao
            web protegida pelo backend do MetalShopping.
          </p>

          <div className={styles.stats}>
            {loginHighlights.map((item) => (
              <div key={item.label} className={styles.statCard}>
                <strong>{item.value}</strong>
                <span>{item.label}</span>
              </div>
            ))}
          </div>
        </div>
      </aside>

      <main className={styles.formArea}>
        <div className={styles.card}>
          <header className={styles.header}>
            <img className={styles.logo} src={logoSrc} alt="Metal Nobre Acabamentos" />
            <h2>
              Bem-vindo ao Metal<span>Shopping</span>
            </h2>
            <p>{helperCopy}</p>
          </header>

          <div className={styles.sessionPanel}>
            <div className={styles.copyBlock}>
              <p className={styles.panelEyebrow}>Sessao protegida</p>
              <h3 className={styles.panelTitle}>Entrar com conta segura</h3>
              <p className={styles.panelText}>
                O navegador nao armazena token da aplicacao. A autenticacao e a
                sessao sao controladas pelo backend com cookie HttpOnly.
              </p>
            </div>

            <ul className={styles.featureList}>
              <li>OIDC e issuer externo reais</li>
              <li>Tenancy e autorizacao resolvidas no server_core</li>
              <li>Mesmo modelo de identidade para web e app futuro</li>
            </ul>

            {errorMessage ? <p className={`${styles.alert} ${styles.alertError}`}>{errorMessage}</p> : null}

            <Button
              variant="primary"
              className={styles.submitButton}
              onClick={() => login(defaultReturnTo)}
              disabled={status === "bootstrapping" || status === "starting_login"}
            >
              {status === "starting_login" ? "Redirecionando..." : "Entrar com identidade segura"}
            </Button>
          </div>

          <footer className={styles.footer}>
            <span>
              Metal<span>Shopping</span> v3.0
            </span>
            <small>API: {apiBaseUrl}</small>
          </footer>
        </div>
      </main>
    </section>
  );
}
