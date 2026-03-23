import type {
  ListProductsPortfolioQueryParams,
  ProductsPortfolioSortDirection,
  ProductsPortfolioSortKey,
  ServerCoreSdk,
} from "@metalshopping/sdk-runtime";
import type {
  ProductsPortfolioItemV1,
  ProductsPortfolioListV1,
} from "@metalshopping/sdk-types";

export type ProductsPortfolioQuery = Required<ListProductsPortfolioQueryParams>;
export type ProductsPortfolioResult = ProductsPortfolioListV1;
export type ProductsPortfolioItem = ProductsPortfolioItemV1;
export type ProductsPortfolioApi = Pick<ServerCoreSdk["products"], "listProductsPortfolio">;
export type ProductsShoppingApi = Pick<
  ServerCoreSdk["shopping"],
  "getBootstrap" | "listRuns" | "exportMarketReportXlsx"
>;
export type { ProductsPortfolioSortDirection, ProductsPortfolioSortKey };
