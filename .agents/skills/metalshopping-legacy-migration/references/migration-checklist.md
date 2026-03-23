# Legacy Migration Checklist

## Trigger phrases

Use this workflow when the user says:
- copy legacy
- leave identical
- visual first
- same HTML/CSS as legacy
- use mocks first, adapt later

## Phase checklist

### 1. Inventory

- Locate legacy entry TSX/page component
- Locate legacy CSS modules/global CSS
- Locate tab shell/header component
- Identify required providers/session hooks
- Identify helper registries and text maps
- Identify payload keys read by the legacy viewmodel

### 2. Baseline freeze

- List must-match regions
- List allowed temporary mocks
- List deferred integrations
- Record exact route and tab names

### 3. Runnable copy

- Copy markup hierarchy before rewriting
- Keep class names or semantic equivalents stable
- Add shims for missing providers/helpers
- Add mock DTO matching legacy keys exactly

### 4. Visual parity

- Header strip vs card container
- Brand stack/title/subtitle
- Tab rail background and active pill
- Right-side controls/status/theme
- Card radius/shadow/border
- Spacing rhythm between shell and first fold
- Typography weights and sizes
- Chip/badge colors and borders

### 5. Common failure modes

- Blank route: runtime exception or component receives no DTO
- Flat UI: missing CSS variables/tokens in scope
- Empty cards: mock shape mismatches viewmodel keys
- Wrong shell: copied inner widgets without top shell structure
- Hidden content: `opacity: 0` on base class instead of keyframe `from`
- Stretched tabs: rail placed in `1fr` slot without width constraint

### 6. Only after sign-off

- Add contracts
- Implement read endpoints
- Regenerate SDK
- Replace mocks block-by-block
