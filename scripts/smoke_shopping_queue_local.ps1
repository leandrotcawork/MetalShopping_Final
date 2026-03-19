param(
  [string]$DatabaseUrl = $env:MS_DATABASE_URL,
  [string]$TenantId = $env:MS_TENANT_ID,
  [ValidateSet("catalog", "xlsx")]
  [string]$InputMode = "catalog",
  [string]$ProductIds = $env:MS_PRODUCT_IDS,
  [string]$XlsxFilePath = $env:MS_XLSX_FILE_PATH
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

Write-Step "Shopping queue smoke (ADR-0018 Phase 1)"
Write-Host "Tenant=$TenantId InputMode=$InputMode" -ForegroundColor Yellow

$env:MS_DATABASE_URL = $DatabaseUrl
$env:MS_TENANT_ID = $TenantId
$env:MS_INPUT_MODE = $InputMode
$env:MS_PRODUCT_IDS = $ProductIds
$env:MS_XLSX_FILE_PATH = $XlsxFilePath

Write-Step "Step 1: Insert run_request, claim, run, complete (DB queue simulation tool)"
$output = go run .\apps\server_core\cmd\smoke-shopping-queue
Write-Host $output -ForegroundColor Green

Write-Step "Smoke completed"

