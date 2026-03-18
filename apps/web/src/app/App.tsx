import { useMemo } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { LoginPage, SessionProvider, useSession } from "@metalshopping/feature-auth-session";
import { ProductsPortfolioPage } from "@metalshopping/feature-products";
import { createServerCoreSdk } from "@metalshopping/generated-sdk";
import { AppFrame } from "@metalshopping/ui";

import logoMetalNobre from "../assets/logo_metal_nobre.svg";
import { AppShell } from "./layouts/AppShell";
import { AppRuntimeProvider, useAppRuntime } from "./providers/AppRuntimeProvider";

function ProductsRoute() {
  const { httpClient } = useAppRuntime();
  const sdk = useMemo(() => createServerCoreSdk(httpClient), [httpClient]);

  return <ProductsPortfolioPage api={sdk.products} />;
}

function PlaceholderRoute(props: { title: string; subtitle: string }) {
  return <AppFrame eyebrow="MetalShopping" title={props.title} subtitle={props.subtitle} />;
}

function AppBootstrapScreen() {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        background:
          "radial-gradient(circle at 15% 15%, rgba(145, 19, 42, 0.12), transparent 24%), linear-gradient(180deg, #fdfafc 0%, #f5eef2 100%)",
      }}
    >
      <div
        style={{
          display: "grid",
          gap: "10px",
          justifyItems: "center",
          color: "#5f1227",
        }}
      >
        <strong style={{ fontSize: "1.1rem", letterSpacing: "0.02em" }}>MetalShopping</strong>
        <span style={{ color: "#6f676a", fontSize: "0.95rem" }}>Validando sessao ativa...</span>
      </div>
    </div>
  );
}

function LoginRoute() {
  const { apiBaseUrl, httpClient } = useAppRuntime();
  const { status } = useSession();

  if (status === "authenticated") {
    return <Navigate replace to="/products" />;
  }

  return <LoginPage apiBaseUrl={apiBaseUrl} defaultReturnTo="/products" logoSrc={logoMetalNobre} />;
}

function RequireAuthenticated() {
  const { status } = useSession();

  if (status === "bootstrapping") {
    return <AppBootstrapScreen />;
  }

  if (status === "unauthenticated" || status === "starting_login") {
    return <Navigate replace to="/login" />;
  }

  return <AppShell />;
}

function RoutedApp() {
  const { apiBaseUrl, httpClient } = useAppRuntime();
  const sdk = useMemo(() => createServerCoreSdk(httpClient), [httpClient]);

  return (
    <SessionProvider api={sdk.authSession} apiBaseUrl={apiBaseUrl} defaultReturnTo="/products">
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginRoute />} />
          <Route element={<RequireAuthenticated />}>
            <Route index element={<Navigate replace to="/products" />} />
            <Route
              path="home"
              element={
                <PlaceholderRoute
                  title="Home"
                  subtitle="Surface reservada. Abertura bloqueada ate fecharmos o hardening de Products."
                />
              }
            />
            <Route path="products" element={<ProductsRoute />} />
            <Route
              path="shopping"
              element={
                <PlaceholderRoute
                  title="Shopping de Precos"
                  subtitle="Surface reservada. Abertura bloqueada ate fecharmos o hardening de Products."
                />
              }
            />
            <Route
              path="analytics"
              element={
                <PlaceholderRoute
                  title="Analytics"
                  subtitle="Surface reservada. Abertura bloqueada ate fecharmos o hardening de Products."
                />
              }
            />
            <Route
              path="settings"
              element={
                <PlaceholderRoute
                  title="Configuracoes"
                  subtitle="Surface reservada. Abertura bloqueada ate fecharmos o hardening de Products."
                />
              }
            />
          </Route>
          <Route path="*" element={<Navigate replace to="/products" />} />
        </Routes>
      </BrowserRouter>
    </SessionProvider>
  );
}

export function App() {
  return (
    <AppRuntimeProvider>
      <RoutedApp />
    </AppRuntimeProvider>
  );
}
