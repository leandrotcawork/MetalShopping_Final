# Page Delivery Checklist

## Before writing code
- [ ] SDK regenerated after contract was finalized
- [ ] legacy page (if any) inspected for visual patterns to preserve
- [ ] widget extraction decision made (3+ rule applied)

## While implementing
- [ ] data loaded via `@metalshopping/platform-sdk` only
- [ ] no direct `fetch()` in page or component
- [ ] no manual DTO type files created
- [ ] generated types used for all API response shapes
- [ ] loading state present (spinner or skeleton)
- [ ] error state present (plain message is enough)

## Final review
- [ ] real data visible in the browser (no mock)
- [ ] `pnpm tsc --noEmit` passes
- [ ] no console errors in the browser
- [ ] page does not break other existing pages
- [ ] repeated visual components extracted to `packages/ui` if 3+ uses
- [ ] no new shared folders with ambiguous ownership created
