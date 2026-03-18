import type {
  ListProductsPortfolioQueryParams,
  ProductsPortfolioSortDirection,
  ProductsPortfolioSortKey,
  ServerCoreSdk,
} from "@metalshopping/generated-sdk";
import type {
  ProductsPortfolioItemV1,
  ProductsPortfolioListV1,
} from "@metalshopping/generated-types";

export type ProductsPortfolioQuery = Required<ListProductsPortfolioQueryParams>;
export type ProductsPortfolioResult = ProductsPortfolioListV1;
export type ProductsPortfolioItem = ProductsPortfolioItemV1;
export type ProductsPortfolioApi = Pick<ServerCoreSdk["products"], "listProductsPortfolio">;
export type { ProductsPortfolioSortDirection, ProductsPortfolioSortKey };
