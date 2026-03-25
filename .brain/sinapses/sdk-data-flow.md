---
id: sinapse-sdk-data-flow
title: SDK Data Flow
region: sinapses
tags: [sdk, contracts, code-generation, frontend, cross-cutting]
links:
  - cortex/backend/index
  - cortex/frontend/index
  - hippocampus/conventions
weight: 0.92
updated_at: 2026-03-24T10:00:00Z
---

# SDK Data Flow

How OpenAPI contracts → generated SDK runtime → frontend data access works end-to-end.

## Step 1: Hand-Author OpenAPI Contracts

Developers write API specs in `contracts/api/openapi/`:

```yaml
# contracts/api/openapi/products.yaml
openapi: 3.0.0
info:
  title: Products API
  version: 1.0.0
paths:
  /api/products:
    get:
      operationId: listProducts
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: List of products
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Product'
    post:
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '201':
          description: Product created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Product'

components:
  schemas:
    Product:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        price:
          type: number
      required: [id, name, price]

    CreateProductRequest:
      type: object
      properties:
        name:
          type: string
        price:
          type: number
      required: [name, price]
```

**Rule:** Contracts are the single source of truth. Everything else is generated from them.

## Step 2: Generate SDK & Types

Run the contract generation script:

```bash
./scripts/generate_contract_artifacts.ps1 -Target all
```

This generates SDK artifacts:

```
SDK output (auto-generated)
  ├── TypeScript SDK methods
  │   ├── listProducts(limit?: number): Promise<Product[]>
  │   └── createProduct(req: CreateProductRequest): Promise<Product>
  └── TypeScript type definitions
      ├── interface Product { id: string; name: string; price: number; }
      └── interface CreateProductRequest { name: string; price: number; }
```

**Critical rule:** Never edit generated artifacts by hand. All changes must come from contracts. Re-run the generation script after any contract change.

## Step 3: Build SDK Runtime

The SDK runtime wraps generated types with HTTP client logic:

```typescript
// @metalshopping/sdk-runtime
export class ProductsSDK {
  private baseUrl: string;
  private auth: AuthProvider;

  async listProducts(limit?: number): Promise<Product[]> {
    return this.get(`/api/products?limit=${limit || 20}`);
  }

  async createProduct(req: CreateProductRequest): Promise<Product> {
    return this.post(`/api/products`, req);
  }

  private async get<T>(path: string): Promise<T> {
    const headers = await this.auth.getHeaders();
    const response = await fetch(this.baseUrl + path, { headers });
    if (!response.ok) throw new Error(`${response.status}`);
    return response.json();
  }

  private async post<T>(path: string, body: any): Promise<T> {
    const headers = await this.auth.getHeaders();
    const response = await fetch(this.baseUrl + path, {
      method: 'POST',
      headers: { ...headers, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!response.ok) throw new Error(`${response.status}`);
    return response.json();
  }
}

export const sdk = new ProductsSDK(
  process.env.REACT_APP_API_URL || 'http://localhost:8080',
  authProvider
);
```

## Step 4: Frontend Uses SDK in Components

Components import the SDK and use it for data access:

```typescript
// apps/web/src/pages/ProductList.tsx
import { sdk } from '@metalshopping/sdk-runtime';

export function ProductList() {
  const [products, setProducts] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;

    sdk.products
      .listProducts(20)
      .then(data => { if (!cancelled) setProducts(data); })
      .catch(err => { if (!cancelled) setError(err.message); })
      .finally(() => { if (!cancelled) setLoading(false); });

    return () => { cancelled = true; };
  }, []);

  if (loading) return <Spinner />;
  if (error) return <Error message={error} />;
  if (!products?.length) return <Empty />;

  return (
    <ul>
      {products.map(p => (
        <li key={p.id}>{p.name} — ${p.price.toFixed(2)}</li>
      ))}
    </ul>
  );
}
```

## Contract Changes Ripple to Frontend

When a contract changes:

```yaml
# CHANGED: Added quantity to Product schema
components:
  schemas:
    Product:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        price:
          type: number
        quantity:          # ← NEW FIELD
          type: integer
      required: [id, name, price, quantity]
```

```bash
./scripts/generate_contract_artifacts.ps1 -Target all
```

TypeScript compilation now fails:

```typescript
// ✗ COMPILE ERROR: quantity is now required
products.map(p => p.quantity)  // Must handle missing field or update type
```

**This is correct behavior.** Contract changes force frontend updates. No stale code slips through.

## End-to-End Flow

```
1. Backend developer writes OpenAPI spec in contracts/api/openapi/
2. CI validates contract against JSON Schema
3. Developer runs generate_contract_artifacts.ps1
4. SDK runtime and type definitions are regenerated
5. Frontend app imports types and methods from SDK runtime
6. TypeScript compilation validates frontend code against schema
7. Runtime SDK handles HTTP calls, error handling, auth headers
8. Components use the typed SDK methods with full autocomplete
```

## Advantages Over Raw fetch()

| Aspect | Raw fetch() | SDK Pattern |
|--------|-----------|------------|
| **Type safety** | None; any field access succeeds at compile time | Full type checking; refactoring is safe |
| **Contract changes** | Stale code still compiles; runtime errors in prod | Compile-time errors; forced updates |
| **Auth headers** | Manually added to every fetch() | SDK handles transparently |
| **Error handling** | Each component implements own retry logic | Centralized error policy |
| **API versioning** | Hard to coordinate v1 vs v2 | Contract manages versions |

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Fix |
|-------------|------|---|
| Manual edits to generated SDK artifacts | Changes are overwritten on next generation run | All changes go to contracts (source of truth) |
| Using `fetch()` in components | Bypasses contract, creates type-unsafe code | Always use SDK methods |
| Not updating contract when API changes | Frontend and backend specs diverge | Update contract first, regenerate, then implement |
| Accepting any JSON from SDK without validation | Runtime data might not match schema | SDK methods are already type-checked; trust them |

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.92
