# Phase 8: Background Jobs & Production Readiness - Context

**Gathered:** 2026-02-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Production-ready binary with background job processing (River), real-time SSE events (PostgreSQL LISTEN/NOTIFY), build/deploy packaging, and observability (OpenTelemetry, Prometheus, slog). Developer gets a single `forge build` command producing a deployable binary with all operational infrastructure wired in.

</domain>

<decisions>
## Implementation Decisions

### Job hooks & lifecycle
- Separate job registry: schema hooks reference jobs by `Kind` string (e.g., `{Kind: "notify_new_product", Queue: "notifications"}`), actual workers defined separately in `resources/<name>/jobs.go`
- This fits the AST parsing constraint — schema files use string literals, not Go functions; job workers need real dependencies (DB pool, mailer) that don't exist at schema definition time
- Retry policy is configurable per-job with sensible defaults (e.g., 3 retries with exponential backoff)
- Exhausted jobs: discard + slog warning + optional OnDiscard callback (configurable in forge.toml or per-job)
- Scaffold job workers into `resources/<name>/jobs.go` when schema uses Hooks — scaffold-once pattern consistent with handler scaffolding

### Real-time event delivery
- Application-level NOTIFY: action layer calls pg_notify() after successful writes — explicit control over what events fire and payload shape
- SSE events carry Datastar merge fragments (rendered Templ HTML) — consistent with existing Phase 6 Datastar pattern
- SSE connection limits configurable in forge.toml: global max connections and per-user max (e.g., 10,000 global / 5 per user), exceeding returns 429
- Fresh state on reconnect — client reconnects and gets current state; no last-event-id replay complexity

### Build & deployment
- Minimal embed: migrations and static assets only — Templ templates compile into Go code (already in binary), no need to embed template sources
- `forge build` runs the full pipeline: generate → templ generate → tailwind build → go build — one command from clean state to production binary
- Dockerfile generation via `forge deploy` — universal container target, works with any platform (fly.io, Railway, AWS ECS, etc.)
- forge.toml as defaults, env vars win — every config key has a corresponding env var (e.g., FORGE_DATABASE_URL); env vars override forge.toml values (12-factor)
- Build info via ldflags: git tag (version), commit SHA, and build date injected at compile time — available via `forge version` and health check endpoint

### Observability defaults
- OpenTelemetry auto-instruments: inbound HTTP requests, outgoing DB queries, and River job execution — each job run is a span linked to originating request trace if available
- Separate admin port (e.g., :9090) for /metrics, /healthz, /debug/pprof — operational endpoints stay off the public interface
- slog: JSON handler in production (for log aggregators), text handler with colors in development
- Default Prometheus metrics: HTTP request duration/count by route, DB query duration/count, job execution duration/count/failures, SSE connection count/events sent/connection duration — plus standard Go runtime metrics

### Claude's Discretion
- Exact River client configuration and queue defaults
- LISTEN/NOTIFY channel naming convention
- Dockerfile base image and multi-stage build structure
- OTEL exporter configuration (stdout vs OTLP endpoint)
- Admin port number and health check response shape
- Log level defaults and level-change mechanism

</decisions>

<specifics>
## Specific Ideas

- PRD Section 14.1/14.2 defines the job pattern: `schema.Hooks{AfterCreate: []schema.JobRef{...}}` with workers in separate files — follow this shape exactly
- Job workers need real dependencies (DB pool, mailer, etc.) so they live in resources/ not in schema — schema is AST-parsed, workers are runtime code

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-background-jobs-production-readiness*
*Context gathered: 2026-02-19*
