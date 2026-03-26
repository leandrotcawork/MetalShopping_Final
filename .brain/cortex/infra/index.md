---
id: cortex-infra-index
title: Infra Cortex Index
region: infra
type: cortex-index
tags: [infra, contracts, codegen, docker, k8s, keycloak, observability]
updated_at: 2026-03-26
---

# Infra Cortex

## Scope

Contracts layer, codegen pipeline, ops/ (Docker/K8s/Keycloak), observability, CI/CD.

## Contracts Layer

Source of truth for the entire platform API surface:

| Path | Content |
|------|---------|
| `contracts/api/openapi/` | OpenAPI 3.0 specs — one per domain |
| `contracts/api/jsonschema/` | JSON Schema payloads |
| `contracts/events/` | Versioned event schemas |
| `contracts/governance/` | Feature flags, policies, thresholds schemas |

**Rule**: contracts are hand-authored. Never derive contracts from implementation code.

## Codegen Pipeline

```powershell
# Generate SDK and types from contracts (requires Docker)
./scripts/generate_contract_artifacts.ps1 -Target all

# Validate all contracts
./scripts/validate_contracts.ps1 -Scope all
```

Generated artifacts go into `packages/` subdirectories — never edit these manually. CI rejects manual edits unless PR is labeled `codegen`.

## Ops Stack

| Path | Purpose |
|------|---------|
| `ops/docker/` | Docker compose files for local dev |
| `ops/k8s/` | Kubernetes manifests |
| `ops/keycloak/` | Keycloak realm config, OIDC setup |
| `ops/observability/` | Prometheus, Grafana, tracing config |
| `ops/provisioning/` | Tenant provisioning scripts |
| `ops/runbooks/` | Incident runbooks |

## Local Development

```bash
# Start full stack (PowerShell)
npm run dev:local
# Or directly:
powershell -ExecutionPolicy Bypass -File ./scripts/start_metalshopping_local.ps1
```

## Event System

- Outbox pattern: `outbox.AppendInTx` inside the same transaction as the write
- Events published after successful commit
- Schema-versioned events in `contracts/events/`
- Event consumers never call back into `server_core` HTTP

## Sinapses in This Region

_Add links to `.brain/sinapses/<infra-topic>.md` files as they are created._
