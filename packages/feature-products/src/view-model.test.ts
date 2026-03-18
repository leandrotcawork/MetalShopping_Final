import { describe, expect, it } from "vitest";

import type { ProductsPortfolioResult } from "./api";
import { buildProductsPortfolioSummary, formatCurrency, formatQuantity } from "./view-model";

const baseResult: ProductsPortfolioResult = {
  rows: [
    {
      product_id: "prd_1",
      sku: "SKU-1",
      name: "Produto 1",
      product_status: "active",
      current_price_amount: 100,
      replacement_cost_amount: 70,
      average_cost_amount: 65,
      on_hand_quantity: 8,
      updated_at: "2026-03-18T12:00:00Z",
    },
    {
      product_id: "prd_2",
      sku: "SKU-2",
      name: "Produto 2",
      product_status: "inactive",
      current_price_amount: null,
      replacement_cost_amount: null,
      average_cost_amount: null,
      on_hand_quantity: 0,
      updated_at: "2026-03-18T12:00:00Z",
    },
  ],
  filters: {
    brands: [],
    taxonomy_leaf0_names: [],
    taxonomy_leaf0_label: "Taxonomia",
    status: [],
  },
  paging: {
    offset: 0,
    limit: 50,
    returned: 2,
    total: 2,
  },
};

describe("buildProductsPortfolioSummary", () => {
  it("summarizes visible priced and stocked products", () => {
    const summary = buildProductsPortfolioSummary(baseResult);
    expect(summary.totalProducts).toBe(2);
    expect(summary.pricedProducts).toBe(1);
    expect(summary.stockedProducts).toBe(1);
    expect(summary.averageVisiblePrice).toContain("100");
  });
});

describe("formatters", () => {
  it("formats currency and quantity safely", () => {
    expect(formatCurrency(125.5)).toContain("125");
    expect(formatCurrency(null)).toBe("—");
    expect(formatQuantity(12.5)).toContain("12");
    expect(formatQuantity(undefined)).toBe("—");
  });
});
