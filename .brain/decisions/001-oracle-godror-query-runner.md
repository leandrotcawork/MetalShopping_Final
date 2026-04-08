# ADR-001: Oracle Connectivity via Godror Query Runner

**Date:** 2026-04-06
**Status:** accepted

## Context
MetalShopping must ingest data from Sankhya Oracle databases for ERP integration.
We need a production-grade Oracle driver and a safe, typed row-reading API for ingestion.

## Decision
Use the godror driver for Oracle connectivity and a typed query-runner/row-reader API
as the standard ERP Oracle access pattern in integration_worker.

## Rationale
Godror is the production-grade Go driver for Oracle and supports robust connection
handling and performance. A typed row reader reduces scan ambiguity and centralizes
column handling for ingestion reliability.

## Consequences
Oracle connectivity depends on godror and its system prerequisites.
DB access code follows the query-runner abstraction for consistency and testing.

## Alternatives Considered
- go-oci8: older driver with more operational friction and less Go-native support.
