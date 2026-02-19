# Phase 8: Background Jobs & Production Readiness - Research

**Researched:** 2026-02-19
**Domain:** River background jobs, PostgreSQL LISTEN/NOTIFY SSE, production binary build, OpenTelemetry observability, 12-factor config
**Confidence:** HIGH (core stack verified via official docs and pkg.go.dev)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Job hooks & lifecycle
- Separate job registry: schema hooks reference jobs by `Kind` string (e.g., `{Kind: "notify_new_product", Queue: "notifications"}`), actual workers defined separately in `resources/<name>/jobs.go`
- This fits the AST parsing constraint — schema files use string literals, not Go functions; job workers need real dependencies (DB pool, mailer) that don't exist at schema definition time
- Retry policy is configurable per-job with sensible defaults (e.g., 3 retries with exponential backoff)
- Exhausted jobs: discard + slog warning + optional OnDiscard callback (configurable in forge.toml or per-job)
- Scaffold job workers into `resources/<name>/jobs.go` when schema uses Hooks — scaffold-once pattern consistent with handler scaffolding

#### Real-time event delivery
- Application-level NOTIFY: action layer calls pg_notify() after successful writes — explicit control over what events fire and payload shape
- SSE events carry Datastar merge fragments (rendered Templ HTML) — consistent with existing Phase 6 Datastar pattern
- SSE connection limits configurable in forge.toml: global max connections and per-user max (e.g., 10,000 global / 5 per user), exceeding returns 429
- Fresh state on reconnect — client reconnects and gets current state; no last-event-id replay complexity

#### Build & deployment
- Minimal embed: migrations and static assets only — Templ templates compile into Go code (already in binary), no need to embed template sources
- `forge build` runs the full pipeline: generate → templ generate → tailwind build → go build — one command from clean state to production binary
- Dockerfile generation via `forge deploy` — universal container target, works with any platform (fly.io, Railway, AWS ECS, etc.)
- forge.toml as defaults, env vars win — every config key has a corresponding env var (e.g., FORGE_DATABASE_URL); env vars override forge.toml values (12-factor)
- Build info via ldflags: git tag (version), commit SHA, and build date injected at compile time — available via `forge version` and health check endpoint

#### Observability defaults
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

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SCHEMA-09 | Developer can define schema Hooks (AfterCreate, AfterUpdate) referencing River job kinds | Hooks IR extension to ResourceOptionsIR; parser extracts JobRef structs with Kind+Queue strings; generator emits InsertTx calls in actions template |
| JOBS-01 | River client is integrated as first-class citizen with transactional enqueueing | River v0.x with riverpgxv5 driver; `river.NewClient(riverpgxv5.New(pool), &river.Config{})` + `client.InsertTx(ctx, tx, args, opts)` pattern verified |
| JOBS-02 | Schema-defined AfterCreate/AfterUpdate hooks automatically enqueue River jobs in the same DB transaction | Generated actions.go.tmpl wraps insert/update in pgx.Tx; calls InsertTx for each JobRef in Hooks.AfterCreate/AfterUpdate within same transaction |
| JOBS-03 | Jobs carry TenantID explicitly for scoped queries in workers | JobRef args struct scaffolded with TenantID field; generator emits TenantFromContext(ctx) assignment into enqueue call |
| JOBS-04 | Job queues are configurable via forge.toml (queue names, concurrency limits) | forge.toml [jobs] section with queues map; config.JobsConfig.Queues map[string]int; River config reads from this map |
| SSE-01 | Global SSE connection limiter caps concurrent connections per process (default: 5000) | SSELimiter struct with atomic.Int64 counter; PRD Section 11.3 design verified; 429 on exceed |
| SSE-02 | Per-user SSE connection limit prevents single-user exhaustion (default: 10) | sync.Map keyed by userID; per-user atomic counter; checked in Acquire() |
| SSE-03 | On server shutdown, SSE connections receive close event and drain gracefully | Context cancellation propagated to SSE handler goroutines; http.Server.Shutdown waits for active connections |
| SSE-04 | Single shared PostgreSQL LISTEN connection fans out events to subscribers via Go channels | pgxlisten package (jackc/pgxlisten) handles reconnection; NotifyHub owns one conn; fan-out to buffered Go channels per subscriber |
| SSE-05 | Backpressure: full subscriber channel buffers drop events and send refresh signal | Non-blocking select with default case in fan-out; dropped event → send "refresh" event to client |
| SSE-06 | NotifyHub interface allows swapping to Redis/NATS for connection pooler deployments | Interface definition with Subscribe/Publish; default PostgreSQL impl; swappable |
| CLI-03 | forge build compiles a single production binary with embedded templates, static assets, and migrations | forge build command: generate → templ generate → tailwind build → go build -trimpath -ldflags="-s -w -X..." |
| CLI-07 | forge deploy builds and packages for deployment | forge deploy command generates Dockerfile in project root; runs forge build; docker build if docker available |
| DEPLOY-01 | All configuration via forge.toml, overridable by environment variables (12-factor) | Config loader: reads TOML then overlays env vars by field mapping; FORGE_ prefix convention |
| DEPLOY-02 | Production binary is a single file < 30MB | go build -trimpath -ldflags="-s -w" proven to reduce size significantly; static assets + migrations embedded via //go:embed |
| DEPLOY-03 | Server starts in < 100ms cold start | No JIT, no reflection-heavy init; Go binary startup is native; pgx pool lazy-connects; River client starts async |
| OTEL-01 | OpenTelemetry traces on HTTP requests and database queries | otelhttp.NewHandler() for HTTP; otelpgx.NewTracer() on pgxpool.Config.ConnConfig.Tracer for DB |
| OTEL-02 | Prometheus-compatible metrics endpoint | prometheus/client_golang promhttp.Handler(); served on admin port :9090/metrics; OTel metrics bridge optional |
| OTEL-03 | Structured logging via slog with JSON format in production | log/slog JSONHandler in prod; TextHandler in dev; env var FORGE_ENV=production switches; otelslog bridge for trace correlation |
</phase_requirements>

---

## Summary

