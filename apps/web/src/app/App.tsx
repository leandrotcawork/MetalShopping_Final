import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import {
  AuthenticatedRoute,
  LoginRoutePage,
  SessionProvider,
} from "@metalshopping/feature-auth-session";
import { ProductsPortfolioPage } from "@metalshopping/feature-products";
import { AppFrame } from "@metalshopping/ui";

import logoMetalNobre from "../assets/logo_metal_nobre.svg";
import { AppShell } from "./layouts/AppShell";
import { AppRuntimeProvider, useAppRuntime } from "./providers/AppRuntimeProvider";

function ProductsRoute() {
  const { sdk } = useAppRuntime();

  return <ProductsPortfolioPage api={sdk.products} />;
}

function PlaceholderRoute(props: { title: string; subtitle: string }) {
  return <AppFrame eyebrow="MetalShopping" title={props.title} subtitle={props.subtitle} />;
}

function RoutedApp() {
  const { sdk } = useAppRuntime();

  return (
    <SessionProvider api={sdk.authSession} defaultReturnTo="/products">
      <BrowserRouter>
        <Routes>
          <Route
            path="/login"
            element={<LoginRoutePage defaultReturnTo="/products" logoSrc={logoMetalNobre} />}
          />
          <Route element={<AuthenticatedRoute defaultReturnTo="/products" logoSrc={logoMetalNobre} />}>
            <Route element={<AppShell />}>
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
