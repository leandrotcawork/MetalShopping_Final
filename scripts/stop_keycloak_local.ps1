$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "ops\keycloak\docker-compose.yml"

docker compose -f $composeFile down

Write-Host "Keycloak local stopped." -ForegroundColor Yellow