Phase 8 adds the four pillars that make Forge production-ready: background job processing (River), real-time SSE event delivery (PostgreSQL LISTEN/NOTIFY), production binary packaging (forge build / forge deploy), and observability (OpenTelemetry, Prometheus, structured slog). These features are largely additive — they extend the existing architecture rather than replacing it.

The River integration requires two code-generation changes: (1) extending the schema DSL and parser IR to support `Hooks` with `JobRef` references, and (2) updating the `actions.go.tmpl` to emit transactional InsertTx calls inside Create/Update database transactions. The generated actions already use `pgx.Tx` transactions for soft-delete operations, so the hooks integration is a natural extension. Worker scaffolding follows the existing `forge generate resource` scaffold-once pattern.

The LISTEN/NOTIFY hub requires a new internal package (`internal/notify/`) owning a single dedicated pgx connection (via `jackc/pgxlisten`) with fan-out to per-subscriber buffered channels. The SSE limiter is a straightforward atomic counter with `sync.Map` for per-user tracking. The `forge build` and `forge deploy` commands are new CLI additions orchestrating existing tools. Observability wiring uses three mature libraries: `otelhttp` for HTTP, `otelpgx` for database, and `otelriver` (via `rivercontrib`) for jobs.

**Primary recommendation:** Decompose into five discrete work streams: (1) schema/parser/generator for Hooks, (2) River client integration + worker scaffolding, (3) SSE limiter + LISTEN/NOTIFY hub, (4) forge build/deploy CLI + embed + ldflags, (5) OTel + Prometheus + slog + admin server. Each stream is independently testable.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/riverqueue/river` | v0.x (latest) | Background job processing with transactional enqueueing | Project's chosen job queue; PostgreSQL-backed; generics-based; production-proven |
| `github.com/riverqueue/river/riverpgxv5` | (bundled) | River driver for jackc/pgx v5 | Required adapter for pgx; `riverpgxv5.New(pool)` is the standard wiring |
| `github.com/riverqueue/river/rivermigrate` | (bundled) | Programmatic River schema migration | Allows running River migrations embedded in forge boot sequence |
| `github.com/jackc/pgxlisten` | v0.x | Higher-level LISTEN/NOTIFY on pgx with auto-reconnect | Built by pgx author; handles reconnection natively; avoids raw WaitForNotification complexity |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | v0.x | HTTP tracing middleware for net/http | Official OTel contrib; wraps `http.Handler`; zero-signature-change |
| `github.com/exaring/otelpgx` | v0.x | OpenTelemetry tracing for pgx v5 database queries | Plugs into `pgxpool.Config.ConnConfig.Tracer`; spans per query |
| `github.com/riverqueue/rivercontrib/otelriver` | v0.x | OpenTelemetry middleware for River job execution | Official River OTel integration; traces inserts and work; links spans to queues/kinds |
| `github.com/prometheus/client_golang` | v1.x | Prometheus metrics exposition | Industry standard; `promhttp.Handler()` for /metrics endpoint |
| `go.opentelemetry.io/otel/exporters/prometheus` | v0.x | OTel metrics → Prometheus bridge | Exposes OTel metrics as Prometheus metrics without dual-registration |
| `go.opentelemetry.io/contrib/bridges/otelslog` | v0.x | slog handler that injects OTel trace context | Correlates log entries with active spans; production logging pattern |
| `log/slog` | stdlib (Go 1.21+) | Structured logging | Standard library; JSONHandler for prod, TextHandler for dev |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `embed` | stdlib | Embed migrations and static assets into binary | `//go:embed migrations/*.sql static/*` directives in the generated project's main package |
| `net/http/pprof` | stdlib | CPU/memory profiling endpoints | Registered on admin mux only; never on public port |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | v0.x | OTLP trace exporter (HTTP transport) | When `FORGE_OTEL_ENDPOINT` is set in production |
| `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` | v0.x | Stdout trace exporter | Development/debugging only; not for production |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `pgxlisten` | Raw `pgx.Conn.WaitForNotification` | pgxlisten adds auto-reconnect which is essential in production; raw approach requires manual reconnect goroutine |
| `otelpgx` (pgx v5) | `XSAM/otelsql` (database/sql) | Project uses pgx directly, not database/sql; otelpgx is the native fit |
| Prometheus client directly | OTel metrics with Prometheus exporter | OTel metrics bridge avoids dual-registration; consistent with OTel-first approach |
| goreleaser | Custom `forge build` | Project has custom pipeline (generate → templ → tailwind → go build); goreleaser is release tooling, not build pipeline |

**Installation:**
```bash
go get github.com/riverqueue/river github.com/riverqueue/river/riverpgxv5 github.com/riverqueue/river/rivermigrate
go get github.com/jackc/pgxlisten
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go get github.com/exaring/otelpgx
go get github.com/prometheus/client_golang/prometheus/promhttp
go get go.opentelemetry.io/otel/exporters/prometheus
go get go.opentelemetry.io/contrib/bridges/otelslog
# Optional — only needed if users opt into OTLP export
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
```

---

## Architecture Patterns

### Recommended Project Structure for New Phase 8 Code

```
internal/
├── jobs/
│   └── client.go           # River client setup, queue config, worker registration bootstrap
├── notify/
│   ├── hub.go              # NotifyHub interface + PostgresHub implementation (pgxlisten)
│   └── subscription.go     # Subscription type; event fan-out to buffered channels
├── sse/
│   ├── limiter.go          # SSELimiter with global + per-user atomic counters
│   └── handler.go          # SSE connection lifecycle; integrates with NotifyHub
├── observe/
│   ├── otel.go             # OTel SDK bootstrap: TracerProvider, MeterProvider, LoggerProvider
│   ├── slog.go             # slog handler factory (JSON prod / Text dev)
│   └── admin.go            # Admin HTTP server (:9090) with /metrics, /healthz, /debug/pprof
└── cli/
    ├── build.go            # forge build command
    └── deploy.go           # forge deploy command

internal/config/
└── config.go               # Extended with JobsConfig, SSEConfig, ObserveConfig, AdminConfig

internal/generator/templates/
├── actions.go.tmpl         # Extended: InsertTx calls for Hooks.AfterCreate/AfterUpdate
└── scaffold_jobs.go.tmpl   # NEW: scaffold-once jobs.go for resources with Hooks

internal/parser/
└── ir.go                   # Extended: JobRefIR, HooksIR on ResourceOptionsIR

internal/schema/
└── hooks.go                # NEW: Hooks type + JobRef type for schema DSL
```

