import { createContext, useCallback, useContext, useMemo, useRef, useState, type PropsWithChildren } from "react";

import { buildMockAnalyticsHomeDto, buildMockTaxonomyScopeOverview } from "../../mocks/analyticsMocks";
import {
  buildMockAnalyticsProductWorkspaceDto,
  buildMockAnalyticsProductsIndexDto,
  buildMockAnalyticsProductsOverviewDto,
} from "../../mocks/analyticsProductsMocks";

type AnalyticsHomeSnapshot = {
  data: Record<string, unknown>;
  asOf: string;
  updatedAt: number;
} | null;

type ProductsOverviewSnapshot = {
  data: Record<string, unknown>;
  updatedAt: number;
} | null;

type TaxonomyScopeSnapshot = {
  data: Record<string, unknown>;
  updatedAt: number;
} | null;

type AppSessionValue = {
  api: {
    home: {
      workspace: (snapshotId?: string, options?: Record<string, unknown>) => Promise<{ data: Record<string, unknown> }>;
      operational: () => Promise<{ data: { operational: Record<string, unknown> } }>;
    };
    analytics: {
      meta: () => Promise<{ data: any }>;
      workspaceTaxonomyIndex: (snapshotId?: string, options?: Record<string, unknown>) => Promise<{ data: { rows: Record<string, unknown>[] } }>;
      workspaceTaxonomyScope: (params?: Record<string, unknown>) => Promise<{ data: Record<string, unknown> }>;
    };
    products: {
      analyticsIndex: (params?: Record<string, unknown>) => Promise<{ data: any }>;
      analyticsOverview: (params?: Record<string, unknown>) => Promise<{ data: any }>;
      workspace: (pn: string) => Promise<{ data: any }>;
      workspaceInsights: (pn: string) => Promise<{ data: any }>;
    };
    taxonomy: {
      levels: (params?: Record<string, unknown>) => Promise<{ data: { rows: Record<string, unknown>[] } }>;
      tree: (params?: Record<string, unknown>) => Promise<{ data: { roots: Record<string, unknown>[] } }>;
      node: (nodeId: number, params?: Record<string, unknown>) => Promise<{ data: Record<string, unknown> }>;
    };
  };
  shellState: "sidecar_online" | "sidecar_offline";
  analyticsHomeSnapshot: AnalyticsHomeSnapshot;
  setAnalyticsHomeSnapshot: (next: AnalyticsHomeSnapshot) => void;
  getProductsOverviewSnapshot: (key: string) => ProductsOverviewSnapshot;
  setProductsOverviewSnapshot: (key: string, dto: Record<string, unknown>) => void;
  getTaxonomyScopeSnapshot: (key: string) => TaxonomyScopeSnapshot;
  setTaxonomyScopeSnapshot: (key: string, dto: Record<string, unknown>) => void;
};

const AppSessionContext = createContext<AppSessionValue | null>(null);

const MOCK_DTO = buildMockAnalyticsHomeDto();
const MOCK_SCOPE = buildMockTaxonomyScopeOverview();

const MOCK_LEVELS = [
  { level: 0, label: "Departamento", short_label: "Dep", is_enabled: true },
  { level: 1, label: "Categoria", short_label: "Cat", is_enabled: true },
  { level: 2, label: "Subcategoria", short_label: "Sub", is_enabled: true },
];

const MOCK_TREE = [
  {
    id: 1,
    name: "Revestimentos",
    parent_id: null,
    level: 0,
    is_active: true,
    children: [
      { id: 11, name: "Porcelanato", parent_id: 1, level: 1, is_active: true, children: [] },
      { id: 12, name: "Cerâmico", parent_id: 1, level: 1, is_active: true, children: [] },
    ],
  },
  {
    id: 2,
    name: "Hidráulica",
    parent_id: null,
    level: 0,
    is_active: true,
    children: [
      { id: 21, name: "Conexões", parent_id: 2, level: 1, is_active: true, children: [] },
      { id: 22, name: "Torneiras", parent_id: 2, level: 1, is_active: true, children: [] },
    ],
  },
];

const MOCK_RANKING_ROWS = [
  {
    taxonomy_leaf_id: "11",
    taxonomy_leaf_name: "Porcelanato",
    operational_severity: "HIGH",
    operational_severity_rank: 1,
    actionability_score: 87.5,
    products_metrics: {
      capital_brl_total: 120000,
      potential_revenue_brl_total_market: 210000,
      weighted_margin_pct_total: 24.1,
    },
  },
  {
    taxonomy_leaf_id: "22",
    taxonomy_leaf_name: "Torneiras",
    operational_severity: "MEDIUM",
    operational_severity_rank: 2,
    actionability_score: 64.8,
    products_metrics: {
      capital_brl_total: 86000,
      potential_revenue_brl_total_market: 175000,
      weighted_margin_pct_total: 21.7,
    },
  },
];

function findNodeById(nodes: Array<Record<string, unknown>>, nodeId: number): Record<string, unknown> | null {
  for (const node of nodes) {
    const id = Number(node.id);
    if (id === nodeId) return node;
    const children = Array.isArray(node.children) ? (node.children as Array<Record<string, unknown>>) : [];
    const found = findNodeById(children, nodeId);
    if (found) return found;
  }
  return null;
}

