import { describe, expect, it } from "vitest";

import { defaultProductsPortfolioQuery, toggleProductsPortfolioSort } from "./query-state";

describe("toggleProductsPortfolioSort", () => {
  it("switches to ascending when changing the sort key", () => {
    const next = toggleProductsPortfolioSort(defaultProductsPortfolioQuery, "current_price_amount");
    expect(next.sortKey).toBe("current_price_amount");
    expect(next.sortDirection).toBe("asc");
    expect(next.offset).toBe(0);
  });

  it("toggles the direction when the same key is clicked again", () => {
    const desc = toggleProductsPortfolioSort(defaultProductsPortfolioQuery, "pn_interno");
    expect(desc.sortDirection).toBe("desc");

    const asc = toggleProductsPortfolioSort(desc, "pn_interno");
    expect(asc.sortDirection).toBe("asc");
  });
});
