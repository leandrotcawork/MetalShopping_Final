$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "ops\keycloak\docker-compose.yml"
$readinessUrl = "http://127.0.0.1:18081/realms/metalshopping/.well-known/openid-configuration"
$maxAttempts = 60

docker compose -f $composeFile up -d

for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
    try {
        $response = Invoke-WebRequest -Uri $readinessUrl -UseBasicParsing -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            Write-Host "Keycloak local started at http://127.0.0.1:18081" -ForegroundColor Green
            Write-Host "OIDC discovery is ready: $readinessUrl" -ForegroundColor Green
            exit 0
        }
    } catch {
        Start-Sleep -Seconds 2
    }
}

throw "Keycloak local did not become ready in time. Check 'docker logs metalshopping-keycloak'."
