import type {
  CommonErrorV1,
  ProductsPortfolioItemV1,
  ProductsPortfolioListV1,
} from "@metalshopping/generated-types";

export type ProductsPortfolioQuery = {
  search: string;
  brandName: string;
  taxonomyLeaf0Name: string;
  status: string;
  limit: number;
  offset: number;
};

export type ProductsPortfolioResult = ProductsPortfolioListV1;
export type ProductsPortfolioItem = ProductsPortfolioItemV1;

const defaultApiBaseUrl = "http://127.0.0.1:8080";
const defaultBearerToken = "local-dev-token";

function apiBaseUrl() {
  return (import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl).replace(/\/$/, "");
}

function apiBearerToken() {
  return import.meta.env.VITE_API_BEARER_TOKEN ?? defaultBearerToken;
}

function toQueryString(query: ProductsPortfolioQuery) {
  const params = new URLSearchParams();
  if (query.search.trim() !== "") {
    params.set("search", query.search.trim());
  }
  if (query.brandName.trim() !== "") {
    params.set("brand_name", query.brandName.trim());
  }
  if (query.taxonomyLeaf0Name.trim() !== "") {
    params.set("taxonomy_leaf0_name", query.taxonomyLeaf0Name.trim());
  }
  if (query.status.trim() !== "") {
    params.set("status", query.status.trim());
  }
  params.set("limit", String(query.limit));
  params.set("offset", String(query.offset));
  return params.toString();
}

export async function listProductsPortfolio(
  query: ProductsPortfolioQuery,
): Promise<ProductsPortfolioResult> {
  const response = await fetch(
    `${apiBaseUrl()}/api/v1/products/portfolio?${toQueryString(query)}`,
    {
      headers: {
        Authorization: `Bearer ${apiBearerToken()}`,
      },
    },
  );

  if (!response.ok) {
    const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;
    const message =
      errorPayload?.error?.message ??
      `Products portfolio request failed with status ${response.status}`;
    throw new Error(message);
  }

  return (await response.json()) as ProductsPortfolioResult;
}
