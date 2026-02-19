---
status: complete
phase: 08-background-jobs-production-readiness
source: 08-01-SUMMARY.md, 08-02-SUMMARY.md, 08-03-SUMMARY.md, 08-04-SUMMARY.md, 08-05-SUMMARY.md
started: 2026-02-19T22:00:00Z
updated: 2026-02-19T22:05:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Project compiles and passes vet
expected: `go build ./...` and `go vet ./...` both complete with zero errors. All new packages (observe, sse, notify, jobs) compile without issues.
result: pass

### 2. All existing tests pass
expected: `go test ./...` completes with all tests passing (PASS). No test failures or panics.
result: pass

### 3. forge build command exists
expected: Running `go run ./cmd/forge build --help` shows usage text describing the production build pipeline (generate, templ, tailwind, go build steps). Flags for output path visible.
result: pass

### 4. forge deploy command exists
expected: Running `go run ./cmd/forge deploy --help` shows usage text describing multi-stage Dockerfile generation with alpine runtime.
result: pass

### 5. Hooks schema DSL available
expected: `internal/schema/hooks.go` exports WithHooks(), JobRef, and Hooks types. A schema definition using `schema.WithHooks(schema.Hooks{AfterCreate: []schema.JobRef{{Kind: "SendWelcome", Queue: "default"}}})` compiles as a valid SchemaItem.
result: pass

### 6. 12-factor config env overrides
expected: `config.Load()` calls `ApplyEnvOverrides()` which maps FORGE_* env vars (FORGE_JOBS_ENABLED, FORGE_SSE_MAX_CONNECTIONS, FORGE_OBSERVE_LOG_LEVEL, FORGE_ADMIN_PORT, etc.) over TOML values. FORGE_ENV=production forces LogFormat=json.
result: pass

### 7. Observability stack (OTel, Prometheus, admin server)
expected: `internal/observe` package exports: NewHandler() (slog factory), Setup() (OTel bootstrap with TracerProvider + MeterProvider), NewHTTPHandler() (HTTP tracing), PGXTracer() (DB tracing), StartAdminServer() (serves /metrics, /healthz, /debug/pprof/ on port 9090).
result: pass

### 8. SSE connection limiter and LISTEN/NOTIFY hub
expected: `internal/sse/limiter.go` exports SSELimiter with Acquire/release pattern enforcing global + per-user caps. `internal/notify/hub.go` exports NotifyHub interface and PostgresHub using pgxlisten for single-connection fan-out with non-blocking backpressure.
result: pass

### 9. River background jobs integration
expected: `internal/jobs/client.go` exports NewRiverClient (with otelriver middleware), RunRiverMigrations. `actions.go.tmpl` generates transactional InsertTx inside pgx.BeginFunc for resources with Hooks. `scaffold_jobs.go.tmpl` generates per-hook Worker stubs.
result: pass

## Summary

total: 9
passed: 9
issues: 0
pending: 0
skipped: 0

## Gaps

[none yet]
