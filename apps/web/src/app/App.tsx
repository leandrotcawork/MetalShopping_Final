import { ProductsPortfolioPage } from "@metalshopping/feature-products";

import { AppShell } from "./layouts/AppShell";

export function App() {
  return (
    <AppShell activeItemKey="products">
      <ProductsPortfolioPage />
    </AppShell>
  );
}
