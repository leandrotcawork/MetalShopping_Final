param(
  [string]$DatabaseUrl = $env:MS_DATABASE_URL,
  [string]$TenantId = $env:MS_TENANT_ID,
  [ValidateSet("catalog", "xlsx")]
  [string]$InputMode = "catalog",
  [string]$CatalogProductIds = $env:MS_CATALOG_PRODUCT_IDS,
  [string]$PythonPath = $env:MS_PYTHON_PATH,
  [string]$ReportPath = "docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md",
  [string]$SuiteConfigPath = ""
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
if ($InputMode -eq "catalog" -and [string]::IsNullOrWhiteSpace($CatalogProductIds)) {
  throw "When InputMode=catalog, MS_CATALOG_PRODUCT_IDS (or -CatalogProductIds) is required."
}

$python = if ([string]::IsNullOrWhiteSpace($PythonPath)) { ".venv\Scripts\python.exe" } else { $PythonPath }

function New-DefaultSuiteConfig {
  return @(
    @{
      supplier_code = "DEXCO"
      strategy = "http.vtex_persisted_query.v1"
      expected_statuses = @("OK", "NOT_FOUND", "AMBIGUOUS", "ERROR")
      allow_empty = $true
    },
    @{
      supplier_code = "TELHA_NORTE"
      strategy = "http.vtex_persisted_query.v1"
      expected_statuses = @("OK", "NOT_FOUND", "AMBIGUOUS", "ERROR")
      allow_empty = $true
    },
    @{
      supplier_code = "CONDEC"
      strategy = "http.html_search.v1"
      expected_statuses = @("OK", "NOT_FOUND", "AMBIGUOUS", "ERROR")
      allow_empty = $true
    },
    @{
      supplier_code = "OBRA_FACIL"
      strategy = "http.mock.v1"
      expected_statuses = @("OK", "NOT_FOUND", "AMBIGUOUS", "ERROR")
      allow_empty = $true
      notes = "PLAYWRIGHT non-mock runtime validation remains under ADR-0034."
    }
  )
}

function Load-SuiteConfig {
  param([string]$Path)
  if ([string]::IsNullOrWhiteSpace($Path)) {
    return (New-DefaultSuiteConfig)
  }
  if (-not (Test-Path $Path)) {
    throw "SuiteConfigPath not found: $Path"
  }
  $raw = Get-Content $Path -Raw
  $parsed = $raw | ConvertFrom-Json
  if ($null -eq $parsed) {
    throw "Invalid suite config JSON: $Path"
  }
  return @($parsed)
}

function Evaluate-Result {
  param(
    [hashtable]$Config,
    [hashtable]$Summary
  )
  $allowed = @()
  if ($Config.ContainsKey("expected_statuses") -and $null -ne $Config.expected_statuses) {
    $allowed = @($Config.expected_statuses)
  }
  $allowEmpty = $false
  if ($Config.ContainsKey("allow_empty")) {
    $allowEmpty = [bool]$Config.allow_empty
  }

  if ($Summary.request_status -ne "completed") {
    return @{ ok = $false; reason = "request_status=$($Summary.request_status)" }
  }
  if ([string]::IsNullOrWhiteSpace($Summary.run_id)) {
    return @{ ok = $false; reason = "run_id missing" }
  }
  if ($Summary.total_items -eq 0 -and -not $allowEmpty) {
    return @{ ok = $false; reason = "total_items=0 and allow_empty=false" }
  }

  if ($allowed.Count -gt 0) {
    $statuses = @($Summary.status_counts | ForEach-Object { $_.item_status })
    $invalid = @($statuses | Where-Object { $_ -notin $allowed })
    if ($invalid.Count -gt 0) {
      return @{ ok = $false; reason = "unexpected_statuses=$($invalid -join ',')" }
    }
  }

  return @{ ok = $true; reason = "ok" }
}

function Build-Report {
  param(
    [string]$Path,
    [string]$Tenant,
    [string]$Mode,
    [string]$CatalogIds,
    [array]$Results
  )
  $now = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss K")
  $lines = @()
  $lines += "# Shopping Driver Suite Acceptance"
  $lines += ""
  $lines += "- Generated at: $now"
  $lines += "- Tenant: $Tenant"
  $lines += "- InputMode: $Mode"
  $lines += "- CatalogProductIds: $CatalogIds"
  $lines += ""
  $lines += "## Supplier Results"
  $lines += ""
  $lines += "| Supplier | Strategy | RunRequest | RunID | TotalItems | Outcome | Reason |"
  $lines += "|---|---|---|---|---:|---|---|"

  foreach ($row in $Results) {
    $outcome = if ($row.validation.ok) { "PASS" } else { "FAIL" }
    $lines += "| $($row.supplier_code) | $($row.strategy) | $($row.summary.run_request_id) | $($row.summary.run_id) | $($row.summary.total_items) | $outcome | $($row.validation.reason) |"
  }

  foreach ($row in $Results) {
    $lines += ""
    $lines += "### $($row.supplier_code)"
    $lines += ""
    if ($row.notes) {
      $lines += "- Notes: $($row.notes)"
    }
    $lines += "- RequestStatus: $($row.summary.request_status)"
    $lines += "- RunID: $($row.summary.run_id)"
    $lines += "- TotalItems: $($row.summary.total_items)"
    $lines += "- StatusCounts: $((($row.summary.status_counts | ConvertTo-Json -Compress) -replace '\|','/'))"
    $lines += "- ChannelStatusCounts: $((($row.summary.channel_status_counts | ConvertTo-Json -Compress) -replace '\|','/'))"
    $lines += "- Samples: $((($row.summary.samples | ConvertTo-Json -Compress) -replace '\|','/'))"
  }

  $dir = Split-Path -Parent $Path
  if (-not [string]::IsNullOrWhiteSpace($dir) -and -not (Test-Path $dir)) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
  }
  Set-Content -Path $Path -Value ($lines -join "`r`n") -Encoding UTF8
}

