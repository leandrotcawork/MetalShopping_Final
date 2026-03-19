$ErrorActionPreference = "Stop"

$keycloakBaseUrl = "http://127.0.0.1:18081"
$realmName = "metalshopping"
$themeName = "metalshopping"
$readinessUrl = "$keycloakBaseUrl/realms/$realmName/.well-known/openid-configuration"
$maxAttempts = 60

for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
  try {
    $readinessResponse = Invoke-WebRequest -Uri $readinessUrl -UseBasicParsing -TimeoutSec 5
    if ($readinessResponse.StatusCode -eq 200) {
      break
    }
  } catch {
    Start-Sleep -Seconds 2
  }

  if ($attempt -eq $maxAttempts) {
    throw "Keycloak did not become ready in time at $readinessUrl."
  }
}

$tokenResponse = curl.exe -sS -X POST "$keycloakBaseUrl/realms/master/protocol/openid-connect/token" `
  -H "Content-Type: application/x-www-form-urlencoded" `
  -d "client_id=admin-cli&username=admin&password=admin&grant_type=password"

if ($LASTEXITCODE -ne 0) {
  throw "Failed to obtain an admin token from Keycloak at $keycloakBaseUrl."
}

$tokenResponseJson = $tokenResponse | ConvertFrom-Json
$accessToken = $tokenResponseJson.access_token
if ([string]::IsNullOrWhiteSpace($accessToken)) {
  throw "Keycloak admin token response did not include an access_token. Response: $tokenResponse"
}

$payload = @{
  loginTheme = $themeName
} | ConvertTo-Json -Compress

$headers = @{
  Authorization = "Bearer $accessToken"
  "Content-Type" = "application/json"
}

Invoke-RestMethod `
  -Uri "$keycloakBaseUrl/admin/realms/$realmName" `
  -Method PUT `
  -Headers $headers `
  -Body $payload | Out-Null

Write-Host "Applied Keycloak login theme '$themeName' to realm '$realmName'." -ForegroundColor Green