---

### Pattern 1: River Transactional Enqueueing

**What:** Job is inserted in same pgx transaction as the record mutation — job is guaranteed to appear if transaction commits and vanish if it rolls back.

**When to use:** Always for schema-hook-triggered jobs. Never use `Insert` (non-transactional) for hook jobs.

```go
// Source: https://pkg.go.dev/github.com/riverqueue/river
// Source: https://riverqueue.com/docs/transactional-enqueueing

// Generated in actions.go.tmpl when Hooks.AfterCreate has entries:
func (a *DefaultProductActions) Create(ctx context.Context, input models.ProductCreate) (*models.Product, error) {
    // ... validation ...

    var result *models.Product
    err := forge.Transaction(ctx, a.DB, func(tx pgx.Tx) error {
        // 1. Insert the record
        var insertErr error
        result, insertErr = insertProduct(ctx, tx, input)
        if insertErr != nil {
            return insertErr
        }

        // 2. Enqueue AfterCreate jobs in same transaction (JOBS-02)
        _, err := a.River.InsertTx(ctx, tx, NotifyNewProductArgs{
            ProductID: result.ID,
            TenantID:  forgeauth.TenantFromContext(ctx), // JOBS-03: explicit tenant
        }, &river.InsertOpts{
            Queue: "notifications",
        })
        return err
    })
    return result, err
}
```

**Key:** `a.River` is `*river.Client[pgx.Tx]` — the generic parameter must match the transaction type from the driver.

---

### Pattern 2: River Client Setup

**What:** Single River client per process; queues and concurrency from forge.toml; workers registered at startup.

```go
// Source: https://riverqueue.com/docs + pkg.go.dev/github.com/riverqueue/river
// internal/jobs/client.go

func NewRiverClient(pool *pgxpool.Pool, cfg config.JobsConfig) (*river.Client[pgx.Tx], error) {
    workers := river.NewWorkers()
    // Workers registered by each resource package at startup (not generated)
    // e.g.: river.AddWorker(workers, &product.NotifyNewProductWorker{DB: pool})

    queues := make(map[string]river.QueueConfig)
    for name, maxWorkers := range cfg.Queues {
        queues[name] = river.QueueConfig{MaxWorkers: maxWorkers}
    }
    if _, ok := queues[river.QueueDefault]; !ok {
        queues[river.QueueDefault] = river.QueueConfig{MaxWorkers: 100}
    }

    return river.NewClient(riverpgxv5.New(pool), &river.Config{
        Queues:       queues,
        Workers:      workers,
        ErrorHandler: &riverErrorHandler{}, // slog warning on exhaustion
        Middleware: []rivertype.Middleware{
            otelriver.NewMiddleware(nil), // OTEL-01: trace spans per job
        },
    })
}
```

---

### Pattern 3: River Migrations (Programmatic)

**What:** River needs its own tables (`river_jobs`, `river_queue`, etc.). Run via `rivermigrate` in forge's bootstrap or `forge migrate up`.

```go
// Source: https://riverqueue.com/docs/migrations + pkg.go.dev/github.com/riverqueue/river/rivermigrate

func RunRiverMigrations(ctx context.Context, pool *pgxpool.Pool) error {
    migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
    if err != nil {
        return fmt.Errorf("creating river migrator: %w", err)
    }
    _, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
    return err
}
```

**Decision:** River migrations run separately from Atlas migrations. Forge runs `river migrate-up` as part of `forge migrate up` command, before Atlas. River tables are in the same database but managed by River's own migration system.

---

### Pattern 4: NotifyHub (LISTEN/NOTIFY Fan-Out)

**What:** Single dedicated pgx connection for all LISTEN channels. Fan-out to per-subscriber buffered Go channels. Uses `pgxlisten` for auto-reconnect.

```go
// Source: https://pkg.go.dev/github.com/jackc/pgxlisten

// internal/notify/hub.go
type NotifyHub interface {
    Subscribe(channel string, tenantID uuid.UUID) *Subscription
    Publish(ctx context.Context, channel string, tenantID uuid.UUID, payload []byte) error
    Start(ctx context.Context) error
}

type Subscription struct {
    Events <-chan Event
    cancel func()
}

type Event struct {
    Channel string
    Payload []byte
}

type PostgresHub struct {
    connConfig *pgx.ConnConfig
    listener   *pgxlisten.Listener
    mu         sync.RWMutex
    subs       map[string][]*internalSub // channel:tenantID -> subscribers
}

func (h *PostgresHub) Subscribe(channel string, tenantID uuid.UUID) *Subscription {
    key := channel + ":" + tenantID.String()
    ch := make(chan Event, 32) // BUFFERED — SSE-05: backpressure via drop
    sub := &internalSub{ch: ch}

    h.mu.Lock()
    h.subs[key] = append(h.subs[key], sub)
    h.mu.Unlock()

    return &Subscription{
        Events: ch,
        cancel: func() { h.unsubscribe(key, sub) },
    }
}

func (h *PostgresHub) Start(ctx context.Context) error {
    h.listener = &pgxlisten.Listener{
        Connect: func(ctx context.Context) (*pgx.Conn, error) {
            return pgx.ConnectConfig(ctx, h.connConfig)
        },
        ReconnectDelay: 5 * time.Second,
    }
    // Register handler for each LISTEN channel
    h.listener.Handle("forge_events", pgxlisten.HandlerFunc(h.handleNotification))
    return h.listener.Listen(ctx) // blocks until ctx cancelled
}

func (h *PostgresHub) handleNotification(ctx context.Context, n *pgconn.Notification, conn *pgx.Conn) error {
    // Parse payload: {"channel":"products","tenant_id":"...","payload":{...}}
    var msg notifyMessage
    if err := json.Unmarshal([]byte(n.Payload), &msg); err != nil {
        slog.WarnContext(ctx, "invalid notify payload", "err", err)
        return nil
    }
    key := msg.Channel + ":" + msg.TenantID
    h.mu.RLock()
    subs := h.subs[key]
    h.mu.RUnlock()

    for _, sub := range subs {
        select {
        case sub.ch <- Event{Channel: msg.Channel, Payload: msg.RawPayload}:
        default:
            // SSE-05: buffer full — drop event, send refresh signal
            select {
            case sub.ch <- Event{Channel: "refresh", Payload: nil}:
            default:
            }
        }
    }
    return nil
}
```

