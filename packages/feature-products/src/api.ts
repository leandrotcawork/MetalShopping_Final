import type {
  ProductsPortfolioItemV1,
  ProductsPortfolioListV1,
} from "@metalshopping/generated-types";

export type ProductsPortfolioSortKey =
  | "pn_interno"
  | "name"
  | "brand_name"
  | "taxonomy_leaf0_name"
  | "product_status"
  | "current_price_amount"
  | "replacement_cost_amount"
  | "on_hand_quantity";

export type ProductsPortfolioSortDirection = "asc" | "desc";

export type ProductsPortfolioQuery = {
  search: string;
  brandName: string;
  taxonomyLeaf0Name: string;
  status: string;
  sortKey: ProductsPortfolioSortKey;
  sortDirection: ProductsPortfolioSortDirection;
  limit: number;
  offset: number;
};

export type ProductsPortfolioResult = ProductsPortfolioListV1;
export type ProductsPortfolioItem = ProductsPortfolioItemV1;

export type QueryParamValue = string | number | boolean | null | undefined;

export type HttpClient = {
  getJson<T>(path: string, options?: { query?: Record<string, QueryParamValue> }): Promise<T>;
};

export type ProductsPortfolioApi = {
  listProductsPortfolio(query: ProductsPortfolioQuery): Promise<ProductsPortfolioResult>;
};

export function toProductsPortfolioQueryParams(
  query: ProductsPortfolioQuery,
): Record<string, QueryParamValue> {
  return {
    search: query.search.trim() || undefined,
    brand_name: query.brandName.trim() || undefined,
    taxonomy_leaf0_name: query.taxonomyLeaf0Name.trim() || undefined,
    status: query.status.trim() || undefined,
    sort_key: query.sortKey,
    sort_direction: query.sortDirection,
    limit: query.limit,
    offset: query.offset,
  };
}

export function createProductsPortfolioApi(client: HttpClient): ProductsPortfolioApi {
  return {
    listProductsPortfolio(query) {
      return client.getJson<ProductsPortfolioResult>("/api/v1/products/portfolio", {
        query: toProductsPortfolioQueryParams(query),
      });
    },
  };
}
