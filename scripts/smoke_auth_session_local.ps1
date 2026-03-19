param(
  [string]$BackendBaseUrl = "http://127.0.0.1:8080",
  [string]$KeycloakBaseUrl = "http://127.0.0.1:18081",
  [string]$WebOrigin = "http://127.0.0.1:5173",
  [string]$Username = "ms_admin",
  [string]$Password = "ChangeMe123!",
  [switch]$SkipInvalidCredentialsCheck
)

$ErrorActionPreference = "Stop"

function Write-Step {
  param([string]$Message)
  Write-Host "==> $Message" -ForegroundColor Cyan
}

function Assert-Status {
  param(
    [string]$Label,
    [int]$Actual,
    [int]$Expected
  )
  if ($Actual -ne $Expected) {
    throw "$Label expected HTTP $Expected, got $Actual."
  }
  Write-Host "$Label -> HTTP $Actual" -ForegroundColor Green
}

function Get-HeaderValue {
  param(
    [string]$HeadersFile,
    [string]$HeaderName
  )

  $pattern = "^{0}:\s*(.+)$" -f [Regex]::Escape($HeaderName)
  foreach ($line in Get-Content -Path $HeadersFile) {
    $match = [Regex]::Match($line, $pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
    if ($match.Success) {
      return $match.Groups[1].Value.Trim()
    }
  }
  return ""
}

function Resolve-FormActionUrl {
  param(
    [string]$HtmlPath,
    [string]$KeycloakBaseUrl
  )

  $html = Get-Content -Path $HtmlPath -Raw
  $match = [Regex]::Match($html, '<form[^>]*id="kc-form-login"[^>]*action="([^"]+)"', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
  if (-not $match.Success) {
    throw "Could not locate Keycloak login form action in $HtmlPath."
  }

  $action = [System.Net.WebUtility]::HtmlDecode($match.Groups[1].Value)
  if ($action.StartsWith("http://") -or $action.StartsWith("https://")) {
    return $action
  }

  if ($action.StartsWith("/")) {
    return "$KeycloakBaseUrl$action"
  }

  return "$KeycloakBaseUrl/$action"
}

function Get-CookieValue {
  param(
    [string]$CookieJarPath,
    [string]$CookieName
  )

  foreach ($line in Get-Content -Path $CookieJarPath) {
    if ($line.StartsWith("#") -or [string]::IsNullOrWhiteSpace($line)) {
      continue
    }

    $parts = $line -split "\t"
    if ($parts.Count -lt 7) {
      continue
    }

    if ($parts[5] -eq $CookieName) {
      return $parts[6]
    }
  }

  return ""
}

$tmpDir = Join-Path $env:TEMP ("ms_auth_smoke_" + [Guid]::NewGuid().ToString())
New-Item -Path $tmpDir -ItemType Directory | Out-Null

$cookieJar = Join-Path $tmpDir "cookies.txt"
$headersLogin = Join-Path $tmpDir "headers_login.txt"
$headersKeycloak = Join-Path $tmpDir "headers_keycloak.txt"
$headersSubmit = Join-Path $tmpDir "headers_submit.txt"
$headersMe = Join-Path $tmpDir "headers_me.txt"
$headersRefreshNoCsrf = Join-Path $tmpDir "headers_refresh_no_csrf.txt"
$headersRefreshBadCsrf = Join-Path $tmpDir "headers_refresh_bad_csrf.txt"
$headersRefresh = Join-Path $tmpDir "headers_refresh.txt"
$headersLogout = Join-Path $tmpDir "headers_logout.txt"
$headersMeAfterLogout = Join-Path $tmpDir "headers_me_after_logout.txt"
$headersLoginInvalid = Join-Path $tmpDir "headers_login_invalid.txt"
$headersKeycloakInvalid = Join-Path $tmpDir "headers_keycloak_invalid.txt"
$headersSubmitInvalid = Join-Path $tmpDir "headers_submit_invalid.txt"
$headersMeAfterInvalid = Join-Path $tmpDir "headers_me_after_invalid.txt"

$loginHtml = Join-Path $tmpDir "keycloak_login.html"
$submitHtml = Join-Path $tmpDir "submit_login.html"
$meJson = Join-Path $tmpDir "me.json"
$refreshNoCsrfJson = Join-Path $tmpDir "refresh_no_csrf.json"
$refreshBadCsrfJson = Join-Path $tmpDir "refresh_bad_csrf.json"
$refreshJson = Join-Path $tmpDir "refresh.json"
$logoutJson = Join-Path $tmpDir "logout.json"
$meAfterLogoutJson = Join-Path $tmpDir "me_after_logout.json"
$invalidLoginHtml = Join-Path $tmpDir "keycloak_invalid_login.html"
$invalidSubmitHtml = Join-Path $tmpDir "submit_invalid_login.html"
$meAfterInvalidJson = Join-Path $tmpDir "me_after_invalid.json"

Write-Step "Starting auth/session smoke at $tmpDir"

$loginUrl = "$BackendBaseUrl/api/v1/auth/session/login?return_to=%2Fproducts"
Write-Step "Step 1: Start login flow through backend"
$loginStatus = curl.exe -sS -o NUL -D $headersLogin -c $cookieJar -b $cookieJar -w "%{http_code}" $loginUrl
Assert-Status -Label "GET /auth/session/login" -Actual ([int]$loginStatus) -Expected 302

$redirectToKeycloak = Get-HeaderValue -HeadersFile $headersLogin -HeaderName "Location"
if ([string]::IsNullOrWhiteSpace($redirectToKeycloak)) {
  throw "Missing Location header on login redirect."
}
Write-Host "Redirect URL captured." -ForegroundColor Green

Write-Step "Step 2: Fetch Keycloak login page"
$keycloakPageStatus = curl.exe -sS -o $loginHtml -D $headersKeycloak -c $cookieJar -b $cookieJar -w "%{http_code}" $redirectToKeycloak
Assert-Status -Label "GET Keycloak login page" -Actual ([int]$keycloakPageStatus) -Expected 200

$formActionUrl = Resolve-FormActionUrl -HtmlPath $loginHtml -KeycloakBaseUrl $KeycloakBaseUrl
Write-Host "Keycloak form action resolved." -ForegroundColor Green

Write-Step "Step 3: Submit Keycloak credentials and follow callback redirects"
$submitStatus = curl.exe -sS -L -o $submitHtml -D $headersSubmit -c $cookieJar -b $cookieJar -w "%{http_code}" `
  -H "Content-Type: application/x-www-form-urlencoded" `
  --data-urlencode "username=$Username" `
  --data-urlencode "password=$Password" `
  --data-urlencode "credentialId=" `
  $formActionUrl

if ([int]$submitStatus -lt 200 -or [int]$submitStatus -ge 500) {
  throw "Submitting Keycloak credentials failed with HTTP $submitStatus."
}
Write-Host "Credential submit completed with HTTP $submitStatus." -ForegroundColor Yellow

Write-Step "Step 4: Validate authenticated session state"
$meStatus = curl.exe -sS -o $meJson -D $headersMe -c $cookieJar -b $cookieJar -w "%{http_code}" "$BackendBaseUrl/api/v1/auth/session/me"
Assert-Status -Label "GET /auth/session/me (authenticated)" -Actual ([int]$meStatus) -Expected 200

$csrfToken = Get-CookieValue -CookieJarPath $cookieJar -CookieName "ms_web_csrf"
if ([string]::IsNullOrWhiteSpace($csrfToken)) {
  throw "CSRF cookie ms_web_csrf not found after authenticated /me response."
}
Write-Host "CSRF cookie captured." -ForegroundColor Green

Write-Step "Step 5: Refresh without CSRF header must fail"
$refreshNoCsrfStatus = curl.exe -sS -o $refreshNoCsrfJson -D $headersRefreshNoCsrf -c $cookieJar -b $cookieJar -w "%{http_code}" `
  -X POST `
  -H "Origin: $WebOrigin" `
  "$BackendBaseUrl/api/v1/auth/session/refresh"
Assert-Status -Label "POST /auth/session/refresh (missing CSRF)" -Actual ([int]$refreshNoCsrfStatus) -Expected 403

Write-Step "Step 6: Refresh with invalid CSRF header must fail"
$refreshBadCsrfStatus = curl.exe -sS -o $refreshBadCsrfJson -D $headersRefreshBadCsrf -c $cookieJar -b $cookieJar -w "%{http_code}" `
  -X POST `
  -H "Origin: $WebOrigin" `
  -H "X-CSRF-Token: invalid-token" `
  "$BackendBaseUrl/api/v1/auth/session/refresh"
Assert-Status -Label "POST /auth/session/refresh (invalid CSRF)" -Actual ([int]$refreshBadCsrfStatus) -Expected 403

Write-Step "Step 7: Refresh with valid CSRF header must succeed"
$refreshStatus = curl.exe -sS -o $refreshJson -D $headersRefresh -c $cookieJar -b $cookieJar -w "%{http_code}" `
  -X POST `
  -H "Origin: $WebOrigin" `
  -H "X-CSRF-Token: $csrfToken" `
  "$BackendBaseUrl/api/v1/auth/session/refresh"
Assert-Status -Label "POST /auth/session/refresh (valid CSRF)" -Actual ([int]$refreshStatus) -Expected 200

$csrfTokenAfterRefresh = Get-CookieValue -CookieJarPath $cookieJar -CookieName "ms_web_csrf"
if (-not [string]::IsNullOrWhiteSpace($csrfTokenAfterRefresh)) {
  $csrfToken = $csrfTokenAfterRefresh
}

Write-Step "Step 8: Logout with valid CSRF header must succeed"
$logoutStatus = curl.exe -sS -o $logoutJson -D $headersLogout -c $cookieJar -b $cookieJar -w "%{http_code}" `
  -X POST `
  -H "Origin: $WebOrigin" `
  -H "X-CSRF-Token: $csrfToken" `
  "$BackendBaseUrl/api/v1/auth/session/logout"
Assert-Status -Label "POST /auth/session/logout (valid CSRF)" -Actual ([int]$logoutStatus) -Expected 200

Write-Step "Step 9: /me after logout must return unauthenticated"
$meAfterLogoutStatus = curl.exe -sS -o $meAfterLogoutJson -D $headersMeAfterLogout -c $cookieJar -b $cookieJar -w "%{http_code}" "$BackendBaseUrl/api/v1/auth/session/me"
Assert-Status -Label "GET /auth/session/me (after logout)" -Actual ([int]$meAfterLogoutStatus) -Expected 401

if (-not $SkipInvalidCredentialsCheck) {
  Write-Step "Step 10: Invalid credential attempt must not create authenticated session"

  $invalidCookieJar = Join-Path $tmpDir "cookies_invalid.txt"
  $invalidLoginUrl = "$BackendBaseUrl/api/v1/auth/session/login?return_to=%2Fproducts"
  $invalidStartStatus = curl.exe -sS -o NUL -D $headersLoginInvalid -c $invalidCookieJar -b $invalidCookieJar -w "%{http_code}" $invalidLoginUrl
  Assert-Status -Label "GET /auth/session/login (invalid path bootstrap)" -Actual ([int]$invalidStartStatus) -Expected 302

  $invalidRedirect = Get-HeaderValue -HeadersFile $headersLoginInvalid -HeaderName "Location"
  if ([string]::IsNullOrWhiteSpace($invalidRedirect)) {
    throw "Missing Location header for invalid credential bootstrap."
  }

  $invalidPageStatus = curl.exe -sS -o $invalidLoginHtml -D $headersKeycloakInvalid -c $invalidCookieJar -b $invalidCookieJar -w "%{http_code}" $invalidRedirect
  Assert-Status -Label "GET Keycloak login page (invalid path bootstrap)" -Actual ([int]$invalidPageStatus) -Expected 200

  $invalidFormActionUrl = Resolve-FormActionUrl -HtmlPath $invalidLoginHtml -KeycloakBaseUrl $KeycloakBaseUrl
  $invalidPassword = "$Password-invalid"
  $invalidSubmitStatus = curl.exe -sS -L -o $invalidSubmitHtml -D $headersSubmitInvalid -c $invalidCookieJar -b $invalidCookieJar -w "%{http_code}" `
    -H "Content-Type: application/x-www-form-urlencoded" `
    --data-urlencode "username=$Username" `
    --data-urlencode "password=$invalidPassword" `
    --data-urlencode "credentialId=" `
    $invalidFormActionUrl

  if ([int]$invalidSubmitStatus -lt 200 -or [int]$invalidSubmitStatus -ge 500) {
    throw "Invalid credential submit returned unexpected HTTP $invalidSubmitStatus."
  }
  Write-Host "Invalid credential submit completed with HTTP $invalidSubmitStatus." -ForegroundColor Yellow

  $meAfterInvalidStatus = curl.exe -sS -o $meAfterInvalidJson -D $headersMeAfterInvalid -c $invalidCookieJar -b $invalidCookieJar -w "%{http_code}" "$BackendBaseUrl/api/v1/auth/session/me"
  Assert-Status -Label "GET /auth/session/me (after invalid credentials)" -Actual ([int]$meAfterInvalidStatus) -Expected 401
}

Write-Host ""
Write-Host "Auth/session smoke completed successfully." -ForegroundColor Green
Write-Host "Evidence directory: $tmpDir" -ForegroundColor Green
