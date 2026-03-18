import { describe, expect, it } from "vitest";

import { defaultProductsPortfolioQuery, toggleProductsPortfolioSort } from "./query-state";

describe("toggleProductsPortfolioSort", () => {
  it("switches to ascending when changing the sort key", () => {
    const next = toggleProductsPortfolioSort(defaultProductsPortfolioQuery, "current_price_amount");
    expect(next.sort_key).toBe("current_price_amount");
    expect(next.sort_direction).toBe("asc");
    expect(next.offset).toBe(0);
  });

  it("toggles the direction when the same key is clicked again", () => {
    const desc = toggleProductsPortfolioSort(defaultProductsPortfolioQuery, "pn_interno");
    expect(desc.sort_direction).toBe("desc");

    const asc = toggleProductsPortfolioSort(desc, "pn_interno");
    expect(asc.sort_direction).toBe("asc");
  });
});
