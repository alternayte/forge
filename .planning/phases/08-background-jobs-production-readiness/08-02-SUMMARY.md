---
phase: 08-background-jobs-production-readiness
plan: 02
subsystem: infra
tags: [opentelemetry, prometheus, pprof, slog, otlp, otelhttp, otelpgx, config, 12-factor]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: config package and CLI version variables used by admin server healthz handler

provides:
  - "config.JobsConfig, SSEConfig, ObserveConfig, AdminConfig structs with TOML tags"
  - "config.ApplyEnvOverrides() method applying 12+ FORGE_* env vars over TOML values"
  - "observe.NewHandler() slog factory selecting JSON/text handler based on config"
  - "observe.Setup() OTel SDK bootstrap with TracerProvider and MeterProvider"
  - "observe.NewHTTPHandler() HTTP instrumentation wrapper (OTEL-01)"
  - "observe.PGXTracer() pgx tracer constructor (OTEL-01 DB traces)"
  - "observe.StartAdminServer() serving /metrics, /healthz, /debug/pprof/ on port 9090"

affects: [08-03, 08-04, 08-05, all future phases needing logging/tracing/metrics]

# Tech tracking
tech-stack:
  added:
    - go.opentelemetry.io/otel v1.40.0
    - go.opentelemetry.io/otel/sdk v1.40.0
    - go.opentelemetry.io/otel/sdk/metric v1.40.0
    - go.opentelemetry.io/otel/exporters/prometheus v0.62.0
    - go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.40.0
    - go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.40.0
    - go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.65.0
    - github.com/exaring/otelpgx v0.10.0
    - github.com/prometheus/client_golang v1.23.2
  patterns:
    - "12-factor config: env vars always win over TOML via ApplyEnvOverrides() called in Load()"
    - "OTel-Prometheus bridge: MeterProvider feeds admin /metrics endpoint without separate push"
    - "Dedicated admin mux: pprof registered only on admin server, never exposed on public handler"
    - "slog handler factory: format selected at startup from config, never changed at runtime"

key-files:
  created:
    - internal/observe/slog.go
    - internal/observe/otel.go
    - internal/observe/admin.go
  modified:
    - internal/config/config.go

key-decisions:
  - "OTel stdout exporter used in dev (cfg.OTLPEndpoint=empty); OTLP HTTP exporter in prod — avoids requiring a collector for local development"
  - "pprof registered explicitly on admin mux (not DefaultServeMux) — prevents accidental exposure on public port"
  - "Admin server started in goroutine, returns *http.Server — caller controls Shutdown for clean teardown"
  - "ApplyEnvOverrides silently ignores unparseable int/bool values after logging a warning — avoids crash on misconfiguration"
  - "FORGE_ENV=production forces LogFormat=json regardless of toml setting — production guarantee"

patterns-established:
  - "observe package: all observability infrastructure in one package, imported by main and test harness"
  - "config env override: FORGE_{SECTION}_{FIELD} convention for 12-factor compliance"

requirements-completed: [DEPLOY-01, OTEL-01, OTEL-02, OTEL-03]

# Metrics
duration: 6min
completed: 2026-02-19
---

# Phase 08 Plan 02: Observability Foundation Summary

**OTel SDK + Prometheus bridge + admin server on port 9090 with 12-factor config env overrides via FORGE_* variables**

## Performance

- **Duration:** 6 min
- **Started:** 2026-02-19T17:20:35Z
- **Completed:** 2026-02-19T17:26:09Z
- **Tasks:** 2
- **Files modified:** 4 (1 modified, 3 created)

## Accomplishments

