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

function Get-RepoRelativeUnixPath {
    param([string]$Path)

    $resolvedPath = [System.IO.Path]::GetFullPath($Path)
    $resolvedRepoRoot = [System.IO.Path]::GetFullPath($repoRoot)
    $relativePath = $resolvedPath.Substring($resolvedRepoRoot.Length).TrimStart('\', '/')
    return ($relativePath -replace '\\', '/')
}

function Get-FilteredGeneratedSdkFiles {
    param([string]$RootPath)

    if (-not (Test-Path $RootPath)) {
        return @()
    }

    return Get-ChildItem -Path $RootPath -Recurse -File | Where-Object {
        $relativePath = $_.FullName.Substring($RootPath.Length).TrimStart('\', '/')
        -not (
            $relativePath.StartsWith("docs\") -or
            $relativePath.StartsWith(".openapi-generator\") -or
            $relativePath -eq ".openapi-generator-ignore"
        )
    }
}

function Sync-GeneratedSdkDirectory {
    param(
        [string]$SourceRoot,
        [string]$TargetRoot
    )

    $sourceFiles = Get-FilteredGeneratedSdkFiles -RootPath $SourceRoot
    $expectedRelativePaths = $sourceFiles | ForEach-Object {
        $_.FullName.Substring($SourceRoot.Length).TrimStart('\', '/')
    } | Sort-Object

    if ($Check) {
        if (-not (Test-Path $TargetRoot)) {
            throw "Generated artifact missing in check mode: $TargetRoot"
        }

        $actualRelativePaths = Get-FilteredGeneratedSdkFiles -RootPath $TargetRoot | ForEach-Object {
            $_.FullName.Substring($TargetRoot.Length).TrimStart('\', '/')
        } | Sort-Object

        $comparison = Compare-Object -ReferenceObject $expectedRelativePaths -DifferenceObject $actualRelativePaths
        if ($comparison) {
            throw "Generated artifact directory out of date: $TargetRoot"
        }

        foreach ($relativePath in $expectedRelativePaths) {
            $expectedContent = Get-Content -Path (Join-Path $SourceRoot $relativePath) -Raw -Encoding UTF8
            $actualContent = Get-Content -Path (Join-Path $TargetRoot $relativePath) -Raw -Encoding UTF8
            if ($expectedContent.TrimEnd("`r", "`n") -ne $actualContent.TrimEnd("`r", "`n")) {
                throw "Generated artifact out of date: $(Join-Path $TargetRoot $relativePath)"
            }
        }
        return
    }

    if (Test-Path $TargetRoot) {
        Remove-Item -Path $TargetRoot -Recurse -Force
    }
    Ensure-Directory $TargetRoot

    foreach ($sourceFile in $sourceFiles) {
        $relativePath = $sourceFile.FullName.Substring($SourceRoot.Length).TrimStart('\', '/')
        $targetPath = Join-Path $TargetRoot $relativePath
        $targetDir = Split-Path -Parent $targetPath
        Ensure-Directory $targetDir
        Copy-Item -Path $sourceFile.FullName -Destination $targetPath -Force
    }
}

function Invoke-OpenApiGeneratorTsFetch {
    param(
        [string]$InputSpecPath,
        [string]$OutputPath
    )

    Ensure-Directory (Split-Path -Parent $OutputPath)
    if (Test-Path $OutputPath) {
        Remove-Item -Path $OutputPath -Recurse -Force
    }

    $dockerImage = "openapitools/openapi-generator-cli:v7.20.0"
    $inputSpecUnixPath = Get-RepoRelativeUnixPath -Path $InputSpecPath
    $outputUnixPath = Get-RepoRelativeUnixPath -Path $OutputPath
    $dockerArgs = @(
        "run",
        "--rm",
        "-v",
        "${repoRoot}:/local",
        $dockerImage,
        "generate",
        "-i",
        "/local/$inputSpecUnixPath",
        "-g",
        "typescript-fetch",
        "-o",
        "/local/$outputUnixPath",
        "-p",
        "modelPropertyNaming=original,enumPropertyNaming=original,useSingleRequestParameter=true"
    )

    & docker @dockerArgs
    if ($LASTEXITCODE -ne 0) {
        throw "OpenAPI Generator failed for spec: $InputSpecPath"
    }
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
    $sdkGeneratedDir = Join-Path $sdkTsDir "generated"
    $sdkCacheDir = Join-Path $repoRoot ".cache\openapi-generator-sdk-ts"
    $sdkContracts = @(
        @{ Name = "auth_session"; Spec = "contracts\api\openapi\auth_session_v1.openapi.yaml" },
        @{ Name = "catalog"; Spec = "contracts\api\openapi\catalog_v1.openapi.yaml" },
        @{ Name = "iam"; Spec = "contracts\api\openapi\iam_v1.openapi.yaml" },
        @{ Name = "inventory"; Spec = "contracts\api\openapi\inventory_v1.openapi.yaml" },
        @{ Name = "pricing"; Spec = "contracts\api\openapi\pricing_v1.openapi.yaml" },
        @{ Name = "products"; Spec = "contracts\api\openapi\products_v1.openapi.yaml" }
    )

    foreach ($sdkContract in $sdkContracts) {
        $specPath = Join-Path $repoRoot $sdkContract.Spec
        $cacheOutputDir = Join-Path $sdkCacheDir $sdkContract.Name
        $targetOutputDir = Join-Path $sdkGeneratedDir $sdkContract.Name
        Invoke-OpenApiGeneratorTsFetch -InputSpecPath $specPath -OutputPath $cacheOutputDir
        Sync-GeneratedSdkDirectory -SourceRoot $cacheOutputDir -TargetRoot $targetOutputDir
    }

    $operationLines = $operationIds | ForEach-Object { "  ""$_""" }
    $content = @(
        "// generated by scripts/generate_contract_artifacts.ps1",
        "import {",
        "  DefaultApi as AuthSessionDefaultApi,",
        "} from ""./generated/auth_session/apis/DefaultApi"";",
        "import {",
        "  FetchError as AuthSessionFetchError,",
        "  type FetchAPI,",
        "  type HTTPHeaders,",
        "  ResponseError as AuthSessionResponseError,",
        "  Configuration as AuthSessionConfiguration,",
        "} from ""./generated/auth_session/runtime"";",
        "import {",
        "  DefaultApi as ProductsDefaultApi,",
        "} from ""./generated/products/apis/DefaultApi"";",
        "import {",
        "  FetchError as ProductsFetchError,",
        "  ListProductsPortfolioSortDirectionEnum,",
        "  ListProductsPortfolioSortKeyEnum,",
        "  ResponseError as ProductsResponseError,",
        "  Configuration as ProductsConfiguration,",
        "} from ""./generated/products/index"";",
        "import type {",
        "  AuthSessionLogoutResponseV1,",
        "  AuthSessionStateV1,",
        "  CommonErrorV1,",
        "  ProductsPortfolioListV1,",
        "} from ""../types_ts/contracts.generated"";",
        "",
        "export * as AuthSessionGenerated from ""./generated/auth_session/index"";",
        "export * as CatalogGenerated from ""./generated/catalog/index"";",
        "export * as IamGenerated from ""./generated/iam/index"";",
        "export * as InventoryGenerated from ""./generated/inventory/index"";",
        "export * as PricingGenerated from ""./generated/pricing/index"";",
        "export * as ProductsGenerated from ""./generated/products/index"";",
        "",
        "export const operationIds = [",
        ($operationLines -join ",`n"),
        "] as const;",
        "",
        "export type OperationId = typeof operationIds[number];",
        "",
        "export type QueryParamValue = string | number | boolean | null | undefined;",
        "",
        "export type GeneratedHttpClient = {",
        "  baseUrl: string;",
        "  fetchApi: FetchAPI;",
        "  defaultHeaders?: Record<string, string>;",
        "  bearerToken?: string;",
        "  csrfCookieName: string;",
        "  csrfHeaderName: string;",
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
        "function buildRequestUrl(baseUrl: string, path: string, query?: Record<string, QueryParamValue>) {",
        "  const queryString = buildQueryString(query);",
        '  return `${baseUrl.replace(/\/$/, "")}${path}${queryString ? `?${queryString}` : ""}`;',
        "}",
        "",
        "async function parseCommonError(response: Response) {",
        "  return (await response.clone().json().catch(() => null)) as CommonErrorV1 | null;",
        "}",
        "",
        "function mapCommonError(response: Response, errorPayload: CommonErrorV1 | null) {",
        '  const message = errorPayload?.error?.message ?? `Request failed with status ${response.status}`;',
        "  const error = new Error(message) as HttpClientError;",
        "  error.status = response.status;",
        "  error.code = errorPayload?.error?.code;",
        "  error.traceId = errorPayload?.error?.trace_id;",
        "  return error;",
        "}",
        "",
        "async function throwGeneratedError(error: unknown): Promise<never> {",
        "  if (error instanceof AuthSessionResponseError || error instanceof ProductsResponseError) {",
        "    const errorPayload = await parseCommonError(error.response);",
        "    throw mapCommonError(error.response, errorPayload);",
        "  }",
        "",
        "  if (error instanceof AuthSessionFetchError || error instanceof ProductsFetchError) {",
        "    throw error.cause;",
        "  }",
        "",
        "  throw error;",
        "}",
        "",
        "function createBrowserFetch(config: BrowserGeneratedHttpClientConfig): FetchAPI {",
        '  const csrfCookieName = config.csrfCookieName ?? "ms_web_csrf";',
        '  const csrfHeaderName = config.csrfHeaderName ?? "X-CSRF-Token";',
        "  const defaultHeaders = config.defaultHeaders ?? {};",
        "",
        "  return async (input, init) => {",
        "    const request = input instanceof Request ? input : null;",
        "    const headers = new Headers(init?.headers ?? request?.headers ?? undefined);",
        '    headers.set("Accept", headers.get("Accept") ?? "application/json");',
        '    headers.set("X-Trace-Id", headers.get("X-Trace-Id") ?? createTraceId());',
        "",
        "    if ((config.bearerToken ?? """").trim() !== """" && !headers.has(""Authorization"")) {",
        '      headers.set("Authorization", `Bearer ${config.bearerToken}`);',
        "    }",
        "",
        '    const method = (init?.method ?? request?.method ?? "GET").toUpperCase();',
        '    if (!["GET", "HEAD", "OPTIONS"].includes(method)) {',
        "      const csrfToken = readCookie(csrfCookieName);",
        "      if (csrfToken && !headers.has(csrfHeaderName)) {",
        "        headers.set(csrfHeaderName, csrfToken);",
        "      }",
        "    }",
        "",
        "    for (const [name, value] of Object.entries(defaultHeaders)) {",
        "      if (!headers.has(name)) {",
        "        headers.set(name, value);",
        "      }",
        "    }",
        "",
        "    return fetch(input, {",
        "      ...init,",
        '      credentials: init?.credentials ?? "include",',
        "      headers,",
        "    });",
        "  };",
        "}",
        "",
        "export function createBrowserGeneratedHttpClient(config: BrowserGeneratedHttpClientConfig): GeneratedHttpClient {",
        "  return {",
        '    baseUrl: config.baseUrl.replace(/\/$/, ""),',
        "    fetchApi: createBrowserFetch(config),",
        "    defaultHeaders: config.defaultHeaders,",
        "    bearerToken: config.bearerToken,",
        '    csrfCookieName: config.csrfCookieName ?? "ms_web_csrf",',
        '    csrfHeaderName: config.csrfHeaderName ?? "X-CSRF-Token",',
        "  };",
        "}",
        "",
        "export function createServerCoreSdk(client: GeneratedHttpClient): ServerCoreSdk {",
        "  const authSessionApi = new AuthSessionDefaultApi(",
        "    new AuthSessionConfiguration({",
        "      basePath: client.baseUrl,",
        '      credentials: "include",',
        "      fetchApi: client.fetchApi,",
        "      headers: client.defaultHeaders as HTTPHeaders | undefined,",
        "    }),",
        "  );",
        "",
        "  const productsApi = new ProductsDefaultApi(",
        "    new ProductsConfiguration({",
        "      basePath: client.baseUrl,",
        '      credentials: "include",',
        "      fetchApi: client.fetchApi,",
        "      headers: client.defaultHeaders as HTTPHeaders | undefined,",
        "    }),",
        "  );",
        "",
        "  return {",
        "    authSession: {",
        "      buildStartLoginUrl(baseUrl, query) {",
        '        return buildRequestUrl(baseUrl.replace(/\/$/, ""), "/api/v1/auth/session/login", query);',
        "      },",
        "      async getSessionState() {",
        "        try {",
        "          return (await authSessionApi.getAuthSessionState()) as unknown as AuthSessionStateV1;",
        "        } catch (error) {",
        "          return throwGeneratedError(error);",
        "        }",
        "      },",
        "      async refreshSession() {",
        "        try {",
        "          return (await authSessionApi.refreshAuthSession({ xCSRFToken: readCookie(client.csrfCookieName) ?? """" })) as unknown as AuthSessionStateV1;",
        "        } catch (error) {",
        "          return throwGeneratedError(error);",
        "        }",
        "      },",
        "      async logoutSession() {",
        "        try {",
        "          return (await authSessionApi.logoutAuthSession({ xCSRFToken: readCookie(client.csrfCookieName) ?? """" })) as unknown as AuthSessionLogoutResponseV1;",
        "        } catch (error) {",
        "          return throwGeneratedError(error);",
        "        }",
        "      },",
        "    },",
        "    products: {",
        "      async listProductsPortfolio(query = {}) {",
        "        try {",
        "          return (await productsApi.listProductsPortfolio({",
        "            search: query.search,",
        "            brandName: query.brand_name,",
        "            taxonomyLeaf0Name: query.taxonomy_leaf0_name,",
        "            status: query.status,",
        "            sortKey: query.sort_key ? ListProductsPortfolioSortKeyEnum[query.sort_key] : undefined,",
        "            sortDirection: query.sort_direction ? ListProductsPortfolioSortDirectionEnum[query.sort_direction] : undefined,",
        "            limit: query.limit,",
        "            offset: query.offset,",
        "          })) as unknown as ProductsPortfolioListV1;",
        "        } catch (error) {",
        "          return throwGeneratedError(error);",
        "        }",
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
