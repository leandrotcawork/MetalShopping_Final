# Login Visual Baseline

## Purpose

Freeze the shared visual baseline for the MetalShopping login experience across:

- `apps/web` fallback login surface
- Keycloak login theme

## Rules

- Keycloak remains the final credential entry surface
- the React login remains a launch/fallback surface only
- both surfaces must preserve the same brand direction
- assets, copy hierarchy, and spacing rhythm must stay aligned

## Shared baseline

- brand eyebrow: `Metal Nobre Acabamentos`
- hero title:
  - `Precificacao`
  - `inteligente.`
- primary copy:
  - secure session
  - backend-owned auth/session
  - no app token storage in the browser
- primary CTA:
  - `Entrar com identidade segura`

## Visual tokens to preserve

- `Inter` typography
- wine primary axis centered on `#91132a`
- soft neutral background with light gradient treatment
- compact desktop login without unnecessary scroll
- lightly rounded controls and cards
- stronger branded CTA than secondary actions

## Drift guardrails

- do not redesign Keycloak independently from the React fallback
- do not add copy blocks in one surface without deciding whether they belong in both
- do not let the fallback React login become the credential form
- when the login visual language changes, update both surfaces in the same tranche