**Channel naming convention (Claude's Discretion):** Single PostgreSQL LISTEN channel `forge_events`. Routing is done in payload JSON by `channel` + `tenant_id` fields. This keeps LISTEN count to 1 regardless of resource count.

---

### Pattern 5: SSE Limiter

**What:** Atomic counters for global and per-user SSE connection caps.

```go
// Source: PRD Section 11.3

// internal/sse/limiter.go
type SSELimiter struct {
    maxTotal   int64
    maxPerUser int64
    active     atomic.Int64
    perUser    sync.Map // string(userID) -> *atomic.Int64
}

func (l *SSELimiter) Acquire(userID string) (release func(), err error) {
    // Global check
    if l.active.Load() >= l.maxTotal {
        return nil, ErrTooManyConnections
    }

    // Per-user check
    userCount := l.getUserCounter(userID)
    if userCount.Load() >= l.maxPerUser {
        return nil, ErrTooManyConnectionsForUser
    }

    l.active.Add(1)
    userCount.Add(1)

    return func() {
        l.active.Add(-1)
        userCount.Add(-1)
    }, nil
}
```

**HTTP integration:** SSE handler calls `Acquire` before opening the stream; returns 429 with `Retry-After: 5` header on error.

---

### Pattern 6: Application-Level pg_notify (Action Layer)

**What:** After a successful write, the action layer calls `pg_notify` directly via the existing DB connection. No separate channel needed.

```go
// Generated in actions.go.tmpl when Hooks.AfterCreate has SSE trigger configured
// OR added manually for explicit real-time events

func notifyEvent(ctx context.Context, db DB, channel string, tenantID uuid.UUID, payload any) {
    data, _ := json.Marshal(notifyMessage{
        Channel:    channel,
        TenantID:   tenantID.String(),
        RawPayload: mustMarshal(payload),
    })
    // pg_notify payload max is 8000 bytes — keep payloads small (IDs, not full objects)
    _, _ = db.Exec(ctx, "SELECT pg_notify($1, $2)", "forge_events", string(data))
}
```

**Important constraint:** pg_notify payload is limited to 8000 bytes (PostgreSQL internal limit). Payloads should carry IDs only, not full serialized records.

---

### Pattern 7: forge build Command

**What:** Orchestrates the full production build pipeline in one command.

```go
// Source: PRD Section 18.1 + 21.1
// internal/cli/build.go

func runBuild(cmd *cobra.Command, args []string) error {
    // 1. Run forge generate (regenerate gen/)
    // 2. Run templ generate ./... (compile .templ → _templ.go)
    // 3. Run tailwind build (compile CSS)
    // 4. Run go build with production flags:
    //    go build -trimpath \
    //      -ldflags="-s -w \
    //        -X github.com/forge-framework/forge/internal/cli.Version=$(git describe --tags) \
    //        -X github.com/forge-framework/forge/internal/cli.Commit=$(git rev-parse --short HEAD) \
    //        -X github.com/forge-framework/forge/internal/cli.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    //      -o ./bin/<project-name> .
}
```

**Binary size target:** `-trimpath -ldflags="-s -w"` strips DWARF debug info and symbol table. Typical reduction: 20-35% from base. A Go binary with embedded migrations (~200KB) and static assets (~500KB) + application code should land well under 30MB.

---

### Pattern 8: OTel Bootstrap

**What:** Single `setupOTel(ctx)` function returning a shutdown func. Called at server startup before HTTP server starts.

```go
// Source: https://opentelemetry.io/docs/languages/go/getting-started/
// internal/observe/otel.go

func Setup(ctx context.Context, cfg ObserveConfig) (shutdown func(context.Context) error, err error) {
    var shutdowns []func(context.Context) error

    // Trace exporter: OTLP if endpoint configured, stdout in dev
    var traceExporter trace.SpanExporter
    if cfg.OTLPEndpoint != "" {
        traceExporter, err = otlptracehttp.New(ctx,
            otlptracehttp.WithEndpoint(cfg.OTLPEndpoint))
    } else {
        traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
    }

    // Metrics exporter: Prometheus (always available on /metrics)
    promExporter, err := prometheusexporter.New()
    meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))

    otel.SetTracerProvider(sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(traceExporter),
        sdktrace.WithResource(resource.NewWithAttributes(...)),
    ))
    otel.SetMeterProvider(meterProvider)

    // OTel slog bridge: logs carry trace IDs
    slogHandler := otelslog.NewHandler("forge")
    slog.SetDefault(slog.New(slogHandler))

    return func(ctx context.Context) error { /* shutdown all */ }, nil
}
```

---

### Pattern 9: Admin HTTP Server

**What:** Separate `http.Server` on port 9090 (default) serving operational endpoints. Not accessible from public traffic.

```go
// internal/observe/admin.go

func StartAdminServer(ctx context.Context, port int) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    mux.HandleFunc("/healthz", healthzHandler)
    mux.Handle("/debug/pprof/", http.DefaultServeMux) // pprof registers on DefaultServeMux

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    go srv.ListenAndServe()
    return srv
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "status":  "ok",
        "version": cli.Version,
        "commit":  cli.Commit,
        "built":   cli.Date,
    })
}
```

**Claude's Discretion:** Port 9090 chosen (standard Prometheus pattern); configurable via `FORGE_ADMIN_PORT`.

---

### Pattern 10: Config Env Var Override

**What:** Config loaded from forge.toml, then each field overridden by corresponding env var.

```go
// internal/config/config.go — extended

type Config struct {
    // ... existing fields ...
    Jobs    JobsConfig    `toml:"jobs"`
    SSE     SSEConfig     `toml:"sse"`
    Observe ObserveConfig `toml:"telemetry"`
    Admin   AdminConfig   `toml:"admin"`
}

type JobsConfig struct {
    Enabled bool              `toml:"enabled"`
    Queues  map[string]int    `toml:"queues"` // queue_name -> max_workers
}

type SSEConfig struct {
    MaxTotalConnections int `toml:"max_total_connections"` // default: 5000
    MaxPerUser          int `toml:"max_per_user"`          // default: 10
    BufferSize          int `toml:"buffer_size"`           // default: 32
}

type ObserveConfig struct {
    OTLPEndpoint string `toml:"otlp_endpoint"` // empty = stdout
    LogLevel     string `toml:"log_level"`     // "debug", "info", "warn", "error"
    LogFormat    string `toml:"log_format"`    // "json" (prod), "text" (dev)
}

type AdminConfig struct {
    Port int `toml:"port"` // default: 9090
}

// ApplyEnvOverrides overlays env vars onto the loaded config.
// Convention: FORGE_{SECTION}_{FIELD} e.g. FORGE_DATABASE_URL, FORGE_JOBS_ENABLED
func (c *Config) ApplyEnvOverrides() {
    if v := os.Getenv("FORGE_DATABASE_URL"); v != "" { c.Database.URL = v }
    if v := os.Getenv("FORGE_SERVER_PORT"); v != "" { /* parse int */ }
    if v := os.Getenv("FORGE_JOBS_ENABLED"); v != "" { /* parse bool */ }
    if v := os.Getenv("FORGE_SSE_MAX_TOTAL_CONNECTIONS"); v != "" { /* parse int */ }
    if v := os.Getenv("FORGE_SSE_MAX_PER_USER"); v != "" { /* parse int */ }
    if v := os.Getenv("FORGE_OTEL_ENDPOINT"); v != "" { c.Observe.OTLPEndpoint = v }
    if v := os.Getenv("FORGE_LOG_LEVEL"); v != "" { c.Observe.LogLevel = v }
    if v := os.Getenv("FORGE_LOG_FORMAT"); v != "" { c.Observe.LogFormat = v }
    if v := os.Getenv("FORGE_ADMIN_PORT"); v != "" { /* parse int */ }
    // FORGE_ENV=production triggers JSON logging even if forge.toml says "text"
    if v := os.Getenv("FORGE_ENV"); v == "production" { c.Observe.LogFormat = "json" }
}
```

---

### Pattern 11: Schema DSL Extension (Hooks)

**What:** `schema.Hooks` type added to the DSL; JobRef references jobs by kind string only. Schema files remain AST-parseable because no functions are referenced.

```go
// internal/schema/hooks.go (NEW)

// JobRef identifies a River job to enqueue when a lifecycle event fires.
// Kind must match the Kind() string returned by the job args type in resources/<name>/jobs.go.
type JobRef struct {
    Kind  string // e.g., "notify_new_product"
    Queue string // e.g., "notifications"; empty = river.QueueDefault
}

// Hooks defines lifecycle events that trigger River job enqueueing.
type Hooks struct {
    AfterCreate []JobRef
    AfterUpdate []JobRef
}

// HooksItem wraps Hooks for the SchemaItem interface.
type HooksItem struct {
    hooks Hooks
}

func (h *HooksItem) schemaItem() {}

// WithHooks adds lifecycle job hooks to a resource schema definition.
func WithHooks(h Hooks) *HooksItem {
    return &HooksItem{hooks: h}
}
```

**Schema usage (PRD Section 14.2):**
```go
var Resource = schema.Define("Product",
    schema.WithHooks(schema.Hooks{
        AfterCreate: []schema.JobRef{
            {Kind: "notify_new_product", Queue: "notifications"},
        },
        AfterUpdate: []schema.JobRef{
            {Kind: "reindex_product_search", Queue: "indexing"},
        },
    }),
    // ... other fields ...
)
```

**Parser IR extension:**
```go
// internal/parser/ir.go — additions to ResourceOptionsIR

type JobRefIR struct {
    Kind  string
    Queue string // empty string = default queue
}

type HooksIR struct {
    AfterCreate []JobRefIR
    AfterUpdate []JobRefIR
}

// In ResourceOptionsIR:
type ResourceOptionsIR struct {
    // ... existing fields ...
    Hooks HooksIR
}
```

---

### Pattern 12: Scaffolded Job Worker

**What:** When a schema has `Hooks`, scaffold a `jobs.go` file in `resources/<name>/` — scaffold-once, then developer owns it.

```go
// Template: internal/generator/templates/scaffold_jobs.go.tmpl
// Output: resources/{{snake .Resource.Name}}/jobs.go

package {{lower .Resource.Name}}

import (
    "context"

    "github.com/google/uuid"
    "github.com/riverqueue/river"
    "github.com/forge-framework/forge/internal/auth"
    "{{.ProjectModule}}/gen/models"
)

{{range .Resource.Options.Hooks.AfterCreate}}
// {{pascal .Kind}}Args carries the arguments for the {{.Kind}} job.
// TenantID is required for multi-tenant workers (JOBS-03).
type {{pascal .Kind}}Args struct {
    ResourceID uuid.UUID `json:"resource_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
}

