---
id: cortex-frontend-index
title: Frontend Domain
region: cortex/frontend
tags: [frontend, react, sdk-driven, thin-client]
links:
  - hippocampus/architecture
  - hippocampus/conventions
  - hippocampus/decisions_log
weight: 0.85
updated_at: 2026-03-24T10:00:00Z
---

# Frontend Domain

MetalShopping's React 18 frontend is a **thin-client** that derives all data from the contract-generated SDK. Logic lives in feature packages; `apps/web` is an orchestration shell.

## Architecture

### Thin-Client Pattern

The frontend:
- **Never makes raw `fetch()` calls** — all data flows through `@metalshopping/sdk-runtime`
- **Hosts no business logic** — services live in the Go backend, not in React components
- **Assembles features from packages** — `packages/feature-*` modules compose into pages

### Feature Packages

Each feature surface (Analytics, Products, CRM, etc.) lives in `packages/feature-<name>/`:

```
packages/feature-products/
  ├── src/
  │   ├── components/     Feature-specific components
  │   ├── hooks/          useProduct, useProductList, etc.
  │   ├── routes/         Page definitions
  │   ├── index.ts        Entrypoint
  │   └── styles/         CSS modules (design tokens only)
  └── package.json
```

**Rule:** Features are independent. They should never import from other feature packages.

### SDK-First Data Access

Every component that fetches data uses the generated SDK:

```typescript
import { sdk } from '@metalshopping/sdk-runtime';

function ProductList() {
  const [products, setProducts] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;

    sdk.products.list()
      .then(data => { if (!cancelled) setProducts(data); })
      .catch(err => { if (!cancelled) setError(err); })
      .finally(() => { if (!cancelled) setLoading(false); });

    return () => { cancelled = true; };
  }, []);

  if (loading) return <Spinner />;
  if (error) return <Error message={error.message} />;
  if (!products?.length) return <Empty />;
  return <ul>{products.map(p => <li key={p.id}>{p.name}</li>)}</ul>;
}
```

**Pattern:** useEffect + cancelled flag is standard. Always load + error + empty.

## Design System & Styling

### Tokens (No Hardcoded Values)

All colors, spacing, typography come from `@metalshopping/design-system`:

```typescript
import tokens from '@metalshopping/design-system/tokens.css';
// Use CSS variables: var(--surface), var(--text-primary), var(--spacing-md), etc.
```

**Anti-pattern:** Using hex values (`#2E2E2E`) directly in component styles.

### Shared UI Components

Before creating a component, check `packages/ui/src/index.ts`:

```typescript
// packages/ui/src/index.ts
export { Button } from './Button';
export { Card } from './Card';
export { Table } from './Table';
export { Modal } from './Modal';
// ... etc
```

**Rule:** Reuse from `packages/ui` before building new ones. This keeps the UI consistent.

### CSS Modules

Feature styles use CSS Modules with design tokens:

```css
/* packages/feature-products/src/styles/ProductCard.module.css */
.card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  padding: var(--spacing-lg);
}

.title {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}
```

## Async State Management

Three explicit states for every data-fetching component:

1. **Loading** — show spinner or skeleton
2. **Error** — show error message with retry option
3. **Empty** — show empty state if data is empty
4. **Success** — render the data

Never render without explicitly checking all four.

## Routing

Routes are defined per feature and composed in `apps/web/src/router.ts`:

```typescript
// packages/feature-products/src/routes.tsx
export const routes = [
  { path: '/products', element: <ProductsPage /> },
  { path: '/products/:id', element: <ProductDetailPage /> },
];

// apps/web/src/router.ts
import { routes as productRoutes } from '@metalshopping/feature-products';
export const router = createBrowserRouter([
  ...productRoutes,
  ...analyticsRoutes,
  ...crmRoutes,
]);
```

## Known Pitfalls

These lessons capture repeated mistakes in frontend development:

- [[lessons/lesson-0005]] — Frontend data flow must use platform SDK contracts
- [[lessons/lesson-0006]] — Reuse design system before adding UI primitives
- [[lessons/lesson-0010]] — Legacy migration follows parity-first sequencing
- [[lessons/lesson-0015]] — Legacy migration must preserve interactive behavior
- [[lessons/lesson-0023]] — Feature code must import shared UI from package entrypoint

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.85
