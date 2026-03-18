import { useMemo } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { createProductsPortfolioApi, ProductsPortfolioPage } from "@metalshopping/feature-products";
import { AppFrame } from "@metalshopping/ui";

import { AppShell } from "./layouts/AppShell";
import { AppRuntimeProvider, useAppRuntime } from "./providers/AppRuntimeProvider";

function ProductsRoute() {
  const { httpClient } = useAppRuntime();
  const api = useMemo(() => createProductsPortfolioApi(httpClient), [httpClient]);

  return <ProductsPortfolioPage api={api} />;
}

function PlaceholderRoute(props: { title: string; subtitle: string }) {
  return (
    <AppFrame
      eyebrow="MetalShopping"
      title={props.title}
      subtitle={props.subtitle}
    />
  );
}

function RoutedApp() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AppShell />}>
          <Route index element={<Navigate replace to="/products" />} />
          <Route
            path="home"
            element={
              <PlaceholderRoute
                title="Home"
                subtitle="Surface reservada. Abertura bloqueada atÃ© fecharmos o hardening de Products."
              />
            }
          />
          <Route path="products" element={<ProductsRoute />} />
          <Route
            path="shopping"
            element={
              <PlaceholderRoute
                title="Shopping de PreÃ§os"
                subtitle="Surface reservada. Abertura bloqueada atÃ© fecharmos o hardening de Products."
              />
            }
          />
          <Route
            path="analytics"
            element={
              <PlaceholderRoute
                title="Analytics"
                subtitle="Surface reservada. Abertura bloqueada atÃ© fecharmos o hardening de Products."
              />
            }
          />
          <Route
            path="settings"
            element={
              <PlaceholderRoute
                title="ConfiguraÃ§Ãµes"
                subtitle="Surface reservada. Abertura bloqueada atÃ© fecharmos o hardening de Products."
              />
            }
          />
          <Route path="*" element={<Navigate replace to="/products" />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export function App() {
  return (
    <AppRuntimeProvider>
      <RoutedApp />
    </AppRuntimeProvider>
  );
}