func ({{pascal .Kind}}Args) Kind() string { return "{{.Kind}}" }

// {{pascal .Kind}}Worker processes the {{.Kind}} job.
// Add your dependencies to this struct (DB, mailer, etc.).
type {{pascal .Kind}}Worker struct {
    river.WorkerDefaults[{{pascal .Kind}}Args]
    // TODO: add dependencies here (e.g., DB *pgxpool.Pool)
}

func (w *{{pascal .Kind}}Worker) Work(ctx context.Context, job *river.Job[{{pascal .Kind}}Args]) error {
    // Restore tenant context for scoped queries (JOBS-03)
    ctx = forgeauth.WithTenant(ctx, job.Args.TenantID)

    // TODO: implement job logic
    _ = job.Args.ResourceID
    return nil
}
{{end}}
```

---

### Pattern 13: Dockerfile Generation (forge deploy)

**What:** `forge deploy` writes a multi-stage Dockerfile to the project root. Developer reviews and commits.

```dockerfile
# Claude's Discretion: golang:1.25-alpine builder, alpine:3.21 runtime

FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w \
    -X main.Version=$(git describe --tags --always) \
    -X main.Commit=$(git rev-parse --short HEAD) \
    -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/bin/app .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/bin/app /usr/local/bin/app
EXPOSE 8080 9090
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:9090/healthz || exit 1
CMD ["app", "serve"]
```

**Size characteristics:** golang:1.25-alpine builder discarded; only alpine:3.21 + binary in final image. Expected image size: 15-25MB.

---

### Pattern 14: slog Handler Factory

**What:** Create the slog handler based on `FORGE_ENV` / `FORGE_LOG_FORMAT` at startup.

```go
// internal/observe/slog.go
// Source: https://pkg.go.dev/log/slog

