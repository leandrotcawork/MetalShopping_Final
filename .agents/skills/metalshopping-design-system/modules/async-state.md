# async-state.md — Loading e Error State

## Problema atual
O codebase tem três padrões inconsistentes para erro:
1. `ShoppingPage`: `<p className={styles.error}>` inline no container
2. `HomePage`: `<p className={styles.error}>` com margin no stack
3. `ProductsHero`: passa erro para `StatusBanner` via props

Loading é sempre texto inline sem componente. Isso causa inconsistência visual.

## Padrão canônico a adotar

### Para erro inline no stack de page
Use `StatusBanner` de `@metalshopping/ui` com `tone="error"`:

```tsx
import { StatusBanner } from "@metalshopping/ui";

{error ? <StatusBanner tone="error">{error}</StatusBanner> : null}
```

### Para loading inline no stack
Enquanto não existe `LoadingState` em `packages/ui`, use o padrão de texto com SurfaceCard:

```tsx
{loading ? (
  <SurfaceCard>
    <p style={{ margin: 0, color: "var(--ms-ink-500)", fontSize: "0.88rem" }}>
      Carregando...
    </p>
  </SurfaceCard>
) : null}
```

### Para empty state em tabela ou lista
```tsx
// dentro de tbody
<tr>
  <td colSpan={N} style={{ padding: 32, textAlign: "center", color: "var(--ms-ink-500)" }}>
    {loading ? "Carregando..." : "Nenhum item encontrado para o filtro atual."}
  </td>
</tr>

// dentro de lista/div
{items.length === 0 && !loading ? (
  <p style={{ margin: 0, color: "var(--ms-ink-500)", fontSize: "0.88rem" }}>
    Nenhum item encontrado.
  </p>
) : null}
```

## Padrão de fetch — useEffect com cancelled flag

Este é o padrão obrigatório para todo fetch em page ou componente de feature.
Nunca faça fetch sem o `cancelled` flag — causa setState em componente desmontado.

```tsx
const [data, setData] = useState<MyType | null>(null);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);

useEffect(() => {
  let cancelled = false;

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const result = await api.getSomething();
      if (!cancelled) setData(result);
    } catch (err) {
      if (!cancelled) {
        setError(err instanceof Error ? err.message : "Falha ao carregar.");
      }
    } finally {
      if (!cancelled) setLoading(false);
    }
  }

  void load();
  return () => { cancelled = true; };
}, [api, dependency]);
```

## Candidato futuro para packages/ui

Quando o padrão se repetir 3+ vezes, criar:
- `LoadingState` — spinner ou texto, sem domínio
- `ErrorState` — wrapper de StatusBanner tone="error" com mensagem

Não criar agora. Documentar aqui até atingir o threshold.
