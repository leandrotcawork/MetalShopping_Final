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
        "import type {",
        "  AuthSessionLogoutResponseV1,",
        "  AuthSessionStateV1,",
        "  CommonErrorV1,",
        "  ProductsPortfolioListV1,",
        "} from ""../types_ts/contracts.generated"";",
        "",
        "export const operationIds = [",
        ($operationLines -join ",`n"),
        "] as const;",
        "",
        "export type OperationId = typeof operationIds[number];",
        "",
        "export type QueryParamValue = string | number | boolean | null | undefined;",
        "",
        "function buildQueryString(query?: Record<string, QueryParamValue>) {",
        "  const params = new URLSearchParams();",
        '  if (query === undefined) {',
        '    return "";',
        '  }',
        "",
        "  for (const [key, rawValue] of Object.entries(query)) {",
        '    if (rawValue === undefined || rawValue === null || rawValue === "") {',
        '      continue;',
        '    }',
        "    params.set(key, String(rawValue));",
        "  }",
        "",
        "  return params.toString();",
        "}",
        "",
        "export type GeneratedRequestOptions = {",
        "  query?: Record<string, QueryParamValue>;",
        "  headers?: Record<string, string>;",
        "  body?: unknown;",
        "};",
        "",
        "export type GeneratedHttpClient = {",
        "  getJson<T>(path: string, options?: GeneratedRequestOptions): Promise<T>;",
        "  postJson<T>(path: string, options?: GeneratedRequestOptions): Promise<T>;",
        "};",
        "",
        "export type BrowserGeneratedHttpClientConfig = {",
        "  baseUrl: string;",
        "  bearerToken?: string;",
        "  csrfCookieName?: string;",
        "  csrfHeaderName?: string;",
        "  defaultHeaders?: Record<string, string>;",
        "};",
        "",
        "type HttpClientError = Error & {",
        "  status?: number;",
        "  code?: string;",
        "  traceId?: string;",
        "};",
        "",
        "function createTraceId() {",
        "  if (typeof crypto !== ""undefined"" && typeof crypto.randomUUID === ""function"") {",
        "    return crypto.randomUUID();",
        "  }",
        '  return `trace-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`;',
        "}",
        "",
        "function readCookie(name: string) {",
        "  if (typeof document === ""undefined"") {",
        "    return null;",
        "  }",
        '  const prefix = `${encodeURIComponent(name)}=`;',
        "  const entries = document.cookie.split(/;\\s*/);",
        "  for (const entry of entries) {",
        "    if (!entry.startsWith(prefix)) {",
        "      continue;",
        "    }",
        "    return decodeURIComponent(entry.slice(prefix.length));",
        "  }",
        "  return null;",
        "}",
        "",
        "function buildRequestUrl(baseUrl: string, path: string, query?: Record<string, QueryParamValue>) {",
        "  const queryString = buildQueryString(query);",
        '  return `${baseUrl.replace(/\/$/, "")}${path}${queryString ? `?${queryString}` : ""}`;',
        "}",
        "",
        "function parseHttpError(response: Response, errorPayload: CommonErrorV1 | null) {",
        '  const message = errorPayload?.error?.message ?? `Request failed with status ${response.status}`;',
        "  const error = new Error(message) as HttpClientError;",
        "  error.status = response.status;",
        "  error.code = errorPayload?.error?.code;",
        "  error.traceId = errorPayload?.error?.trace_id;",
        "  return error;",
        "}",
        "",
        "export function createBrowserGeneratedHttpClient(config: BrowserGeneratedHttpClientConfig): GeneratedHttpClient {",
        '  const baseUrl = config.baseUrl.replace(/\/$/, "");',
        '  const csrfCookieName = config.csrfCookieName ?? "ms_web_csrf";',
        '  const csrfHeaderName = config.csrfHeaderName ?? "X-CSRF-Token";',
        "  const defaultHeaders = config.defaultHeaders ?? {};",
        "",
        "  function buildHeaders(method: string, headers?: Record<string, string>) {",
        "    const merged: Record<string, string> = {",
        "      Accept: ""application/json"",",
        "      ""X-Trace-Id"": createTraceId(),",
        "      ...defaultHeaders,",
        "      ...(headers ?? {}),",
        "    };",
        "    if ((config.bearerToken ?? """").trim() !== """") {",
        '      merged.Authorization = `Bearer ${config.bearerToken}`;',
        "    }",
        "    if (method !== ""GET"") {",
        "      merged[""Content-Type""] = merged[""Content-Type""] ?? ""application/json"";",
        "      const csrfToken = readCookie(csrfCookieName);",
        "      if (csrfToken) {",
        "        merged[csrfHeaderName] = csrfToken;",
        "      }",
        "    }",
        "    return merged;",
        "  }",
        "",
        "  return {",
        "    async getJson<T>(path: string, options?: GeneratedRequestOptions) {",
        "      const response = await fetch(buildRequestUrl(baseUrl, path, options?.query), {",
        '        method: "GET",',
        '        credentials: "include",',
        "        headers: buildHeaders(""GET"", options?.headers),",
        "      });",
        "      if (!response.ok) {",
        "        const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;",
        "        throw parseHttpError(response, errorPayload);",
        "      }",
        "      return (await response.json()) as T;",
        "    },",
        "    async postJson<T>(path: string, options?: GeneratedRequestOptions) {",
        "      const response = await fetch(buildRequestUrl(baseUrl, path, options?.query), {",
        '        method: "POST",',
        '        credentials: "include",',
        "        headers: buildHeaders(""POST"", options?.headers),",
        "        body: options?.body === undefined ? undefined : JSON.stringify(options.body),",
        "      });",
        "      if (!response.ok) {",
        "        const errorPayload = (await response.json().catch(() => null)) as CommonErrorV1 | null;",
        "        throw parseHttpError(response, errorPayload);",
        "      }",
        "      return (await response.json()) as T;",
        "    },",
        "  };",
        "}",
        "",
        "export type StartAuthSessionLoginQueryParams = {",
        "  return_to?: string;",
        "};",
        "",
        "export type ProductsPortfolioSortKey =",
        "  | ""pn_interno""",
        "  | ""name""",
        "  | ""brand_name""",
        "  | ""taxonomy_leaf0_name""",
        "  | ""product_status""",
        "  | ""current_price_amount""",
        "  | ""replacement_cost_amount""",
        "  | ""on_hand_quantity"";",
        "",
        "export type ProductsPortfolioSortDirection = ""asc"" | ""desc"";",
        "",
        "export type ListProductsPortfolioQueryParams = {",
        "  search?: string;",
        "  brand_name?: string;",
        "  taxonomy_leaf0_name?: string;",
        "  status?: string;",
        "  sort_key?: ProductsPortfolioSortKey;",
        "  sort_direction?: ProductsPortfolioSortDirection;",
        "  limit?: number;",
        "  offset?: number;",
        "};",
        "",
        "export type ServerCoreSdk = {",
        "  authSession: {",
        "    buildStartLoginUrl(baseUrl: string, query?: StartAuthSessionLoginQueryParams): string;",
        "    getSessionState(): Promise<AuthSessionStateV1>;",
        "    refreshSession(): Promise<AuthSessionStateV1>;",
        "    logoutSession(): Promise<AuthSessionLogoutResponseV1>;",
        "  };",
        "  products: {",
        "    listProductsPortfolio(query?: ListProductsPortfolioQueryParams): Promise<ProductsPortfolioListV1>;",
        "  };",
        "};",
        "",
        "export function createServerCoreSdk(client: GeneratedHttpClient): ServerCoreSdk {",
        "  return {",
        "    authSession: {",
        '      buildStartLoginUrl(baseUrl, query) {',
        '        const queryString = buildQueryString(query);',
        '        return `${baseUrl.replace(/\/$/, "")}/api/v1/auth/session/login${queryString ? `?${queryString}` : ""}`;',
        '      },',
        "      getSessionState() {",
        "        return client.getJson<AuthSessionStateV1>(""/api/v1/auth/session/me"");",
        "      },",
        "      refreshSession() {",
        "        return client.postJson<AuthSessionStateV1>(""/api/v1/auth/session/refresh"");",
        "      },",
        "      logoutSession() {",
        "        return client.postJson<AuthSessionLogoutResponseV1>(""/api/v1/auth/session/logout"");",
        "      },",
        "    },",
        "    products: {",
        "      listProductsPortfolio(query) {",
        "        return client.getJson<ProductsPortfolioListV1>(""/api/v1/products/portfolio"", {",
        "          query,",
        "        });",
        "      },",
        "    },",
        "  };",
        "}"
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
