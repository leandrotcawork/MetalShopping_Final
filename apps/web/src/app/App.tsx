import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import {
  AuthenticatedRoute,
  LoginRoutePage,
  SessionProvider,
} from "@metalshopping/feature-auth-session";
import { ProductsPortfolioPage } from "@metalshopping/feature-products";
import { AppFrame } from "@metalshopping/ui";

import logoMetalNobre from "../assets/logo_metal_nobre.svg";
import { AnalyticsPage } from "../pages/AnalyticsPage";
import { AnalyticsProductsWorkspaceRoute } from "../pages/AnalyticsProductsWorkspacePage";
import { HomePage } from "../pages/HomePage";
import { ShoppingPage } from "../pages/ShoppingPage";
import { AppShell } from "./layouts/AppShell";
import { AppRuntimeProvider, useAppRuntime } from "./providers/AppRuntimeProvider";

function ProductsRoute() {
  const { sdk } = useAppRuntime();

  return <ProductsPortfolioPage api={sdk.products} />;
}

function HomeRoute() {
  const { sdk } = useAppRuntime();

  return <HomePage api={sdk.home} />;
}

function ShoppingRoute() {
  const { sdk } = useAppRuntime();

  return <ShoppingPage shoppingApi={sdk.shopping} productsApi={sdk.products} />;
}

function AnalyticsRoute() {
  return <AnalyticsPage />;
}

function AnalyticsProductsWorkspace() {
  return <AnalyticsProductsWorkspaceRoute />;
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
            <Route index element={<Navigate replace to="/home" />} />
            <Route path="home" element={<HomeRoute />} />
            <Route path="products" element={<ProductsRoute />} />
            <Route path="shopping" element={<ShoppingRoute />} />
            <Route path="analytics/products/:pn/*" element={<AnalyticsProductsWorkspace />} />
            <Route path="analytics/:tab?" element={<AnalyticsRoute />} />
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
