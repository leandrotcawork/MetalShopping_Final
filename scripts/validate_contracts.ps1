param(
    [ValidateSet("all", "api", "events", "governance")]
    [string]$Scope = "all"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$errors = New-Object System.Collections.Generic.List[string]

function Add-ValidationError {
    param([string]$Message)
    $errors.Add($Message)
}

function Test-JsonFile {
    param([string]$Path)
    try {
        $raw = Get-Content -Path $Path -Raw -Encoding UTF8
        $parsed = $raw | ConvertFrom-Json
        return $parsed
    } catch {
        Add-ValidationError "Invalid JSON: $Path :: $($_.Exception.Message)"
        return $null
    }
}

function Test-OpenAPIYamlFile {
    param([string]$Path)
    $raw = Get-Content -Path $Path -Raw -Encoding UTF8
    if ($raw -notmatch '(?m)^openapi:\s*3\.') {
        Add-ValidationError "OpenAPI file missing valid openapi version header: $Path"
    }
    if ($raw -notmatch '(?m)^info:\s*$') {
        Add-ValidationError "OpenAPI file missing info section: $Path"
    }
    if ($raw -notmatch '(?m)^paths:\s*$') {
        Add-ValidationError "OpenAPI file missing paths section: $Path"
    }
    if ($raw -notmatch 'operationId:\s*[A-Za-z0-9_]') {
        Add-ValidationError "OpenAPI file has no operationId entries: $Path"
    }
}

function Resolve-RelativeContractPath {
    param(
        [string]$BasePath,
        [string]$RelativePath
    )
    $baseDir = Split-Path -Parent $BasePath
    return [System.IO.Path]::GetFullPath((Join-Path $baseDir $RelativePath))
}

function Test-TemplateFile {
    param([System.IO.FileInfo]$File)
    return $File.BaseName.StartsWith("_")
}

function Test-PlaceholderPath {
    param([string]$Path)
    return $Path -match '[<>]'
}

if ($Scope -in @("all", "api")) {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\api\jsonschema") -Filter *.json -File | Where-Object { -not (Test-TemplateFile $_) } | ForEach-Object {
        $json = Test-JsonFile $_.FullName
        if ($null -eq $json) { return }
        if (-not $json.'$id') {
            Add-ValidationError "JSON Schema missing `$id: $($_.FullName)"
        }
        if (-not $json.title) {
            Add-ValidationError "JSON Schema missing title: $($_.FullName)"
        }
    }

    Get-ChildItem -Path (Join-Path $repoRoot "contracts\api\openapi") -Filter *.yaml -File | ForEach-Object {
        Test-OpenAPIYamlFile $_.FullName
    }
}

if ($Scope -in @("all", "events")) {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\events") -Recurse -Filter *.json -File | Where-Object { -not (Test-TemplateFile $_) } | ForEach-Object {
        $event = Test-JsonFile $_.FullName
        if ($null -eq $event) { return }
        foreach ($required in @("event_name", "version", "bounded_context", "payload_schema_ref")) {
            if (-not $event.$required) {
                Add-ValidationError "Event contract missing ${required}: $($_.FullName)"
            }
        }
        if ($event.payload_schema_ref -and -not (Test-PlaceholderPath $event.payload_schema_ref)) {
            $target = Resolve-RelativeContractPath -BasePath $_.FullName -RelativePath $event.payload_schema_ref
            if (-not (Test-Path $target)) {
                Add-ValidationError "Event payload schema ref does not exist: $($_.FullName) -> $target"
            }
        }
    }
}

if ($Scope -in @("all", "governance")) {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\governance") -Recurse -Filter *.json -File | Where-Object { -not (Test-TemplateFile $_) } | ForEach-Object {
        $artifact = Test-JsonFile $_.FullName
        if ($null -eq $artifact) { return }
        $fileName = $_.Name
        if ($fileName -like '*.feature_flag.json') {
            foreach ($required in @("flag_name", "version", "bounded_context", "schema_ref")) {
                if (-not $artifact.$required) {
                    Add-ValidationError "Feature flag contract missing ${required}: $($_.FullName)"
                }
            }
        } elseif ($fileName -like '*.threshold.json') {
            foreach ($required in @("threshold_name", "version", "bounded_context", "schema_ref")) {
                if (-not $artifact.$required) {
                    Add-ValidationError "Threshold contract missing ${required}: $($_.FullName)"
                }
            }
        } elseif ($fileName -like '*.policy.json') {
            foreach ($required in @("policy_name", "version", "bounded_context", "schema_ref")) {
                if (-not $artifact.$required) {
                    Add-ValidationError "Policy contract missing ${required}: $($_.FullName)"
                }
            }
        }

        if ($artifact.schema_ref -and -not (Test-PlaceholderPath $artifact.schema_ref)) {
            $target = Resolve-RelativeContractPath -BasePath $_.FullName -RelativePath $artifact.schema_ref
            if (-not (Test-Path $target)) {
                Add-ValidationError "Governance schema ref does not exist: $($_.FullName) -> $target"
            }
        }
    }
}

if ($errors.Count -gt 0) {
    Write-Host "Contract validation failed:" -ForegroundColor Red
    $errors | ForEach-Object { Write-Host " - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "Contract validation passed for scope '$Scope'." -ForegroundColor Green
