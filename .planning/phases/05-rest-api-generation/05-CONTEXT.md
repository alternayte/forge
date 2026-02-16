# Phase 5: REST API Generation - Context

**Gathered:** 2026-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Generate a production-ready REST API from resource schemas using Huma integration. Includes OpenAPI 3.1 documentation, bearer token and API key authentication, CORS and rate limiting middleware, and CLI commands for route listing and spec export. All API handlers call the action layer from Phase 4 — no business logic in handlers. Session-based auth and OAuth are Phase 6 (HTML). Permissions and field visibility are Phase 7.

</domain>

<decisions>
## Implementation Decisions

### API URL design
- Path-based versioning: `/api/v1/<resource>`
- Flat URLs only — no nested resource routes (use query params like `?post_id=123` for filtering)
- Plural kebab-case resource names in URLs (`/api/v1/blog-posts`)
- Standard 5 CRUD endpoints per resource (List, Get, Create, Update, Delete) — no bulk operations

### Authentication & API keys
- Opaque bearer tokens stored in a database table (not JWTs) — revocable, queryable
- API keys are a separate table from bearer tokens — different use cases, own name/prefix/scopes/expiry
- Prefixed key format: `forg_live_abc123` / `forg_test_abc123` (Stripe/GitHub style)
- All API endpoints require authentication by default — public access is opt-in via schema annotation

### Response shape & errors
- All responses wrapped in envelope: `{"data": ...}` for single resources, `{"data": [...], "pagination": {...}}` for lists
- Pagination metadata inside the envelope: `{"next_cursor": "...", "has_more": true}`
- Error responses follow RFC 9457 (Problem Details for HTTP APIs)
- Validation errors include per-field detail: `"errors": [{"field": "email", "message": "invalid format"}, ...]`

### Docs & developer UX
- Scalar UI embedded directly in the app binary at `/api/docs` — no CDN dependency, matches single-binary philosophy
- OpenAPI spec designed for SDK generation: strict operationIds (listPosts, getPost), consistent naming, passes spectral linting
- `forge routes` output grouped by resource (not flat table)
- `forge openapi export` supports both JSON and YAML formats via `--format` flag

### Claude's Discretion
- Rate limiting strategy and configuration shape in forge.toml
- CORS default configuration
- Exact Huma middleware wiring order
- OpenAPI spec metadata (title, description, contact, license)
- How Scalar UI assets are embedded (go:embed vs inline)

</decisions>

<specifics>
## Specific Ideas

- API key prefix style inspired by Stripe (`forg_live_`, `forg_test_`) for easy identification in logs and config
- `forge routes` should feel like `rails routes` — grouped, scannable, quick to find what you need
- SDK-ready spec quality: should work with openapi-generator out of the box

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-rest-api-generation*
*Context gathered: 2026-02-16*
