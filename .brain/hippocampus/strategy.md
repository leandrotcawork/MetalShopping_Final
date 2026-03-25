---
id: hippocampus-strategy
title: Product Strategy & Goals
region: hippocampus
tags: [strategy, goals, business, product]
links:
  - hippocampus/architecture
weight: 0.85
updated_at: 2026-03-24T10:00:00Z
---

# MetalShopping: Product Strategy & Goals

## Product Vision

MetalShopping is a **server-first B2B platform for commercial operations** — enabling companies to manage strategy, pricing, analytics, procurement, and CRM in a single, unified interface. Production-intended v1.

## Core Goals

1. **Commercial Strategy** — Pricing rules, competitive positioning, market response
2. **Analytics** — Real-time visibility into sales, margins, customer behavior
3. **Procurement** — Supplier management, order fulfillment, cost control
4. **CRM** — Customer relationships, account management, opportunity tracking
5. **Unified Data Model** — Single source of truth across all domains

## Business Constraints

- **Multi-tenant SaaS** — Strict tenant isolation, no cross-tenant data leaks
- **Production grade** — v1 must be suitable for live customer data
- **Compliance-ready** — Audit trails, data retention, GDPR-compatible
- **Real-time analytics** — Customers see up-to-date insights, not stale data
- **API-first** — All features exposed via contracts, not just UI

## Current Phase: 3A — Foundation Hardening

- Analytics legacy migration (Home, Products, Taxonomy, Brands surfaces)
- Shopping module coming next
- Focus: **quality and reliability** over new features

## Success Metrics

- **User adoption** — How many customers use each module
- **Data accuracy** — Are analytics insights trusted by customers?
- **System reliability** — Uptime SLA, error rate
- **Performance** — Page load times, API latency (p99 < 200ms)
- **Cost efficiency** — Cloud spend per customer transaction

## Technical Priorities (Ranked)

1. **Reliability** — Tenant isolation cannot fail; data must never be corrupted
2. **Correctness** — Analytics must be accurate; business logic must not have bugs
3. **Maintainability** — Code patterns must be clear and reusable
4. **Performance** — Customer experience must not degrade as data grows
5. **Developer experience** — New modules must be easy to add without mistakes
6. **Cost efficiency** — Infrastructure must scale linearly with customer base

## Design Philosophy

- **Make it work → make it beautiful → make it fast** (in that order)
- Not a prototype; treat v1 as production from day one
- Engineering bar: Stripe / Google senior engineer approval standard

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.85