- Extended `internal/config/config.go` with 4 new sections (Jobs, SSE, Observe, Admin) and `ApplyEnvOverrides()` covering 12+ `FORGE_*` env vars — `Load()` now applies overrides after TOML parse for 12-factor compliance
- Created `internal/observe/slog.go` with a `NewHandler()` factory that selects JSON or text slog handler based on config (FORGE_ENV=production forces JSON)
- Created `internal/observe/otel.go` bootstrapping TracerProvider (OTLP HTTP or stdout) + MeterProvider (Prometheus bridge), exporting `NewHTTPHandler()` and `PGXTracer()` for HTTP and DB tracing
- Created `internal/observe/admin.go` with `StartAdminServer()` serving `/metrics` (Prometheus), `/healthz` (version JSON via ldflags), and `/debug/pprof/` on a dedicated mux on port 9090

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend config with new sections and env var override** - `3bbad79` (feat)
2. **Task 2: Create observe package with slog, OTel bootstrap, and admin server** - `37b5410` (feat)

## Files Created/Modified

- `internal/config/config.go` - Added JobsConfig, SSEConfig, ObserveConfig, AdminConfig structs; ApplyEnvOverrides() method; updated Default() with new section defaults; Load() calls ApplyEnvOverrides
- `internal/observe/slog.go` - NewHandler() factory returning JSON or text slog.Handler with level from ObserveConfig
- `internal/observe/otel.go` - Setup() OTel SDK bootstrap; NewHTTPHandler() HTTP tracing wrapper; PGXTracer() pgx tracer
- `internal/observe/admin.go` - StartAdminServer() on cfg.Port (default 9090) with /metrics, /healthz, /debug/pprof/ on dedicated mux

## Decisions Made

- OTel stdout exporter used when `cfg.OTLPEndpoint` is empty (dev mode) — avoids requiring a collector locally; OTLP HTTP when endpoint configured (prod mode)
- pprof registered explicitly on admin mux via `pprof.Index` etc. (not `_ "net/http/pprof"` blank import) to prevent exposure on public server
- `ApplyEnvOverrides` silently discards unparseable values after `slog.Warn` — avoids crashing on misconfigured deployments
- `FORGE_ENV=production` forces `LogFormat="json"` regardless of TOML setting, ensuring production always emits structured logs
- Admin server runs in goroutine and returns `*http.Server` — caller (main) owns the Shutdown lifecycle

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed go.sum missing entries after oauth2 transitive upgrade**
- **Found during:** Task 2 (dependency installation)
- **Issue:** Installing OTel dependencies upgraded `golang.org/x/oauth2` from v0.27.0 to v0.34.0, which caused missing go.sum entries for `github.com/markbates/goth` and `cloud.google.com/go/compute/metadata`
- **Fix:** Ran `go get github.com/markbates/goth@v1.82.0` and `go get golang.org/x/oauth2/google@v0.34.0` to refresh go.sum entries
- **Files modified:** go.sum (already committed as part of module state)
- **Verification:** `go build ./...` and `go vet ./...` both pass
- **Committed in:** `37b5410` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking — go.sum consistency after transitive upgrade)
**Impact on plan:** Necessary housekeeping from OTel dependency cascade. No scope creep.

## Issues Encountered

None beyond the go.sum transitive upgrade issue documented above.

## User Setup Required

None - no external service configuration required.

## Self-Check: PASSED

All created files verified present. Both task commits found in git log:
- `3bbad79`: Task 1 (config extension)
- `37b5410`: Task 2 (observe package)

## Next Phase Readiness

- Config env overrides ready: any plan can now use `FORGE_*` env vars for 12-factor configuration
- OTel SDK bootstrapped: HTTP handlers can wrap with `observe.NewHTTPHandler()`, pgxpool can wire `observe.PGXTracer()`
- Admin server: callers invoke `observe.StartAdminServer(ctx, cfg.Admin)` in main to expose ops endpoints
- Plan 03 (SSE limiter) can use `cfg.SSE` directly for `MaxTotalConnections` and `MaxPerUser`
- Plan 04 (River integration) can use `cfg.Jobs` for `Enabled` and `Queues` config

---
*Phase: 08-background-jobs-production-readiness*
*Completed: 2026-02-19*
