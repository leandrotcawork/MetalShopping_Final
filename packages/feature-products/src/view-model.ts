import type { ProductsPortfolioItem, ProductsPortfolioResult } from "./api";

export type ProductsPortfolioSummary = {
  totalProducts: number;
  pricedProducts: number;
  stockedProducts: number;
  averageVisiblePrice: string;
};

export function buildProductsPortfolioSummary(
  result: ProductsPortfolioResult | null,
): ProductsPortfolioSummary {
  if (result === null) {
    return {
      totalProducts: 0,
      pricedProducts: 0,
      stockedProducts: 0,
      averageVisiblePrice: "BRL 0.00",
    };
  }

  const pricedRows = result.rows.filter((row) => row.current_price_amount !== null);
  const stockedRows = result.rows.filter(
    (row) => row.on_hand_quantity !== null && row.on_hand_quantity !== undefined && row.on_hand_quantity > 0,
  );
  const averagePrice =
    pricedRows.length === 0
      ? 0
      : pricedRows.reduce((total, row) => total + (row.current_price_amount ?? 0), 0) /
        pricedRows.length;

  return {
    totalProducts: result.paging.total,
    pricedProducts: pricedRows.length,
    stockedProducts: stockedRows.length,
    averageVisiblePrice: formatCurrency(averagePrice),
  };
}

export function formatCurrency(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "—";
  }

  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(value);
}

export function formatQuantity(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "—";
  }

  return new Intl.NumberFormat("pt-BR", {
    maximumFractionDigits: 2,
  }).format(value);
}

export function rowHasLiveCommercialState(row: ProductsPortfolioItem) {
  return (
    (row.current_price_amount !== null && row.current_price_amount !== undefined) ||
    (row.on_hand_quantity !== null && row.on_hand_quantity !== undefined)
  );
}
