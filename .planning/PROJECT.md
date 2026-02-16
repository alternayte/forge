# Forge

## What This Is

Forge is a schema-driven, hypermedia-native Go framework for building production SaaS applications. A single resource schema definition generates migration SQL, type-safe queries, validation rules, Templ form/list components, OpenAPI documentation, and route handlers — eliminating the manual synchronization that plagues the Go ecosystem. It targets Go developers building server-rendered SaaS apps with real-time capabilities.

## Core Value

Schema is the single source of truth: define a resource once and everything — migrations, types, queries, validation, views, API docs, and handlers — is generated automatically with zero manual sync.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — ship to validate)

### Active

<!-- Current scope. Building toward these. -->

- [ ] Schema DSL with field types (UUID, String, Text, Int, BigInt, Decimal, Bool, DateTime, Date, Enum, JSON, Slug, Email, URL), modifiers (Required, MaxLen, MinLen, Sortable, Filterable, Searchable, Unique, Index, Default, Immutable, Label, Placeholder, Help, Visibility, Mutability), and relationships (BelongsTo, HasMany, HasOne, ManyToMany)
- [ ] go/ast parser that extracts schema definitions without compiling gen/ imports
- [ ] Code generator producing: Go model types, Bob query builder mods, validation functions, Atlas desired state HCL
- [ ] Atlas integration for declarative migration diffing and application
- [ ] Action layer with generated interfaces, default implementations, and embedding-based overrides
- [ ] Error handling with forge.Error, DB error mapping, and panic recovery
- [ ] Scaffolded Templ views (form, list, detail) generated once into resources/
- [ ] Scaffolded HTML handlers calling the action layer
- [ ] Form primitives library (FormField, inputs, error display)
- [ ] Datastar SSE helpers (MergeFragment, Redirect)
- [ ] Development server with file watching and hot reload
- [ ] Test factories generated from schema
- [ ] Soft delete: query scoping, partial unique indexes, WithTrashed/OnlyTrashed, Restore
- [ ] Generated Huma input/output structs from schema for OpenAPI 3.1
- [ ] Generated Huma route registration calling action layer
- [ ] OpenAPI 3.1 docs served at /api/docs
- [ ] Dual pagination: cursor-based for API, offset-based for HTML
- [ ] API auth: bearer tokens, API keys
- [ ] CORS configuration and rate limiting middleware
- [ ] Multi-tenancy with header/subdomain/path strategies and automatic query scoping + RLS
- [ ] Session-based email/password auth and OAuth2 (Google, GitHub)
- [ ] CRUD-level and field-level permissions from schema
- [ ] Audit logging with JSONB change diffs
- [ ] SSE connection management (limits, backpressure, graceful shutdown)
- [ ] LISTEN/NOTIFY hub with single-connection fan-out
- [ ] River integration: schema-triggered jobs, transactional enqueueing
- [ ] OpenTelemetry: traces, metrics, structured logging
- [ ] CLI commands: init, generate, migrate, dev, build, deploy, db, routes, openapi export
- [ ] Zero npm — Go binaries + Tailwind standalone CLI
- [ ] Single production binary with embedded templates, static assets, migrations

### Out of Scope

- Real-time collaboration primitives — future ecosystem phase
- Admin panel generator — future ecosystem phase
- Webhook system (outbound) — future ecosystem phase
- Full-text search (PostgreSQL tsvector) — future ecosystem phase
- Component library (DatastarUI-style) — future ecosystem phase
- Mobile app support — web-first framework
- Non-PostgreSQL database support — PostgreSQL is the platform

## Context

- **Ecosystem gap:** Go has excellent building blocks (net/http, pgx, Templ, Datastar, SQLC, Huma, River) but no cohesive framework integrating them for productive SaaS development
- **Prior art (Andurel):** Demonstrated the right stack (Go + Templ + Datastar + SQLC + PostgreSQL + River) but has gaps: Echo dependency, no schema-as-source-of-truth, no form handling, no pagination/filtering, no multi-tenancy, no OpenAPI, npm-dependent asset pipeline
- **Bootstrapping constraint:** Schema packages must have zero imports from gen/ — the CLI parses schemas using go/ast, not compile-and-execute
- **Key technology choices already made:** Huma for OpenAPI (not custom generation), Atlas for migrations (not custom diffing), Bob for dynamic queries (alongside SQLC for complex SQL), Datastar for hypermedia (not HTMX)

## Constraints

- **Tech stack**: Go 1.23+, net/http stdlib, PostgreSQL — no framework lock-in, single-DB philosophy
- **Zero npm**: All tooling must be Go binaries or standalone CLIs (Tailwind standalone) — no Node.js dependency
- **Bootstrapping**: Schema files must be parseable by go/ast without gen/ existing — no circular dependencies
- **Binary size**: Production binary < 30MB with embedded assets
- **Startup time**: < 100ms cold start
- **Performance**: 10k req/s target throughput

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Huma for OpenAPI (not custom generation) | Best OpenAPI 3.1 from code in Go; router-agnostic; avoids months of spec compliance work | — Pending |
| Atlas for migrations (not custom diffing) | Declarative schema diffing is a multi-year effort; Atlas handles edge cases we never will | — Pending |
| Bob for dynamic queries (alongside SQLC) | SQLC can't express dynamic filter/sort/pagination combinations without combinatorial explosion | — Pending |
| Datastar over HTMX | SSE-native maps to Go goroutines; built-in signals; ~15KB vs ~45KB; official Go SDK | — Pending |
| go/ast parsing (not compile-and-execute) | Solves bootstrapping — schemas parseable before gen/ exists; same approach as Ent/entc | — Pending |
| Action layer shared by HTML + API | Prevents business logic duplication between Datastar handlers and Huma handlers | — Pending |
| Scaffolded-once views (resources/) vs always-regenerated (gen/) | Clear ownership boundary — developers customize views without fighting the generator | — Pending |
| PostgreSQL for everything (data, jobs, sessions, caching) | Single dependency; transactional job enqueueing; RLS for tenancy; simplifies deployment | — Pending |

---
*Last updated: 2026-02-16 after initialization*
