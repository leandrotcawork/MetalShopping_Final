# Products Readmodel Ownership

## Purpose

Freeze where the first `Products` surface read composition belongs so the frontend does not become a hidden composition layer.

## Rule

The `Products` portfolio surface is a backend-owned operational readmodel.

It is not:

- a new business write module
- a frontend-side composition of three unrelated APIs
- a BFF owned by the web app

## Source owners

The `Products` read surface consolidates data from:

- `catalog`
  - product identity
  - identifiers
  - taxonomy
  - stable master data
- `pricing`
  - current effective price
  - replacement cost
  - average cost
- `inventory`
  - current effective stock position

## Backend ownership

The composition belongs in `server_core` as an operational read surface.

Initial implementation options that respect the current architecture:

- readmodel under an existing module owner such as `catalog/readmodel`
- dedicated operational HTTP transport backed by module-owned read queries

The composition must not become a new canonical write owner.

## Frontend rule

The frontend consumes a single generated contract for `Products`.

The frontend does not:

- call `catalog`, `pricing`, and `inventory` separately and stitch them into a portfolio table as the primary design
- define manual DTOs that compete with the generated contract
- treat the page component as the aggregation layer

## Why

This keeps:

- business ownership explicit
- client complexity low
- future mobile or desktop reuse straightforward
- contract evolution controlled by the backend and `contracts/`
