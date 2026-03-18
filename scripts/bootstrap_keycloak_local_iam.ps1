$ErrorActionPreference = "Stop"

$apiBaseUrl = if ($env:MS_BOOTSTRAP_API_BASE_URL) { $env:MS_BOOTSTRAP_API_BASE_URL.TrimEnd("/") } else { "http://127.0.0.1:8080" }
$bearerToken = if ($env:MS_BOOTSTRAP_BEARER_TOKEN) { $env:MS_BOOTSTRAP_BEARER_TOKEN } else { "local-dev-token" }

$headers = @{
  Authorization = "Bearer $bearerToken"
  "Content-Type" = "application/json"
}

$assignments = @(
  @{
    UserId = "11111111-1111-1111-1111-111111111111"
    Payload = @{
      display_name = "MetalShopping Admin"
      role = "admin"
    }
  },
  @{
    UserId = "22222222-2222-2222-2222-222222222222"
    Payload = @{
      display_name = "MetalShopping Viewer"
      role = "viewer"
    }
  }
)

foreach ($assignment in $assignments) {
  $uri = "$apiBaseUrl/api/v1/iam/users/$($assignment.UserId)/roles"
  $body = $assignment.Payload | ConvertTo-Json -Depth 5
  Invoke-RestMethod -Method Post -Uri $uri -Headers $headers -Body $body | Out-Null
}

Write-Host "Local Keycloak IAM bootstrap completed." -ForegroundColor Green