func NewHandler(cfg ObserveConfig) slog.Handler {
    opts := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}

    if cfg.LogFormat == "json" {
        return slog.NewJSONHandler(os.Stderr, opts)
    }
    return slog.NewTextHandler(os.Stderr, opts) // colored text in dev
}

func parseLevel(s string) slog.Level {
    switch s {
    case "debug": return slog.LevelDebug
    case "warn":  return slog.LevelWarn
    case "error": return slog.LevelError
    default:      return slog.LevelInfo
    }
}
```

**When OTel is enabled,** wrap with the otelslog bridge so log records carry trace/span IDs:
```go
// Source: https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog
import "go.opentelemetry.io/contrib/bridges/otelslog"

handler := otelslog.NewHandler("forge", otelslog.WithLoggerProvider(loggerProvider))
slog.SetDefault(slog.New(handler))
```

---

### Anti-Patterns to Avoid

- **Multiple LISTEN connections:** Do NOT open one pgx LISTEN connection per SSE client. One hub connection fans out to all subscribers. The `pgxlisten` package manages this single connection with reconnection.
- **Blocking fan-out:** Do NOT block the notify hub's dispatch goroutine waiting for slow SSE clients. Use non-blocking send with `select { case ch <- event: default: drop }`.
- **Transactional River insert with non-pgx conn:** `InsertTx` requires a `pgx.Tx` — do not pass a `database/sql` transaction. Use `riverpgxv5` driver consistently.
- **Embedding Templ source files:** Templ compiles to Go code; do NOT embed `.templ` files. Only embed `migrations/*.sql` and `static/` assets.
- **Prometheus on public port:** NEVER register `/metrics`, `/debug/pprof`, or `/healthz` on the public application port. Admin server on `:9090` only.
- **pg_notify with large payloads:** PostgreSQL NOTIFY payloads are limited to ~8000 bytes. Send record IDs only; clients fetch full state via SSE refresh.
- **River workers needing schema files:** Workers are in `resources/<name>/jobs.go` (runtime code with real deps). Schema files reference only kind strings (static, AST-parseable).
- **Non-transactional job enqueueing:** Using `river.Client.Insert()` (not `InsertTx`) for hook-triggered jobs means jobs can be enqueued even if the enclosing DB operation rolls back.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| LISTEN/NOTIFY reconnection | Custom reconnect loop around `pgx.Conn.WaitForNotification` | `jackc/pgxlisten` | pgx connection is not thread-safe; WaitForNotification cannot be interrupted without context cancellation; reconnect logic has subtle races |
| Job retry backoff | Custom retry scheduler | River's `DefaultClientRetryPolicy` (`attempts^4 + rand(±10%)`) or per-worker `NextRetry()` | River's retry state is stored in the jobs table; hand-rolling requires reimplementing job state machine |
| HTTP request tracing | Custom middleware recording timing | `otelhttp.NewHandler()` | otelhttp instruments standard W3C trace context propagation, B3, OTLP—doing this correctly by hand requires implementing multiple propagation standards |
| Database query tracing | Wrapping all pgx calls manually | `otelpgx.NewTracer()` on pool config | otelpgx hooks into pgx's tracer interface at the connection level—zero changes to query code |
| Prometheus /metrics handler | Custom text format serializer | `promhttp.Handler()` from `prometheus/client_golang` | Prometheus text format has edge cases; content negotiation with Prometheus server; client_golang handles all of it |
| Binary version injection | Reading a VERSION file at startup | ldflags `-X package.Var=value` at `go build` time | File reading adds I/O on startup; ldflags is the Go standard approach; already partially implemented in `internal/cli/version.go` |

**Key insight:** The operational infrastructure in this phase (observability, job queueing, real-time delivery) has well-defined library seams. The value Forge adds is wiring them together automatically — not reimplementing them.

---

## Common Pitfalls

### Pitfall 1: River Migration Tables Must Exist Before River Client Starts

**What goes wrong:** `river.NewClient()` panics or fails if River's PostgreSQL tables don't exist. This is a startup ordering issue: River migrations must run before the application server starts accepting requests.

**Why it happens:** River client validates its schema on startup. Atlas manages application migrations; River manages its own tables via `rivermigrate`.

**How to avoid:** Run `rivermigrate.Migrate(ctx, rivermigrate.DirectionUp, nil)` in `forge migrate up` BEFORE Atlas migrations, or at minimum as a separate required step. Document the startup sequence: River migrations → Atlas migrations → server start.

**Warning signs:** Startup crashes with "river: table river_jobs does not exist" or similar.

---

### Pitfall 2: SSE Connections Must Not Block HTTP Server Shutdown

**What goes wrong:** `http.Server.Shutdown()` waits for active connections to drain. SSE connections stay open indefinitely. Server never shuts down cleanly.

**Why it happens:** SSE is a long-lived HTTP connection from the server's perspective. Go's `http.Server.Shutdown()` calls `Close()` on idle connections but waits for active ones.

**How to avoid:** Use context cancellation propagated into every SSE handler goroutine. When server receives SIGTERM, cancel the top-level context; SSE handlers detect `ctx.Done()` and return, which closes the SSE stream cleanly. Set a `ShutdownTimeout` on the admin server and public server (e.g., 30 seconds).

**Warning signs:** `forge deploy` shows hanging shutdown in Kubernetes pod termination.

---

### Pitfall 3: pgxlisten Requires a Dedicated Connection, Not Pool Connection

**What goes wrong:** LISTEN/NOTIFY cannot be used on connections borrowed from a pgxpool — the pool may recycle the connection mid-listen.

**Why it happens:** pgxpool's connections are returned to the pool after each query. LISTEN state is connection-specific and lost when a connection is recycled.

**How to avoid:** `pgxlisten` handles this correctly by acquiring a dedicated `pgx.Conn` (not from pool) via its `Connect` callback. Never pass a pool connection to LISTEN. Always use pgxlisten's `Listener` struct.

**Warning signs:** Missed notifications, especially after connection timeouts or pool recycling.

---

### Pitfall 4: River Job Worker Registration Must Happen Before Client Start

**What goes wrong:** If a job kind arrives in the queue before its worker is registered, River logs an error and discards the job after max attempts.

**Why it happens:** Workers are registered in a `river.Workers` bundle BEFORE `river.NewClient()`. The code-generated approach (scaffold `jobs.go`, register manually in main.go) means the developer controls registration order.

**How to avoid:** Document the required registration pattern. The scaffolded `jobs.go` includes a `// TODO: register with river.AddWorker(workers, &NotifyNewProductWorker{})` comment pointing to main.go registration.

**Warning signs:** River logs "unhandled job kind: notify_new_product" at worker startup.

---

### Pitfall 5: Config Env Var Override Must Apply After TOML Load

**What goes wrong:** Env vars read at config struct initialization time (e.g., in struct tags or `init()`) may be overridden by TOML values loaded afterward.

**Why it happens:** If the config loader applies TOML first, then env vars, order is correct. If env vars are read during struct initialization, TOML load overwrites them.

**How to avoid:** `ApplyEnvOverrides()` MUST be called AFTER `config.Load(path)`. Verify in integration test: set `FORGE_DATABASE_URL`, load config with different toml value, assert env var wins.

**Warning signs:** Production ignores DATABASE_URL env var, connects with forge.toml URL instead.

---

### Pitfall 6: Existing Generator Tests Have Known Failures from Phase 7

**What goes wrong:** Phase 7 verification identified two test failures (`TestGenerateAtlasSchema_ProductTable`, `TestAtlasUniqueIndex`, `TestGenerateFactories_ProductSchema`, `TestGenerateFactories_MultipleResources`) from missing struct fields in `FactoryTemplateData` and `AtlasTemplateData`.

**Why it matters for Phase 8:** Phase 8's planner must decide whether to fix these regressions as part of Phase 8 planning or treat them as a prerequisite fix task at the start of Phase 8.

**Recommendation:** Fix the two known failures (atlas.go and factories.go struct field additions) as the FIRST task in Phase 8, before adding any new functionality. These are small, specific fixes documented precisely in `07-VERIFICATION.md`.

---

### Pitfall 7: pg_notify Payload Size Limit

**What goes wrong:** `pg_notify('forge_events', payload)` silently truncates or errors if payload exceeds PostgreSQL's 8000-byte internal limit.

**Why it happens:** PostgreSQL internal notification channels have an 8000-byte payload limit. Serialized full records routinely exceed this.

**How to avoid:** Design notify payloads as slim event envelopes: `{"channel":"products","tenant_id":"<uuid>","event":"created","id":"<uuid>"}`. SSE client handler fetches full state via the actions layer after receiving the notification.

**Warning signs:** Notifications silently fail for large records; SSE clients show stale state for records with large JSON fields.

---

## Code Examples

### River Client Wiring (Complete)

```go
// Source: https://riverqueue.com/docs + pkg.go.dev/github.com/riverqueue/river

import (
    "github.com/riverqueue/river"
    "github.com/riverqueue/river/riverpgxv5"
    "github.com/riverqueue/rivercontrib/otelriver"
    "github.com/riverqueue/river/rivertype"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

workers := river.NewWorkers()
// Register all workers BEFORE calling NewClient
river.AddWorker(workers, &product.NotifyNewProductWorker{DB: pool})

riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
    Queues: map[string]river.QueueConfig{
        river.QueueDefault:  {MaxWorkers: 100},
        "notifications":     {MaxWorkers: 25},
        "indexing":          {MaxWorkers: 10},
    },
    Workers:      workers,
    ErrorHandler: &riverErrorHandler{},
    Middleware: []rivertype.Middleware{
        otelriver.NewMiddleware(nil),
    },
})
if err != nil {
    log.Fatal(err)
}

// Start processing (non-blocking, starts goroutines)
if err := riverClient.Start(ctx); err != nil {
    log.Fatal(err)
}
```

---

### Transactional Enqueue (Generated Code Shape)

```go
// Source: https://pkg.go.dev/github.com/riverqueue/river — InsertTx signature

// In generated actions.go (after Bob query integration):
func (a *DefaultProductActions) Create(ctx context.Context, input models.ProductCreate) (*models.Product, error) {
    var created *models.Product
    err := pgx.BeginFunc(ctx, a.DB, func(tx pgx.Tx) error {
        // Insert via Bob (implemented in this phase or prior)
        var err error
        created, err = bobInsertProduct(ctx, tx, input)
        if err != nil {
            return err
        }

        // Enqueue AfterCreate hooks (JOBS-02)
        _, err = a.River.InsertTx(ctx, tx, NotifyNewProductArgs{
            ResourceID: created.ID,
            TenantID:   forgeauth.TenantFromContext(ctx),
        }, &river.InsertOpts{Queue: "notifications"})
        return err
    })
    return created, err
}
```

---

### OTel pgx Configuration

```go
// Source: https://github.com/exaring/otelpgx

cfg, err := pgxpool.ParseConfig(connStr)
if err != nil {
    return nil, err
}
cfg.ConnConfig.Tracer = otelpgx.NewTracer()

pool, err := pgxpool.NewWithConfig(ctx, cfg)
if err != nil {
    return nil, err
}

// Record pool stats (pool size, acquire count, etc.) as OTel metrics
if err := otelpgx.RecordStats(pool); err != nil {
    return nil, err
}
```

---

### OTel HTTP Middleware

```go
// Source: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp

import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// Wrap the entire mux — captures all routes
handler := otelhttp.NewHandler(mux, "forge-server",
    otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
)

server := &http.Server{
    Addr:    cfg.Server.Addr(),
    Handler: handler,
}
```

---

### embed.FS for Migrations

```go
// In the generated project's main package or a dedicated embed.go file

package main

import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed static
var staticFS embed.FS
```

**Note:** The `//go:embed` directive must be in the same package as the variable. For generated projects, this lives in the project root (user's `main` package), not in Forge's own source tree. `forge build` ensures `migrations/` and `static/` exist before running `go build`.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| One LISTEN connection per SSE client | Single hub connection with fan-out to Go channels | Always best practice; pgxlisten formalizes it | Critical for scaling beyond ~100 concurrent SSE users |
| Separate logging libraries (logrus, zap) | `log/slog` stdlib | Go 1.21 (Aug 2023) | No external dep; OTel bridge available |
| Manual OpenTelemetry wiring | OTel SDK with contrib auto-instrumentation | 2023-2024 | otelhttp, otelpgx, otelriver reduce boilerplate to one-liner per subsystem |
| Multiple exporters registered separately | Single OTel SDK with pluggable exporters | 2023 | Prometheus exporter bridges OTel metrics without dual-registration |

**Deprecated/outdated:**
- `lib/pq` LISTEN/NOTIFY: `jackc/pgx` is the current standard; `pgxlisten` is built on pgx v5
- `ddtrace`, `newrelic-go` direct integration: OpenTelemetry is now the vendor-neutral standard; River explicitly supports otelriver middleware

---

## Open Questions

1. **River client in generated project vs. forge framework itself**
   - What we know: River client needs to be wired in the generated project (user's main.go), not in Forge's own binary
   - What's unclear: Does Forge generate a `cmd/serve.go` or `main.go` bootstrap that wires River? Or does `forge init` scaffold it once?
   - Recommendation: `forge init` scaffolds a `main.go` with a `// TODO: register River workers` section; scaffold-once pattern mirrors handlers pattern. River client init code lives in a generated `app.go` or user's `main.go`.

2. **Bob query integration timing**
   - What we know: Phase 7 verification shows Create/Update actions have TODO placeholders where Bob queries will execute; transactional InsertTx requires a real `pgx.Tx` from a real DB write
   - What's unclear: Is Bob query integration planned for Phase 8 or a subsequent phase?
   - Recommendation: Phase 8 should integrate transactional enqueueing ALONGSIDE actual Bob INSERT — if Bob queries remain as TODOs, the `InsertTx` calls are unreachable. Either plan Bob integration in Phase 8 or accept that job hooks are scaffolded but not exercisable until Bob is wired.

3. **Phase 7 test failures — fix scope**
   - What we know: `TestGenerateAtlasSchema_ProductTable`, `TestAtlasUniqueIndex`, `TestGenerateFactories_ProductSchema`, `TestGenerateFactories_MultipleResources` all fail due to missing struct fields in `factories.go` and `atlas.go`
   - What's unclear: Is fixing these pre-existing failures in scope for Phase 8?
   - Recommendation: YES — fix as Task 1 of Phase 8. They are small, precisely documented fixes that unblock clean test baseline for Phase 8's own tests.

4. **NotifyHub registration pattern**
   - What we know: Different resources may want to subscribe to different channels; the hub needs to know which channels to LISTEN on at startup
   - What's unclear: Do we pre-LISTEN all channels on startup, or subscribe lazily as SSE clients connect?
   - Recommendation: Lazy subscription — when first SSE client subscribes to a channel, the hub issues LISTEN. pgxlisten's `Handle` method supports this pattern. On last unsubscribe, UNLISTEN (or just leave it — excess LISTEN channels are cheap in PostgreSQL).

---

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/riverqueue/river` — River client API, InsertTx signature, Config struct, worker interface, ErrorHandler interface
- `riverqueue.com/docs` — Getting started, transactional enqueueing, migrations, OpenTelemetry integration
- `riverqueue.com/docs/job-retries` — Default 25 max attempts, retry policy configuration
- `riverqueue.com/docs/middleware` — Middleware pattern, trace context injection via job metadata
- `pkg.go.dev/github.com/jackc/pgxlisten` — Listener struct, Handle(), HandlerFunc, ReconnectDelay
- `pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` — NewHandler() wrapping
- `github.com/exaring/otelpgx` — pgxpool.Config.ConnConfig.Tracer wiring, RecordStats()
- `riverqueue.com/docs/open-telemetry` — otelriver middleware, trace spans (river.insert_many, river.work), metrics
- `opentelemetry.io/docs/languages/go/getting-started/` — SDK initialization, shutdown pattern, exporter configuration
- `pkg.go.dev/log/slog` — JSONHandler, TextHandler, HandlerOptions, LevelVar
- `pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp` — Handler() for /metrics

### Secondary (MEDIUM confidence)
- `github.com/exaring/otelpgx` README — verified pgx v5 compatibility and `go 1.22+` requirement
- `riverqueue.com/docs/migrations` — rivermigrate programmatic API; Goose integration pattern
- WebSearch: Go multi-stage Dockerfile patterns — golang:alpine builder + alpine:3.x runtime; -trimpath -ldflags="-s -w" size reduction
- WebSearch: slog environment-based handler switching — confirmed JSON/text pattern

### Tertiary (LOW confidence)
- WebSearch: River exhausted job callbacks — ErrorHandler.HandleError exists but exact "discard + slog" pattern for exhaustion is inferred from docs; verify official source
- WebSearch: pg_notify 8000 byte limit — widely cited in PostgreSQL documentation; verify in postgres official docs before implementation

---

## Metadata

**Confidence breakdown:**
- River integration: HIGH — official docs + pkg.go.dev verified all key APIs
- pgxlisten LISTEN/NOTIFY: HIGH — official pkg.go.dev documentation verified
- OTel stack (otelhttp, otelpgx, otelriver): HIGH — official sources for all three
- Binary build (ldflags, embed): HIGH — Go stdlib (embed) + Go official docs (ldflags)
- Config env override pattern: HIGH — pattern is standard Go; implementation detail is straightforward
- SSE limiter: HIGH — design from PRD Section 11.3; atomic.Int64 + sync.Map is standard Go
- Dockerfile: MEDIUM — recommended structure from community best practices; base images subject to change

**Research date:** 2026-02-19
**Valid until:** 2026-05-19 (stable ecosystem; River APIs rarely break; OTel contrib evolves but not rapidly)
