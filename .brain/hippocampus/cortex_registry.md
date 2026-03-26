---
id: hippocampus-cortex-registry
title: Cortex Region Registry
type: hippocampus
tags: [cortex, registry, regions]
updated_at: 2026-03-26
---

# Cortex Region Registry

Maps all cortex regions to their index files and lesson directories.

| Region | Index | Lessons | Description |
|--------|-------|---------|-------------|
| `backend` | `.brain/cortex/backend/index.md` | `.brain/cortex/backend/lessons/` | Go server_core modules, patterns, adapters |
| `frontend` | `.brain/cortex/frontend/index.md` | `.brain/cortex/frontend/lessons/` | React thin-client, feature packages, SDK boundary |
| `database` | `.brain/cortex/database/index.md` | `.brain/cortex/database/lessons/` | Postgres, tenant isolation, migrations, timeseries |
| `infra` | `.brain/cortex/infra/index.md` | `.brain/cortex/infra/lessons/` | Contracts, codegen, Docker/K8s, observability |

## Sinapse Store

Domain sinapses (rich context files) live in `.brain/sinapses/`. Naming: `<domain>-<topic>.md`.

## Lesson Distribution

- `.brain/cortex/<region>/lessons/` — domain-specific lessons
- `.brain/lessons/cross-domain/` — lessons spanning multiple regions
- `.brain/lessons/inbox/` — uncategorized pending triage
- `.brain/lessons/archived/` — promoted to conventions or superseded

## Adding a New Region

1. Create `.brain/cortex/<region>/` + `.brain/cortex/<region>/lessons/`
2. Create `.brain/cortex/<region>/index.md` with required frontmatter
3. Register here in this file
