---
id: cortex-infra-index
title: Infrastructure Domain
region: cortex/infra
tags: [infra, deployment, devops, kubernetes]
links:
  - hippocampus/architecture
  - hippocampus/conventions
  - hippocampus/decisions_log
weight: 0.8
updated_at: 2026-03-24T10:00:00Z
---

# Infrastructure Domain

MetalShopping runs on Kubernetes with Docker container orchestration. Infrastructure as Code lives in `ops/`.

## Deployment Architecture

### Container Images

| Image | Source | Purpose |
|-------|--------|---------|
| `server-core:latest` | `apps/server_core/Dockerfile` | Go backend, HTTP API, tenant logic |
| `analytics-worker:latest` | `apps/analytics_worker/Dockerfile` | Python async compute, event processing |
| `web:latest` | `apps/web/Dockerfile` | React frontend (SPA), static assets |

### Kubernetes Manifests

All Kubernetes resources live in `ops/k8s/`:

```
ops/k8s/
  ├── overlays/
  │   ├── dev/        Development cluster config (1 replica, dev secrets)
  │   ├── staging/    Staging cluster (2 replicas, staging DB)
  │   └── prod/       Production (HA, monitoring, alerting)
  ├── base/
  │   ├── namespace.yaml
  │   ├── deployments/   server-core, analytics-worker, web
  │   ├── services/      LoadBalancer, ClusterIP
  │   ├── configmaps/    App config, feature flags
  │   ├── secrets/       DB credentials, API keys (injected)
  │   └── ingress.yaml   Route HTTP traffic
  └── kustomization.yaml
```

**Deployment tooling:** Kustomize + Helm (selective use for chart complexity).

## Environment Configuration

### Environment Variables

All config is environment-specific:

| Env | DB | Redis | Feature Flags | Replicas |
|-----|----|----|---|----------|
| `dev` | localhost:5432 | localhost:6379 | All enabled | 1 |
| `staging` | staging-db.aws | staging-redis.aws | Feature-gated | 2 |
| `prod` | prod-db.aws | prod-redis.aws | Governed | 3–5 (auto-scale) |

Config is injected via:
- **ConfigMap** — non-sensitive config (feature flags, thresholds)
- **Secrets** — sensitive (DB password, JWT key)
- **Environment** — pod-level overrides

### Feature Flags & Governance

Feature flags control feature rollouts:

```yaml
# ops/governance/feature-flags.yaml
- id: new_checkout_flow
  enabled: false
  rollout_percentage: 0        # Dev: test in isolation
  # staging: 50               # Staging: canary
  # prod: 100                 # Prod: full rollout
```

Flag resolution happens at runtime in `internal/platform/governance/`.

## Observability

### Logging

Structured logs are shipped to centralized logging (ELK/Datadog):

```json
{
  "timestamp": "2026-03-24T10:00:00Z",
  "level": "info",
  "trace_id": "abc123xyz",
  "service": "server-core",
  "action": "create_product",
  "result": "success",
  "duration_ms": 42,
  "tenant_id": "tenant-123",
  "user_id": "user-456"
}
```

### Metrics

Prometheus metrics are scraped from `/metrics` endpoint:

```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="POST", endpoint="/products", status="201"} 1234

# HELP http_request_duration_seconds HTTP request latency
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1", endpoint="/products"} 1100
http_request_duration_seconds_bucket{le="0.5", endpoint="/products"} 1200
http_request_duration_seconds_bucket{le="1.0", endpoint="/products"} 1234
```

### Tracing

Distributed tracing (OpenTelemetry) tracks requests across services:
- Every request gets a `trace_id`
- Go handler, Python worker, database all log the same `trace_id`
- Tracing backend (Jaeger/Tempo) shows full request waterfall

## CI/CD

### GitHub Actions

Pipeline defined in `.github/workflows/`:

1. **Test** — npm test, go test, contract validation
2. **Build** — Docker images for all services
3. **Push** — Images to registry (AWS ECR, Docker Hub)
4. **Deploy** — kubectl apply for staging/prod (gated by approval)

```yaml
# .github/workflows/deploy.yaml
on:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run web:test
      - run: go test ./apps/server_core/...
      - run: ./scripts/validate_contracts.ps1

  build-and-deploy:
    needs: test
    if: success()
    steps:
      - run: docker build -t server-core:${{ github.sha }} apps/server_core/
      - run: aws ecr push server-core:${{ github.sha }}
      - run: kubectl set image deployment/server-core ...
```

## Database Backups & Recovery

### Backup Strategy

- **Frequency:** Daily snapshots (AWS RDS automated backups)
- **Retention:** 30 days
- **Recovery:** Point-in-time restore available (< 5 minute RTO)

Test recovery procedures quarterly.

### Migration Safety

Before running migrations in prod:
1. Full backup
2. Test migration on staging replica
3. Dry-run on prod (transaction not committed)
4. Execute with rollback plan in place
5. Monitor logs for errors

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.8
