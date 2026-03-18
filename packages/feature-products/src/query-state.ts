import type { ProductsPortfolioQuery, ProductsPortfolioSortDirection, ProductsPortfolioSortKey } from "./api";

export const defaultProductsPortfolioQuery: ProductsPortfolioQuery = {
  search: "",
  brandName: "",
  taxonomyLeaf0Name: "",
  status: "",
  sortKey: "pn_interno",
  sortDirection: "asc",
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
  const sortKey = (params.get("sort_key") ?? defaultProductsPortfolioQuery.sortKey) as ProductsPortfolioSortKey;
  const sortDirection = (params.get("sort_direction") ?? defaultProductsPortfolioQuery.sortDirection) as ProductsPortfolioSortDirection;

  return {
    search: params.get("search") ?? defaultProductsPortfolioQuery.search,
    brandName: params.get("brand_name") ?? defaultProductsPortfolioQuery.brandName,
    taxonomyLeaf0Name: params.get("taxonomy_leaf0_name") ?? defaultProductsPortfolioQuery.taxonomyLeaf0Name,
    status: params.get("status") ?? defaultProductsPortfolioQuery.status,
    sortKey,
    sortDirection,
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
  if (query.brandName.trim() !== "") params.set("brand_name", query.brandName.trim());
  if (query.taxonomyLeaf0Name.trim() !== "") params.set("taxonomy_leaf0_name", query.taxonomyLeaf0Name.trim());
  if (query.status.trim() !== "") params.set("status", query.status.trim());
  if (query.sortKey !== defaultProductsPortfolioQuery.sortKey) params.set("sort_key", query.sortKey);
  if (query.sortDirection !== defaultProductsPortfolioQuery.sortDirection) params.set("sort_direction", query.sortDirection);
  if (query.limit !== defaultProductsPortfolioQuery.limit) params.set("limit", String(query.limit));
  if (query.offset !== defaultProductsPortfolioQuery.offset) params.set("offset", String(query.offset));

  const nextSearch = params.toString();
  const nextUrl = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}`;
  window.history.replaceState(null, "", nextUrl);
}

export function toggleProductsPortfolioSort(
  current: ProductsPortfolioQuery,
  key: ProductsPortfolioSortKey,
): ProductsPortfolioQuery {
  if (current.sortKey === key) {
    return {
      ...current,
      sortDirection: current.sortDirection === "asc" ? "desc" : "asc",
      offset: 0,
    };
  }

  return {
    ...current,
    sortKey: key,
    sortDirection: "asc",
    offset: 0,
  };
}

export function sortIndicator(query: ProductsPortfolioQuery, key: ProductsPortfolioSortKey) {
  if (query.sortKey !== key) {
    return "↕";
  }
  return query.sortDirection === "asc" ? "↑" : "↓";
}
