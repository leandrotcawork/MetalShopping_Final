$ErrorActionPreference = "Stop"

$keycloakBaseUrl = "http://127.0.0.1:18081"
$realmName = "metalshopping"
$themeName = "metalshopping"

$tokenResponse = curl.exe -sS -X POST "$keycloakBaseUrl/realms/master/protocol/openid-connect/token" `
  -H "Content-Type: application/x-www-form-urlencoded" `
  -d "client_id=admin-cli&username=admin&password=admin&grant_type=password"

if ($LASTEXITCODE -ne 0) {
  throw "Failed to obtain an admin token from Keycloak at $keycloakBaseUrl."
}

$accessToken = ($tokenResponse | ConvertFrom-Json).access_token
if ([string]::IsNullOrWhiteSpace($accessToken)) {
  throw "Keycloak admin token response did not include an access_token."
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
