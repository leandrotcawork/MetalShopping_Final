param(
    [string]$BaseRef = "",
    [switch]$SkipWorkingTree
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot

try {
    function Normalize-Path {
        param([string]$PathValue)
        return ($PathValue -replace '\\', '/').Trim()
    }

    function Add-ChangedPath {
        param(
            [System.Collections.Generic.HashSet[string]]$Set,
            [string]$PathValue
        )
        $normalized = Normalize-Path $PathValue
        if ([string]::IsNullOrWhiteSpace($normalized)) {
            return
        }
        [void]$Set.Add($normalized)
    }

    $changedPathSet = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)

    $effectiveBaseRef = $BaseRef.Trim()
    if ([string]::IsNullOrWhiteSpace($effectiveBaseRef)) {
        if ([string]::IsNullOrWhiteSpace($env:MS_SOT_BASE_REF)) {
            $effectiveBaseRef = ""
        } else {
            $effectiveBaseRef = $env:MS_SOT_BASE_REF.Trim()
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($effectiveBaseRef)) {
        $diffLines = git diff --name-only "$effectiveBaseRef...HEAD"
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to diff against base ref '$effectiveBaseRef'."
        }
        foreach ($line in $diffLines) {
            Add-ChangedPath -Set $changedPathSet -PathValue $line
        }
    }

    if (-not $SkipWorkingTree) {
        $statusLines = git status --porcelain
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to read git status."
        }

        foreach ($line in $statusLines) {
            if ([string]::IsNullOrWhiteSpace($line)) {
                continue
            }

            $path = $line.Substring(3).Trim()
            if ($path.Contains(" -> ")) {
                $path = $path.Split(" -> ")[1].Trim()
            }
            Add-ChangedPath -Set $changedPathSet -PathValue $path
        }
    }

    if ($changedPathSet.Count -eq 0) {
        if (-not [string]::IsNullOrWhiteSpace($effectiveBaseRef)) {
            Write-Host "SoT doc drift check passed: no structural changes against base ref '$effectiveBaseRef'." -ForegroundColor Green
        } else {
            Write-Host "SoT doc drift check passed: no working tree changes." -ForegroundColor Green
        }
        exit 0
    }

    $changedPaths = @($changedPathSet)

    $docPaths = @(
        "docs/PROJECT_SOT.md",
        "docs/PROGRESS.md"
    )
    $docChanged = $false
    foreach ($docPath in $docPaths) {
        if ($changedPaths -contains $docPath) {
            $docChanged = $true
            break
        }
    }

    $structuralPrefixes = @(
        "package.json",
        "apps/web/package.json",
        "apps/web/tsconfig.json",
        "apps/web/vite.config.ts",
        "apps/server_core/cmd/metalshopping-server/",
        "apps/web/src/app/",
        "packages/feature-auth-session/package.json",
        "packages/feature-auth-session/src/",
        "packages/feature-products/package.json",
        "packages/generated-types/package.json",
        "packages/generated-types/src/",
        "packages/platform-sdk/package.json",
        "ops/keycloak/themes/metalshopping/login/",
        ".github/workflows/",
        "scripts/generate_contract_artifacts.ps1",
        "packages/platform-sdk/src/"
    )

    $structuralChanged = $false
    foreach ($changedPath in $changedPaths) {
        foreach ($prefix in $structuralPrefixes) {
            if ($changedPath.StartsWith($prefix)) {
                $structuralChanged = $true
                break
            }
        }
        if ($structuralChanged) {
            break
        }
    }

    if ($structuralChanged -and -not $docChanged) {
        throw "SoT doc drift detected: structural code changed without updates to docs/PROJECT_SOT.md and docs/PROGRESS.md."
    }

    Write-Host "SoT doc drift check passed." -ForegroundColor Green
}
finally {
    Pop-Location
}
