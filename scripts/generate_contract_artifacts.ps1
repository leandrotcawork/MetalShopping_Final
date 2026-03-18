param(
    [ValidateSet("all", "sdk_ts", "types_ts", "sdk_py")]
    [string]$Target = "all",
    [switch]$Check
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
& (Join-Path $PSScriptRoot "validate_contracts.ps1") -Scope all

function Ensure-Directory {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        New-Item -ItemType Directory -Force -Path $Path | Out-Null
    }
}

function Write-GeneratedFile {
    param(
        [string]$Path,
        [string]$Content
    )
    if ($Check) {
        if (-not (Test-Path $Path)) {
            throw "Generated artifact missing in check mode: $Path"
        }
        $existing = Get-Content -Path $Path -Raw -Encoding UTF8
        if ($existing.TrimEnd("`r", "`n") -ne $Content.TrimEnd("`r", "`n")) {
            throw "Generated artifact out of date: $Path"
        }
        return
    }
    Set-Content -Path $Path -Value $Content -Encoding UTF8
}

function Get-OperationIds {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\api\openapi") -Filter *.yaml -File | ForEach-Object {
        $raw = Get-Content -Path $_.FullName -Raw -Encoding UTF8
        foreach ($match in [regex]::Matches($raw, 'operationId:\s*([A-Za-z0-9_]+)')) {
            $match.Groups[1].Value
        }
    } | Sort-Object -Unique
}

function Get-JsonSchemaIds {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\api\jsonschema") -Filter *.json -File | Where-Object { -not $_.BaseName.StartsWith("_") } | ForEach-Object {
        $json = (Get-Content -Path $_.FullName -Raw -Encoding UTF8 | ConvertFrom-Json)
        [PSCustomObject]@{
            Name = $_.BaseName
            Id   = $json.'$id'
        }
    } | Sort-Object Name
}

function Convert-NameToTsIdentifier {
    param([string]$Name)

    $parts = ($Name -replace '\.schema$', '') -split '[^A-Za-z0-9]+'
    $normalized = foreach ($part in $parts) {
        if ([string]::IsNullOrWhiteSpace($part)) { continue }
        $lower = $part.ToLowerInvariant()
        if ($lower -eq "id") { "Id"; continue }
        $lower.Substring(0, 1).ToUpperInvariant() + $lower.Substring(1)
    }
    return ($normalized -join '')
}

function Resolve-TsRefTypeName {
    param([string]$Ref)

    $refName = [System.IO.Path]::GetFileNameWithoutExtension($Ref)
    return Convert-NameToTsIdentifier $refName
}

function Convert-JsonSchemaPrimitiveType {
    param([string]$TypeName)

    switch ($TypeName) {
        "string" { return "string" }
        "integer" { return "number" }
        "number" { return "number" }
        "boolean" { return "boolean" }
        "null" { return "null" }
        default { return "unknown" }
    }
}

function Convert-JsonSchemaToTsType {
    param(
        $Schema,
        [int]$IndentLevel = 0
    )

    if ($null -eq $Schema) {
        return "unknown"
    }

    if ($Schema.PSObject.Properties.Name -contains '$ref') {
        return Resolve-TsRefTypeName $Schema.'$ref'
    }

    if ($Schema.PSObject.Properties.Name -contains 'enum') {
        return ((@($Schema.enum) | ForEach-Object { '"' + ($_ -replace '"', '\"') + '"' }) -join ' | ')
    }

    $schemaTypes = @()
    if ($Schema.PSObject.Properties.Name -contains 'type') {
        $schemaTypes = @($Schema.type)
    }

    if ($schemaTypes.Count -gt 1) {
        $converted = foreach ($typeName in $schemaTypes) {
            if ($typeName -eq 'null') {
                'null'
                continue
            }

            $clone = [pscustomobject]@{}
            foreach ($property in $Schema.PSObject.Properties) {
                if ($property.Name -eq 'type') {
                    continue
                }
                Add-Member -InputObject $clone -MemberType NoteProperty -Name $property.Name -Value $property.Value
            }
            Add-Member -InputObject $clone -MemberType NoteProperty -Name 'type' -Value $typeName
            Convert-JsonSchemaToTsType -Schema $clone -IndentLevel $IndentLevel
        }
        return (($converted | Select-Object -Unique) -join ' | ')
    }

    if ($schemaTypes.Count -eq 0) {
        return "unknown"
    }

    switch ($schemaTypes[0]) {
        "object" {
            if (-not ($Schema.PSObject.Properties.Name -contains 'properties')) {
                return "Record<string, unknown>"
            }

            $indent = '  ' * $IndentLevel
            $childIndent = '  ' * ($IndentLevel + 1)
            $required = @($Schema.required)
            $propertyLines = foreach ($property in $Schema.properties.PSObject.Properties) {
                $optionalMarker = if ($required -contains $property.Name) { "" } else { "?" }
                $propertyType = Convert-JsonSchemaToTsType -Schema $property.Value -IndentLevel ($IndentLevel + 1)
                "${childIndent}$($property.Name)${optionalMarker}: $propertyType;"
            }

            if ($propertyLines.Count -eq 0) {
                return "{ }"
            }

            return @(
                "{"
                ($propertyLines -join "`n")
                "$indent}"
            ) -join "`n"
        }
        "array" {
            $itemType = Convert-JsonSchemaToTsType -Schema $Schema.items -IndentLevel $IndentLevel
            return "Array<$itemType>"
        }
        default {
            return Convert-JsonSchemaPrimitiveType $schemaTypes[0]
        }
    }
}

