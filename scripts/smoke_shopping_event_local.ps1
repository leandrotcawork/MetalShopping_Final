param(
  [string]$DatabaseUrl = $env:MS_DATABASE_URL,
  [string]$TenantId = $env:MS_TENANT_ID,
  [ValidateSet("catalog", "xlsx")]
  [string]$InputMode = "xlsx",
  [string]$SupplierCode = $env:MS_SMOKE_SUPPLIER_CODE,
  [string]$CatalogProductIds = $env:MS_CATALOG_PRODUCT_IDS,
  [string]$SmokeStrategy = $env:MS_SMOKE_STRATEGY,
  [string]$SmokeBaseUrl = $env:MS_SMOKE_BASE_URL,
  [string]$SmokeOperationName = $env:MS_SMOKE_OPERATION_NAME,
  [string]$SmokeSha256Hash = $env:MS_SMOKE_SHA256_HASH,
  [string]$SmokePreferredSellerName = $env:MS_SMOKE_PREFERRED_SELLER_NAME,
  [string]$SmokeAllowFallback = $env:MS_SMOKE_ALLOW_FALLBACK,
  [string]$SmokeRequireAvailableOffer = $env:MS_SMOKE_REQUIRE_AVAILABLE_OFFER,
  [string]$SmokeStartUrl = $env:MS_SMOKE_START_URL,
  [string]$SmokeSearchUrl = $env:MS_SMOKE_SEARCH_URL,
  [string]$SmokeWaitUntil = $env:MS_SMOKE_WAIT_UNTIL,
  [string]$SmokeHeadless = $env:MS_SMOKE_HEADLESS,
  [string]$SmokeFallbackSearchEnabled = $env:MS_SMOKE_FALLBACK_SEARCH_ENABLED,
  [string]$SmokePdpSelectorsJson = $env:MS_SMOKE_PDP_SELECTORS_JSON,
  [string]$SmokePdpPriceSelector = $env:MS_SMOKE_PDP_PRICE_SELECTOR,
  [string]$SmokePdpSellerSelector = $env:MS_SMOKE_PDP_SELLER_SELECTOR,
  [string]$SmokePdpChannelSelector = $env:MS_SMOKE_PDP_CHANNEL_SELECTOR,
  [string]$SmokeMaxRetries = $env:MS_SMOKE_MAX_RETRIES,
  [string]$PythonPath = $env:MS_PYTHON_PATH
)

$ErrorActionPreference = "Stop"

function Write-Step {
  param([string]$Message)
  Write-Host "==> $Message" -ForegroundColor Cyan
}

if ([string]::IsNullOrWhiteSpace($DatabaseUrl)) {
  throw "Missing DatabaseUrl (pass -DatabaseUrl or set MS_DATABASE_URL)."
}
if ([string]::IsNullOrWhiteSpace($TenantId)) {
  throw "Missing TenantId (pass -TenantId or set MS_TENANT_ID)."
}

Write-Step "Shopping event smoke (ADR-0025 Phase 2)"
Write-Host "Tenant=$TenantId InputMode=$InputMode Supplier=$SupplierCode" -ForegroundColor Yellow

$env:MS_DATABASE_URL = $DatabaseUrl
$env:MS_TENANT_ID = $TenantId
$env:MS_INPUT_MODE = $InputMode
$env:MS_SMOKE_SUPPLIER_CODE = $SupplierCode
$env:MS_CATALOG_PRODUCT_IDS = $CatalogProductIds
$env:MS_SMOKE_STRATEGY = $SmokeStrategy
$env:MS_SMOKE_BASE_URL = $SmokeBaseUrl
$env:MS_SMOKE_OPERATION_NAME = $SmokeOperationName
$env:MS_SMOKE_SHA256_HASH = $SmokeSha256Hash
$env:MS_SMOKE_PREFERRED_SELLER_NAME = $SmokePreferredSellerName
$env:MS_SMOKE_ALLOW_FALLBACK = $SmokeAllowFallback
$env:MS_SMOKE_REQUIRE_AVAILABLE_OFFER = $SmokeRequireAvailableOffer
$env:MS_SMOKE_START_URL = $SmokeStartUrl
$env:MS_SMOKE_SEARCH_URL = $SmokeSearchUrl
$env:MS_SMOKE_WAIT_UNTIL = $SmokeWaitUntil
$env:MS_SMOKE_HEADLESS = $SmokeHeadless
$env:MS_SMOKE_FALLBACK_SEARCH_ENABLED = $SmokeFallbackSearchEnabled
$env:MS_SMOKE_PDP_SELECTORS_JSON = $SmokePdpSelectorsJson
$env:MS_SMOKE_PDP_PRICE_SELECTOR = $SmokePdpPriceSelector
$env:MS_SMOKE_PDP_SELLER_SELECTOR = $SmokePdpSellerSelector
$env:MS_SMOKE_PDP_CHANNEL_SELECTOR = $SmokePdpChannelSelector
$env:MS_SMOKE_MAX_RETRIES = $SmokeMaxRetries
if ([string]::IsNullOrWhiteSpace($env:GOCACHE)) {
  $env:GOCACHE = (Join-Path (Get-Location) ".gocache-verify")
}

Write-Step "Step 1: Insert run_request + published outbox event + seed supplier/manifest"
$output = go run .\apps\server_core\cmd\smoke-shopping-event
Write-Host $output -ForegroundColor Green

Write-Step "Step 2: Run worker in event mode (single claim)"
$env:MS_SHOPPING_WORKER_MODE = "event"
$env:MS_SHOPPING_MAX_QUEUE_CLAIMS = "1"
$env:MS_SHOPPING_XLSX_FALLBACK_LIMIT = "5"

$python = if ([string]::IsNullOrWhiteSpace($PythonPath)) { "python" } else { $PythonPath }
& $python .\apps\integration_worker\shopping_price_worker.py

Write-Step "Smoke completed"