Write-Step "Loading suite config"
$suite = Load-SuiteConfig -Path $SuiteConfigPath

$env:MS_DATABASE_URL = $DatabaseUrl
$env:MS_TENANT_ID = $TenantId
$env:MS_INPUT_MODE = $InputMode
$env:MS_CATALOG_PRODUCT_IDS = $CatalogProductIds
if ([string]::IsNullOrWhiteSpace($env:GOCACHE)) {
  $env:GOCACHE = (Join-Path (Get-Location) ".gocache-verify")
}

$results = @()

foreach ($entryObj in $suite) {
  $entry = @{}
  if ($entryObj -is [System.Collections.IDictionary]) {
    foreach ($k in $entryObj.Keys) { $entry["$k"] = $entryObj[$k] }
  } else {
    foreach ($p in $entryObj.PSObject.Properties) { $entry[$p.Name] = $p.Value }
  }

  $supplierCode = ("" + $entry.supplier_code).Trim().ToUpper()
  if ([string]::IsNullOrWhiteSpace($supplierCode)) {
    continue
  }
  $strategy = ("" + $entry.strategy).Trim().ToLower()
  if ([string]::IsNullOrWhiteSpace($strategy)) {
    $strategy = "http.mock.v1"
  }

  Write-Step "Supplier $supplierCode ($strategy)"
  $env:MS_SMOKE_SUPPLIER_CODE = $supplierCode
  $env:MS_SMOKE_STRATEGY = $strategy

  if ($entry.ContainsKey("base_url")) { $env:MS_SMOKE_BASE_URL = ("" + $entry.base_url).Trim() }
  if ($entry.ContainsKey("operation_name")) { $env:MS_SMOKE_OPERATION_NAME = ("" + $entry.operation_name).Trim() }
  if ($entry.ContainsKey("sha256_hash")) { $env:MS_SMOKE_SHA256_HASH = ("" + $entry.sha256_hash).Trim() }
  if ($entry.ContainsKey("preferred_seller_name")) { $env:MS_SMOKE_PREFERRED_SELLER_NAME = ("" + $entry.preferred_seller_name).Trim() }
  if ($entry.ContainsKey("allow_fallback")) { $env:MS_SMOKE_ALLOW_FALLBACK = ("" + $entry.allow_fallback).Trim() }
  if ($entry.ContainsKey("require_available_offer")) { $env:MS_SMOKE_REQUIRE_AVAILABLE_OFFER = ("" + $entry.require_available_offer).Trim() }
  if ($entry.ContainsKey("search_url_template")) { $env:MS_SMOKE_SEARCH_URL_TEMPLATE = ("" + $entry.search_url_template).Trim() }
  if ($entry.ContainsKey("price_regex")) { $env:MS_SMOKE_PRICE_REGEX = ("" + $entry.price_regex).Trim() }
  if ($entry.ContainsKey("seller_regex")) { $env:MS_SMOKE_SELLER_REGEX = ("" + $entry.seller_regex).Trim() }

  $publishedRaw = go run .\apps\server_core\cmd\smoke-shopping-event
  $published = $publishedRaw | ConvertFrom-Json
  $runRequestId = ("" + $published.run_request_id).Trim()
  if ([string]::IsNullOrWhiteSpace($runRequestId)) {
    throw "Failed to parse run_request_id from smoke-shopping-event output for $supplierCode."
  }

  $env:MS_SHOPPING_WORKER_MODE = "event"
  $env:MS_SHOPPING_MAX_QUEUE_CLAIMS = "50"
  $env:MS_SHOPPING_XLSX_FALLBACK_LIMIT = "5"
  & $python .\apps\integration_worker\shopping_price_worker.py | Out-Host

  $summaryRaw = & $python .\scripts\smoke_shopping_driver_suite_report.py `
    --database-url "$DatabaseUrl" `
    --tenant-id "$TenantId" `
    --run-request-id "$runRequestId" `
    --supplier-code "$supplierCode"
  $summary = $summaryRaw | ConvertFrom-Json
  $summaryTable = @{}
  foreach ($p in $summary.PSObject.Properties) { $summaryTable[$p.Name] = $p.Value }

  $validation = Evaluate-Result -Config $entry -Summary $summaryTable
  $results += @{
    supplier_code = $supplierCode
    strategy = $strategy
    summary = $summaryTable
    validation = $validation
    notes = if ($entry.ContainsKey("notes")) { ("" + $entry.notes) } else { "" }
  }
}

Build-Report -Path $ReportPath -Tenant $TenantId -Mode $InputMode -CatalogIds $CatalogProductIds -Results $results

$fails = @($results | Where-Object { -not $_.validation.ok })
Write-Step "Driver suite report written: $ReportPath"
if ($fails.Count -gt 0) {
  throw "Driver suite completed with failures: $($fails.Count)"
}
Write-Step "Driver suite completed with all suppliers PASS"
