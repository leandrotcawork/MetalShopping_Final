param(
    [string]$Root = "."
)

$ErrorActionPreference = "Stop"

function Ensure-Dir {
    param([string]$Path)
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

$serverModules = @(
    "iam",
    "tenant_admin",
    "catalog",
    "inventory",
    "pricing",
    "sales",
    "suppliers",
    "customers",
    "procurement",
    "market_intelligence",
    "analytics_serving",
    "crm",
    "automation",
    "integrations_control",
    "alerts"
)

$moduleLayout = @(
    "domain",
    "application",
    "ports",
    "adapters",
    "transport",
    "events",
    "readmodel"
)

$paths = @(
    "apps/server_core/cmd/metalshopping-server",
    "apps/server_core/internal/platform/auth",
    "apps/server_core/internal/platform/tenancy_runtime",
    "apps/server_core/internal/platform/runtime_config",
    "apps/server_core/internal/platform/governance/config_registry",
    "apps/server_core/internal/platform/governance/policy_resolver",
    "apps/server_core/internal/platform/governance/threshold_resolver",
    "apps/server_core/internal/platform/governance/feature_flags",
    "apps/server_core/internal/platform/db/postgres",
    "apps/server_core/internal/platform/db/timeseries",
    "apps/server_core/internal/platform/messaging/outbox",
    "apps/server_core/internal/platform/messaging/inbox",
    "apps/server_core/internal/platform/jobs",
    "apps/server_core/internal/platform/cache",
    "apps/server_core/internal/platform/files",
    "apps/server_core/internal/platform/delivery/email",
    "apps/server_core/internal/platform/delivery/sms",
    "apps/server_core/internal/platform/delivery/whatsapp",
    "apps/server_core/internal/platform/delivery/webhooks",
    "apps/server_core/internal/platform/observability",
    "apps/server_core/internal/platform/security",
    "apps/server_core/internal/platform/auditlog",
    "apps/server_core/internal/shared/errors",
    "apps/server_core/internal/shared/ids",
    "apps/server_core/internal/shared/clock",
    "apps/server_core/internal/shared/money",
    "apps/server_core/internal/shared/pagination",
    "apps/server_core/internal/shared/httpx",
    "apps/server_core/migrations",
    "apps/server_core/tests/unit",
    "apps/server_core/tests/contract",
    "apps/server_core/tests/integration",
    "apps/analytics_worker/src/inputs",
    "apps/analytics_worker/src/jobs",
    "apps/analytics_worker/src/compute",
    "apps/analytics_worker/src/rules",
    "apps/analytics_worker/src/scoring",
    "apps/analytics_worker/src/projections",
    "apps/analytics_worker/src/explainability",
    "apps/analytics_worker/src/publish",
    "apps/analytics_worker/tests",
    "apps/integration_worker/src/connectors/erp",
    "apps/integration_worker/src/connectors/marketplaces",
    "apps/integration_worker/src/connectors/market_intelligence",
    "apps/integration_worker/src/connectors/imports",
    "apps/integration_worker/src/connectors/exports",
    "apps/integration_worker/src/crawlers",
    "apps/integration_worker/src/normalizers",
    "apps/integration_worker/src/publish",
    "apps/integration_worker/tests",
    "apps/automation_worker/src/orchestration",
    "apps/automation_worker/src/triggers",
    "apps/automation_worker/src/actions",
    "apps/automation_worker/src/channels",
    "apps/automation_worker/src/campaigns",
    "apps/automation_worker/src/publish",
    "apps/automation_worker/tests",
    "apps/notifications_worker/src/routing",
    "apps/notifications_worker/src/templates",
    "apps/notifications_worker/src/channels/email",
    "apps/notifications_worker/src/channels/sms",
    "apps/notifications_worker/src/channels/whatsapp",
    "apps/notifications_worker/src/channels/webhook",
    "apps/notifications_worker/src/publish",
    "apps/notifications_worker/tests",
    "apps/web",
    "apps/desktop",
    "apps/admin_console",
    "contracts/api/openapi",
    "contracts/api/jsonschema",
    "contracts/events/v1",
    "contracts/events/v2",
    "contracts/governance/policies",
    "contracts/governance/thresholds",
    "contracts/governance/feature_flags",
    "bootstrap/seeds/iam",
    "bootstrap/seeds/tenants",
    "bootstrap/seeds/suppliers",
    "bootstrap/seeds/governance/policies",
    "bootstrap/seeds/governance/thresholds",
    "bootstrap/seeds/governance/feature_flags",
    "packages/ui",
    "packages/generated/sdk_ts",
    "packages/generated/sdk_py",
    "packages/generated/types_ts",
    "ops/docker",
    "ops/k8s",
    "ops/observability",
    "ops/runbooks",
    "ops/provisioning",
    "docs",
    "scripts"
)

Push-Location $Root
try {
    foreach ($path in $paths) {
        Ensure-Dir -Path $path
    }

    foreach ($module in $serverModules) {
        foreach ($segment in $moduleLayout) {
            Ensure-Dir -Path ("apps/server_core/internal/modules/{0}/{1}" -f $module, $segment)
        }
    }
}
finally {
    Pop-Location
}

Write-Host "Architecture scaffold ensured."