export function AppSessionProvider(props: PropsWithChildren) {
  const [analyticsHomeSnapshot, setAnalyticsHomeSnapshot] = useState<AnalyticsHomeSnapshot>({
    data: MOCK_DTO,
    asOf: String((MOCK_DTO.snapshot as { as_of?: string }).as_of || ""),
    updatedAt: Date.now(),
  });
  const [productsOverviewSnapshots, setProductsOverviewSnapshots] = useState<Record<string, ProductsOverviewSnapshot>>({});
  const productsOverviewSnapshotsRef = useRef(productsOverviewSnapshots);
  productsOverviewSnapshotsRef.current = productsOverviewSnapshots;
  const [taxonomyScopeSnapshots, setTaxonomyScopeSnapshots] = useState<Record<string, TaxonomyScopeSnapshot>>({});
  const taxonomyScopeSnapshotsRef = useRef(taxonomyScopeSnapshots);
  taxonomyScopeSnapshotsRef.current = taxonomyScopeSnapshots;

  const getProductsOverviewSnapshot = useCallback((key: string): ProductsOverviewSnapshot => {
    return productsOverviewSnapshotsRef.current[key] || null;
  }, []);

  const setProductsOverviewSnapshot = useCallback((key: string, dto: Record<string, unknown>) => {
    setProductsOverviewSnapshots((current) => {
      const previous = current[key];
      if (previous?.data === dto) return current;
      const next = {
        ...current,
        [key]: { data: dto, updatedAt: Date.now() },
      };
      productsOverviewSnapshotsRef.current = next;
      return next;
    });
  }, []);

  const getTaxonomyScopeSnapshot = useCallback((key: string): TaxonomyScopeSnapshot => {
    return taxonomyScopeSnapshotsRef.current[key] || null;
  }, []);

  const setTaxonomyScopeSnapshot = useCallback((key: string, dto: Record<string, unknown>) => {
    setTaxonomyScopeSnapshots((current) => {
      const previous = current[key];
      if (previous?.data === dto) return current;
      const next = {
        ...current,
        [key]: { data: dto, updatedAt: Date.now() },
      };
      taxonomyScopeSnapshotsRef.current = next;
      return next;
    });
  }, []);

  const api = useMemo<AppSessionValue["api"]>(
    () => ({
      home: {
        workspace: async () => ({ data: MOCK_DTO }),
        operational: async () => ({ data: { operational: {} } }),
      },
      analytics: {
        meta: async () => ({
          data: {
            as_of: "2026-03-22",
            windows: { primary_months: 6, short_months: 3 },
          },
        }),
        workspaceTaxonomyIndex: async () => ({ data: { rows: MOCK_RANKING_ROWS } }),
        workspaceTaxonomyScope: async () => ({ data: MOCK_SCOPE as unknown as Record<string, unknown> }),
      },
      products: {
        analyticsIndex: async (params?: Record<string, unknown>) => ({
          data: buildMockAnalyticsProductsIndexDto({
            search: typeof params?.search === "string" ? params.search : undefined,
            marca: typeof params?.marca === "string" ? params.marca : undefined,
            taxonomyLeafName: typeof params?.taxonomyLeafName === "string" ? params.taxonomyLeafName : undefined,
            status: typeof params?.status === "string" ? params.status : undefined,
            limit: typeof params?.limit === "number" ? params.limit : undefined,
            offset: typeof params?.offset === "number" ? params.offset : undefined,
          }),
        }),
        analyticsOverview: async () => ({
          data: buildMockAnalyticsProductsOverviewDto(),
        }),
        workspace: async (pn: string) => ({
          data: { pn, ...buildMockAnalyticsProductWorkspaceDto(pn) },
        }),
        workspaceInsights: async () => ({
          data: { insights: null },
        }),
      },
      taxonomy: {
        levels: async () => ({ data: { rows: MOCK_LEVELS } }),
        tree: async () => ({ data: { roots: MOCK_TREE } }),
        node: async (nodeId: number) => {
          const node = findNodeById(MOCK_TREE as unknown as Array<Record<string, unknown>>, nodeId);
          const children = Array.isArray(node?.children) ? (node.children as Array<Record<string, unknown>>) : [];
          return {
            data: {
              node: node || null,
              breadcrumbs: node ? [node] : [],
              children,
            },
          };
        },
      },
    }),
    [],
  );

  const value = useMemo<AppSessionValue>(
    () => ({
      api,
      shellState: "sidecar_online",
      analyticsHomeSnapshot,
      setAnalyticsHomeSnapshot,
      getProductsOverviewSnapshot,
      setProductsOverviewSnapshot,
      getTaxonomyScopeSnapshot,
      setTaxonomyScopeSnapshot,
    }),
    [
      analyticsHomeSnapshot,
      api,
      getProductsOverviewSnapshot,
      setProductsOverviewSnapshot,
      getTaxonomyScopeSnapshot,
      setTaxonomyScopeSnapshot,
    ],
  );

  return <AppSessionContext.Provider value={value}>{props.children}</AppSessionContext.Provider>;
}

export function useAppSession(): AppSessionValue {
  const value = useContext(AppSessionContext);
  if (!value) {
    throw new Error("useAppSession deve ser usado dentro de AppSessionProvider.");
  }
  return value;
}
