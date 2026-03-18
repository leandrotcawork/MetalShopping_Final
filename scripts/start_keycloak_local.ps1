$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "ops\keycloak\docker-compose.yml"

docker compose -f $composeFile up -d

Write-Host "Keycloak local started at http://127.0.0.1:18081" -ForegroundColor Green
