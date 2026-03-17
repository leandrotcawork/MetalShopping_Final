# Module Checklist

## Before editing

- confirm the capability belongs in `modules/`
- confirm the bounded context name
- confirm canonical ownership
- confirm whether supporting contracts already exist

## While editing

- keep the full module structure
- keep `domain/` free of transport and infrastructure
- keep `application/` orchestration-focused
- keep `ports/` explicit
- keep `adapters/` concrete and bounded
- keep `transport/` interface-only
- keep `events/` aligned to contract-driven async behavior
- keep `readmodel/` consumption-focused

## Final review

- module fits `docs/MODULE_STANDARDS.md`
- boundary decisions fit `docs/PLATFORM_BOUNDARIES.md`
- event and readmodel usage fits `docs/READMODEL_AND_EVENTS_RULES.md`
- no hidden drift toward `platform/` or `shared/`
- any missing contracts are called out explicitly

