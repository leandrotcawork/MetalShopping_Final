import { useMemo } from "react";

import type { AnalyticsProductsOverviewV1Dto } from "../../../../legacy_products_dto";
import { ProductsOverview } from "./overview/ProductsOverview";
import { mapProductsOverviewViewModel } from "./overview/products_overview.viewmodel";

type ProductsAnalyticsViewProps = {
  overview: AnalyticsProductsOverviewV1Dto;
};

export function ProductsAnalyticsView({ overview }: ProductsAnalyticsViewProps) {
  const model = useMemo(() => mapProductsOverviewViewModel(overview), [overview]);
  return <ProductsOverview model={model} />;
}
