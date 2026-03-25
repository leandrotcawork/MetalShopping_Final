param(
    [string]$FrontendHost = "127.0.0.1",
    [int]$FrontendPort = 5173,
    [switch]$OpenBrowser
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $repoRoot ".env"

if (-not (Test-Path $envFile)) {
    throw "Missing .env at repo root. Copy .env.example to .env before starting the local stack."
}

$null = Get-Command go -ErrorAction Stop
$null = Get-Command npm.cmd -ErrorAction Stop

$backendCommand = "Set-Location '$repoRoot'; go run ./apps/server_core/cmd/metalshopping-server"
$frontendCommand = "Set-Location '$repoRoot'; npm.cmd --workspace @metalshopping/web run dev -- --host $FrontendHost --port $FrontendPort"

$backendProcess = Start-Process powershell `
    -WorkingDirectory $repoRoot `
    -ArgumentList @("-NoExit", "-Command", $backendCommand) `
    -PassThru

$frontendProcess = Start-Process powershell `
    -WorkingDirectory $repoRoot `
    -ArgumentList @("-NoExit", "-Command", $frontendCommand) `
    -PassThru

$webUrl = "http://{0}:{1}" -f $FrontendHost, $FrontendPort
$apiUrl = "http://127.0.0.1:8080"

Write-Host "MetalShopping local stack started." -ForegroundColor Green
Write-Host ("Backend PID:  {0} ({1})" -f $backendProcess.Id, $apiUrl)
Write-Host ("Frontend PID: {0} ({1})" -f $frontendProcess.Id, $webUrl)
Write-Host "Close the two spawned PowerShell windows to stop the stack."

if ($OpenBrowser) {
    Start-Process $webUrl | Out-Null
}
