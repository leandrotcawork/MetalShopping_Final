# Data Contract Map

## Purpose

Define, before implementation, which data each frontend module needs, where it comes from, and which backend contract serves it.

This is mandatory for make-it-work-first execution.

## Module map

## 1. Home

### Frontend need

- KPI cards with operational summary
- last update timestamp

### API contract

- `GET /api/v1/home/summary`
- OpenAPI: `contracts/api/openapi/home_v1.openapi.yaml`
- Response schema: `contracts/api/jsonschema/home_summary_v1.schema.json`

### Backend data sources

- `catalog_products`
- `pricing_product_prices` (effective current rows)
- `inventory_product_positions` (effective current rows)

### Current status

- implemented

## 2. Shopping Price

### Frontend need

- runs list and run detail
- latest observed competitive price per product
- run status/progress and basic summary

### API contracts to freeze

- `GET /api/v1/shopping/summary`
- `GET /api/v1/shopping/runs`
- `GET /api/v1/shopping/runs/{run_id}`
- `GET /api/v1/shopping/products/{product_id}/latest`

### Backend data sources to create/freeze

- `shopping_price_runs`
- `shopping_price_run_items`
- `shopping_price_latest_snapshot` (or equivalent materialized read table)

### Worker integration rule

- Python worker writes scraping results to Postgres
- Go API reads from Postgres
- no synchronous Go to Python call in request path

### Current status

- pending contract freeze

## 3. Analytics

### Frontend need

- aggregated KPIs and trends required by the existing visual surface
- filterable slices by period/category/brand (to freeze by screen inventory)

### API contracts to freeze

- to be declared after screen-level data inventory

### Backend data sources

- built from canonical tables and shopping snapshots after Shopping phase

### Current status

- pending screen inventory and contract freeze

## Implementation rule

No module implementation starts before:

1. endpoint list is frozen in OpenAPI
2. data source ownership is explicit in this map
3. frontend binding path is `sdk-runtime`, never page-local transport