function Get-JsonSchemaTypeDeclarations {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\api\jsonschema") -Filter *.json -File |
        Where-Object { -not $_.BaseName.StartsWith("_") } |
        Sort-Object BaseName |
        ForEach-Object {
            $schema = Get-Content -Path $_.FullName -Raw -Encoding UTF8 | ConvertFrom-Json
            $typeName = Convert-NameToTsIdentifier $_.BaseName
            $typeBody = Convert-JsonSchemaToTsType -Schema $schema -IndentLevel 0
            "export type $typeName = $typeBody;"
        }
}

function Get-EventNames {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\events") -Recurse -Filter *.json -File | Where-Object { -not $_.BaseName.StartsWith("_") } | ForEach-Object {
        $json = (Get-Content -Path $_.FullName -Raw -Encoding UTF8 | ConvertFrom-Json)
        $json.event_name
    } | Sort-Object -Unique
}

function Get-GovernanceKeys {
    Get-ChildItem -Path (Join-Path $repoRoot "contracts\governance") -Recurse -Filter *.json -File | Where-Object { -not $_.BaseName.StartsWith("_") } | ForEach-Object {
        $json = (Get-Content -Path $_.FullName -Raw -Encoding UTF8 | ConvertFrom-Json)
        foreach ($key in @($json.flag_name, $json.threshold_name, $json.policy_name)) {
            if ($key) { $key }
        }
    } | Sort-Object -Unique
}

$operationIds = Get-OperationIds
$schemaIds = Get-JsonSchemaIds
$schemaTypeDeclarations = Get-JsonSchemaTypeDeclarations
$eventNames = Get-EventNames
$governanceKeys = Get-GovernanceKeys
if ($Target -in @("all", "types_ts")) {
    $typesDir = Join-Path $repoRoot "packages\generated\types_ts"
    Ensure-Directory $typesDir
    $schemaLines = $schemaIds | ForEach-Object { "  ""$($_.Name)"": ""$($_.Id)""" }
    $schemaTypeLines = $schemaTypeDeclarations
    $eventLines = $eventNames | ForEach-Object { "  ""$_""" }
    $govLines = $governanceKeys | ForEach-Object { "  ""$_""" }
    $content = @(
        "// generated by scripts/generate_contract_artifacts.ps1",
        "export const schemaIds = {",
        ($schemaLines -join ",`n"),
        "} as const;",
        "",
        ($schemaTypeLines -join "`n`n"),
        "",
        "export const eventNames = [",
        ($eventLines -join ",`n"),
        "] as const;",
        "",
        "export const governanceKeys = [",
        ($govLines -join ",`n"),
        "] as const;"
    ) -join "`n"
    Write-GeneratedFile -Path (Join-Path $typesDir "contracts.generated.ts") -Content $content
}

if ($Target -in @("all", "sdk_ts")) {
    $sdkTsDir = Join-Path $repoRoot "packages\generated\sdk_ts"
    Ensure-Directory $sdkTsDir
    $operationLines = $operationIds | ForEach-Object { "  ""$_""" }
    $content = @(
        "// generated by scripts/generate_contract_artifacts.ps1",
        "export const operationIds = [",
        ($operationLines -join ",`n"),
        "] as const;",
        "",
        "export type OperationId = typeof operationIds[number];"
    ) -join "`n"
    Write-GeneratedFile -Path (Join-Path $sdkTsDir "openapi.generated.ts") -Content $content
}

if ($Target -in @("all", "sdk_py")) {
    $sdkPyDir = Join-Path $repoRoot "packages\generated\sdk_py"
    Ensure-Directory $sdkPyDir
    $operationLines = $operationIds | ForEach-Object { "    '$_'," }
    $content = @(
        "# generated by scripts/generate_contract_artifacts.ps1",
        "OPERATION_IDS = [",
        ($operationLines -join "`n"),
        "]"
    ) -join "`n"
    Write-GeneratedFile -Path (Join-Path $sdkPyDir "openapi_generated.py") -Content $content
}

Write-Host "Generated contract artifacts for target '$Target'." -ForegroundColor Green
