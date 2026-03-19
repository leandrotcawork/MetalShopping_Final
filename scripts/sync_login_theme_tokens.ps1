param(
    [switch]$Check
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$sourcePath = Join-Path $repoRoot "packages\feature-auth-session\src\login.tokens.css"
$targetPaths = @(
    (Join-Path $repoRoot "ops\keycloak\themes\metalshopping\login\resources\css\login.tokens.css")
)

if (-not (Test-Path $sourcePath)) {
    throw "Login tokens source file not found: $sourcePath"
}

$sourceContent = Get-Content -Path $sourcePath -Raw -Encoding UTF8

foreach ($targetPath in $targetPaths) {
    if ($Check) {
        if (-not (Test-Path $targetPath)) {
            throw "Login token target file missing in check mode: $targetPath"
        }

        $targetContent = Get-Content -Path $targetPath -Raw -Encoding UTF8
        if ($sourceContent.TrimEnd("`r", "`n") -ne $targetContent.TrimEnd("`r", "`n")) {
            throw "Login token target out of date: $targetPath"
        }
        continue
    }

    $targetDirectory = Split-Path -Parent $targetPath
    if (-not (Test-Path $targetDirectory)) {
        New-Item -ItemType Directory -Path $targetDirectory -Force | Out-Null
    }

    Set-Content -Path $targetPath -Value $sourceContent -Encoding UTF8
}

if ($Check) {
    Write-Host "Login theme tokens are in sync." -ForegroundColor Green
} else {
    Write-Host "Login theme tokens synced." -ForegroundColor Green
}
