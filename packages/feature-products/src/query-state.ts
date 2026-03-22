import type { ProductsPortfolioQuery, ProductsPortfolioSortKey } from "./api";

export const defaultProductsPortfolioQuery: ProductsPortfolioQuery = {
  search: "",
  brand_name: [],
  taxonomy_leaf0_name: [],
  status: [],
  sort_key: "pn_interno",
  sort_direction: "asc",
  limit: 50,
  offset: 0,
};

export function readProductsPortfolioQueryFromUrl(): ProductsPortfolioQuery {
  if (typeof window === "undefined") {
    return defaultProductsPortfolioQuery;
  }

  const params = new URLSearchParams(window.location.search);
  const limit = Number(params.get("limit") ?? defaultProductsPortfolioQuery.limit);
  const offset = Number(params.get("offset") ?? defaultProductsPortfolioQuery.offset);

  return {
    search: params.get("search") ?? defaultProductsPortfolioQuery.search,
    brand_name: readMultiParam(params, "brand_name"),
    taxonomy_leaf0_name: readMultiParam(params, "taxonomy_leaf0_name"),
    status: readMultiParam(params, "status"),
    sort_key: (params.get("sort_key") ?? defaultProductsPortfolioQuery.sort_key) as ProductsPortfolioQuery["sort_key"],
    sort_direction: (params.get("sort_direction") ?? defaultProductsPortfolioQuery.sort_direction) as ProductsPortfolioQuery["sort_direction"],
    limit: Number.isFinite(limit) && limit > 0 ? limit : defaultProductsPortfolioQuery.limit,
    offset: Number.isFinite(offset) && offset >= 0 ? offset : defaultProductsPortfolioQuery.offset,
  };
}

export function writeProductsPortfolioQueryToUrl(query: ProductsPortfolioQuery) {
  if (typeof window === "undefined") {
    return;
  }

  const params = new URLSearchParams();
  if (query.search.trim() !== "") params.set("search", query.search.trim());
  appendMultiParam(params, "brand_name", query.brand_name);
  appendMultiParam(params, "taxonomy_leaf0_name", query.taxonomy_leaf0_name);
  appendMultiParam(params, "status", query.status);
  if (query.sort_key !== defaultProductsPortfolioQuery.sort_key) params.set("sort_key", query.sort_key);
  if (query.sort_direction !== defaultProductsPortfolioQuery.sort_direction) params.set("sort_direction", query.sort_direction);
  if (query.limit !== defaultProductsPortfolioQuery.limit) params.set("limit", String(query.limit));
  if (query.offset !== defaultProductsPortfolioQuery.offset) params.set("offset", String(query.offset));

  const nextSearch = params.toString();
  const nextUrl = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}`;
  window.history.replaceState(null, "", nextUrl);
}

function readMultiParam(params: URLSearchParams, key: string): string[] {
  const values = params.getAll(key);
  const out: string[] = [];
  for (const value of values) {
    if (value.includes(",")) {
      for (const part of value.split(",")) {
        const trimmed = part.trim();
        if (trimmed) out.push(trimmed);
      }
      continue;
    }
    const trimmed = value.trim();
    if (trimmed) out.push(trimmed);
  }
  return out;
}

function appendMultiParam(params: URLSearchParams, key: string, values: string[]) {
  for (const value of values) {
    const trimmed = value.trim();
    if (trimmed) params.append(key, trimmed);
  }
}

export function toggleProductsPortfolioSort(
  current: ProductsPortfolioQuery,
  key: ProductsPortfolioSortKey,
): ProductsPortfolioQuery {
  if (current.sort_key === key) {
    return {
      ...current,
      sort_direction: current.sort_direction === "asc" ? "desc" : "asc",
      offset: 0,
    };
  }

  return {
    ...current,
    sort_key: key,
    sort_direction: "asc",
    offset: 0,
  };
}

export function sortIndicator(query: ProductsPortfolioQuery, key: ProductsPortfolioSortKey) {
  if (query.sort_key !== key) {
    return "↕";
  }
  return query.sort_direction === "asc" ? "↑" : "↓";
}
