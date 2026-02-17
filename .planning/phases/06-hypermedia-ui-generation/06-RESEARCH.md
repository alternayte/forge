# Phase 6: Hypermedia UI Generation - Research

**Researched:** 2026-02-17
**Domain:** Templ components, Datastar SSE, session auth, OAuth2, Tailwind CLI, HTTP testing, scaffold-once code generation
**Confidence:** HIGH

## Summary

Phase 6 adds the HTML layer to the forge framework. Three distinct concerns are intertwined: (1) scaffolded-once view files (Templ components + HTML handlers) written into `resources/`, (2) a session-based authentication system (email/password + OAuth2 via Google/GitHub, sessions stored in PostgreSQL), and (3) testing infrastructure (`forgetest` package with `NewTestDB`, `NewApp`, and Datastar-compatible HTTP test helpers).

The technology choices are well-locked by prior phases and requirements: Templ for type-safe HTML components (compiled Go code), Datastar Go SDK for SSE-driven interactivity (no JavaScript bundler needed), SCS v2 with pgxstore for session management in PostgreSQL (no Redis), `alexedwards/scs` satisfies AUTH-03 directly. OAuth2 is best handled by `markbates/goth` which supports Google and GitHub with a gorilla/sessions-compatible interface. Tailwind CSS standalone CLI binary already exists in the tool registry at v3.4.17 — Phase 6 needs to upgrade awareness to v4.x (different binary URL pattern) and integrate compilation into the dev watcher.

The key architectural insight for Phase 6: there are two generation modes. Generated-always code (in `gen/`) contains the API layer. Scaffolded-once code (in `resources/<name>/`) contains the HTML views and handlers. The `--diff` flag for `forge generate resource <name>` renders the scaffold into a temp buffer and diffs against the on-disk file — using `sergi/go-diff` or `golang.org/x/tools/internal/diff` for unified diff output. The `forge routes` command already separates `apiRoutes()` from `htmlRoutes()` by design (Phase 5 plan 04 explicitly structured this for Phase 6).

**Primary recommendation:** Use Templ v0.3.977 + `starfederation/datastar-go` v1.1.0 + `alexedwards/scs` v2.9.0 with pgxstore + `markbates/goth` v1.82.0 for the complete HTML+auth stack. Keep Tailwind at v3.4.17 (already in toolsync) for now but note v4 requires different binary naming (`tailwindcss-<platform>-<arch>` unchanged, but v4 has new input CSS syntax). Use `peterldowns/testdb` for isolated PostgreSQL test schemas.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/a-h/templ | v0.3.977 | HTML component templating | Compiles to Go code — type-safe, no runtime parsing, IDE support, native SSE fragment support via `PatchElementTempl` |
| github.com/starfederation/datastar-go | v1.1.0 | Datastar SSE Go SDK | Official dedicated SDK (moved from main repo), has `PatchElementTempl` for direct templ integration |
| github.com/alexedwards/scs/v2 | v2.9.0 | HTTP session management | OWASP-compliant server-side sessions, 19+ store backends, `LoadAndSave` chi-compatible middleware |
| github.com/alexedwards/scs/pgxstore | (part of scs) | PostgreSQL session store | Uses pgx driver (already in project), satisfies AUTH-03 (no Redis), auto-cleanup goroutine |
| github.com/markbates/goth | v1.82.0 | OAuth2 providers | 60+ providers including Google and GitHub, gorilla/sessions compatible, `gothic.Store` swap for custom stores |
| golang.org/x/crypto/bcrypt | stdlib-ext | Password hashing | Industry standard for password auth, `bcrypt.GenerateFromPassword` + `CompareHashAndPassword` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/peterldowns/testdb | latest | Isolated test PostgreSQL databases | `forgetest.NewTestDB` — clones template DB per test, atlas migrator built-in, ~10ms per test |
| github.com/sergi/go-diff/diffmatchpatch | latest | Diff between current and scaffolded files | `forge generate resource <name> --diff` shows what would change |
| net/http/httptest | stdlib | HTTP handler testing | `forgetest.NewApp` wraps httptest.Server for full HTTP integration tests |
| templ CLI (binary) | v0.3.977 | Compile .templ files to .go | Already in toolsync registry as `templ`; run `templ generate` during `forge generate` |
| Tailwind standalone CLI | v3.4.17 (current), v4.x available | Compile CSS | Already in toolsync as `tailwindcss`; run during `forge dev` via watcher |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| alexedwards/scs | gorilla/sessions | scs has simpler API, built-in PostgreSQL stores, OWASP compliance; gorilla/sessions requires more plumbing |
| markbates/goth | golang.org/x/oauth2 directly | Goth abstracts provider differences (token exchange, user fetching); raw oauth2 requires provider-specific code |
| starfederation/datastar-go | starfederation/datastar/sdk/go | datastar-go is the dedicated current repo; sdk/go path is the older main-repo location |
| peterldowns/testdb | Custom test helpers | testdb provides atlas migrator, template cloning (~10ms), auto-cleanup; custom takes weeks to get right |
| sergi/go-diff | golang.org/x/tools/internal/diff | sergi/go-diff is public API; x/tools/internal/diff is internal package, not safe to import |

**Installation:**
```bash
go get github.com/a-h/templ@v0.3.977
go get github.com/starfederation/datastar-go@v1.1.0
go get github.com/alexedwards/scs/v2@v2.9.0
go get github.com/alexedwards/scs/pgxstore
go get github.com/markbates/goth@v1.82.0
go get golang.org/x/crypto/bcrypt
go get github.com/peterldowns/testdb
go get github.com/sergi/go-diff/diffmatchpatch
```

## Architecture Patterns

### Recommended Project Structure

```
resources/
├── <name>/                      # Scaffolded-once (developer owns)
│   ├── <name>.schema.go         # Schema definition (existing from Phase 1)
│   ├── views/
│   │   ├── form.templ            # Form component (create/edit)
│   │   ├── list.templ            # List component with sort/filter/pagination
│   │   └── detail.templ          # Read-only detail component
│   ├── handlers.go               # HTML handlers calling action layer
│   └── hooks.go                  # Lifecycle hooks (optional)
gen/
├── html/                         # Generated-always (do not edit)
│   ├── primitives/               # FormField, TextInput, etc. (gen'd primitives library)
│   ├── sse/                      # MergeFragment, Redirect helpers
│   └── register_all.go           # HTML route registration
internal/
├── auth/
│   ├── session.go                # SCS session manager setup
│   ├── password.go               # bcrypt hash/verify
│   ├── oauth.go                  # Goth provider setup
│   └── middleware.go             # Auth middleware for HTML routes
├── forgetest/
│   ├── db.go                     # NewTestDB using peterldowns/testdb
│   ├── app.go                    # NewApp using httptest.Server
│   └── datastar.go               # PostDatastar helper for SSE form testing
cmd/forge/
└── generate_resource.go          # forge generate resource <name> [--diff]
```

### Pattern 1: Templ Component Definition
**What:** Components compile to Go functions; they receive typed parameters, not maps
**When to use:** Every scaffolded view (form, list, detail)
**Example:**
```go
// Source: templ.guide, pkg.go.dev/github.com/a-h/templ@v0.3.977
// resources/products/views/form.templ

package views

import "myapp/gen/models"

// ProductForm renders a Datastar-native create/edit form.
templ ProductForm(product *models.Product, errors map[string]string) {
    <form data-on:submit__prevent={ datastar.PostSSE("/products") }>
        <div id="product-form-errors">
            if len(errors) > 0 {
                for field, msg := range errors {
                    <p class="text-red-600" data-testid={ "error-" + field }>{ msg }</p>
                }
            }
        </div>
        @TextInput("name", "Name", product.Name, errors["name"])
        <button type="submit">Save</button>
    </form>
}
```

Templ generates this into `form_templ.go` (a Go function). The CLI binary `templ generate` must run before `go build`.

### Pattern 2: Datastar SSE Handler (HTML Handler calling Action Layer)
**What:** HTTP handler that reads signals, calls action, sends SSE fragment back
**When to use:** All HTML form submissions — replaces traditional POST-redirect-GET
**Example:**
```go
// Source: pkg.go.dev/github.com/starfederation/datastar-go@v1.1.0
// resources/products/handlers.go (scaffolded-once)

package products

import (
    "net/http"
    "myapp/gen/actions"
    "myapp/gen/models"
    "myapp/resources/products/views"
    datastar "github.com/starfederation/datastar-go/datastar"
)

func HandleCreate(acts actions.ProductActions) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Read Datastar signals from request body
        var signals struct {
            Name  string `json:"name"`
            Price string `json:"price"`
        }
        if err := datastar.ReadSignals(r, &signals); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Call action layer (same as API layer)
        product, err := acts.Create(r.Context(), models.ProductCreate{
            Name: signals.Name,
        })

        sse := datastar.NewSSE(w, r)

        if err != nil {
            // Re-render form with errors
            errs := extractFieldErrors(err)
            sse.PatchElementTempl(views.ProductForm(nil, errs))
            return
        }

        // Success: redirect to detail page
        sse.Redirectf("/products/%s", product.ID)
    }
}
```

### Pattern 3: SCS Session Manager Setup with pgxstore
**What:** PostgreSQL-backed server-side sessions; token in cookie, data in DB
**When to use:** Auth sessions for email/password and OAuth2 logins
**Example:**
```go
// Source: pkg.go.dev/github.com/alexedwards/scs/v2, pkg.go.dev/github.com/alexedwards/scs/pgxstore
package auth

import (
    "time"
    "github.com/alexedwards/scs/pgxstore"
    "github.com/alexedwards/scs/v2"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewSessionManager(pool *pgxpool.Pool) *scs.SessionManager {
    sessionManager := scs.New()
    sessionManager.Store = pgxstore.New(pool)  // PostgreSQL store via pgx
    sessionManager.Lifetime = 24 * time.Hour
    sessionManager.Cookie.HttpOnly = true
    sessionManager.Cookie.SameSite = http.SameSiteLaxMode
    sessionManager.Cookie.Secure = true  // Set false in dev
    return sessionManager
}

// Required sessions table (run via Atlas migration):
// CREATE TABLE sessions (
//     token TEXT PRIMARY KEY,
//     data BYTEA NOT NULL,
//     expiry TIMESTAMPTZ NOT NULL
// );
// CREATE INDEX sessions_expiry_idx ON sessions (expiry);

// Chi middleware integration:
// router.Use(sessionManager.LoadAndSave)
```

### Pattern 4: Goth OAuth2 with Custom PostgreSQL Session Store
**What:** goth uses gothic.Store (gorilla/sessions compatible) — swap to SCS-backed store
**When to use:** Google and GitHub OAuth2 flows
**Example:**
```go
// Source: pkg.go.dev/github.com/markbates/goth@v1.82.0
package auth

import (
    "github.com/markbates/goth"
    "github.com/markbates/goth/providers/github"
    "github.com/markbates/goth/providers/google"
    "github.com/markbates/gothic"
    "github.com/gorilla/sessions"
)

func SetupOAuth(cfg OAuthConfig) {
    // Register providers
    goth.UseProviders(
        google.New(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.CallbackURL+"/auth/google/callback"),
        github.New(cfg.GitHub.ClientID, cfg.GitHub.ClientSecret, cfg.CallbackURL+"/auth/github/callback"),
    )

    // gothic.Store is gorilla/sessions compatible
    // Swap to database-backed store (not cookie-only) to satisfy AUTH-03
    gothic.Store = sessions.NewCookieStore([]byte(cfg.SessionSecret))
    // NOTE: For full PostgreSQL storage of OAuth state, implement gorilla/sessions
    // compatible store backed by PostgreSQL, OR accept that OAuth temp state
    // (30 second window) is cookie-only while completed sessions go to SCS+pgxstore.
}

// OAuth routes for Chi router:
// router.Get("/auth/{provider}", gothic.BeginAuthHandler)
// router.Get("/auth/{provider}/callback", handleOAuthCallback)

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
    user, err := gothic.CompleteUserAuth(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // user.Email, user.Name, user.Provider available
    // Create or find user, put userID in SCS session
    sessionManager.Put(r.Context(), "userID", user.UserID)
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

### Pattern 5: Scaffold-Once Code Generation (resources/ vs gen/)
**What:** `forge generate resource <name>` writes to `resources/<name>/` once; `forge generate` rewrites `gen/` always
**When to use:** HTML view scaffolding — developer owns these files after creation
**Implementation:**
```go
// internal/generator/scaffold.go

// ScaffoldResource writes scaffold files to resources/<name>/ ONLY if they don't exist.
// Called by forge generate resource <name>.
func ScaffoldResource(resource parser.ResourceIR, projectRoot, projectModule string) error {
    resourceDir := filepath.Join(projectRoot, "resources", snake(resource.Name))

    // Check-before-write: skip if file already exists (developer may have edited it)
    viewsDir := filepath.Join(resourceDir, "views")
    formPath := filepath.Join(viewsDir, "form.templ")

    if _, err := os.Stat(formPath); err == nil {
        return fmt.Errorf("resource %s already scaffolded; use --diff to see changes", resource.Name)
    }

    // Write each scaffold file
    return writeScaffoldFiles(resource, resourceDir, projectModule)
}

// DiffResource renders scaffold into buffer and diffs against on-disk files.
// Called by forge generate resource <name> --diff.
func DiffResource(resource parser.ResourceIR, projectRoot, projectModule string) (string, error) {
    // Render into memory (not disk)
    current, err := os.ReadFile(filepath.Join(projectRoot, "resources", snake(resource.Name), "views", "form.templ"))
    if err != nil {
        return "", err
    }

    // Render what the scaffold would produce
    fresh, err := renderScaffoldToBuffer(resource, projectModule)
    if err != nil {
        return "", err
    }

    // Produce unified diff
    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(string(current), string(fresh), true)
    return dmp.DiffPrettyText(diffs), nil
}
```

### Pattern 6: forgetest.NewTestDB with Atlas Migrator
**What:** peterldowns/testdb creates isolated PostgreSQL schema per test; atlas migrator built-in
**When to use:** All tests that need a real database (integration/HTTP tests)
**Example:**
```go
// Source: pkg.go.dev/github.com/peterldowns/testdb
// internal/forgetest/db.go

package forgetest

import (
    "testing"
    "github.com/peterldowns/testdb"
    "github.com/peterldowns/testdb/migrators/atlasmigrator"
)

func NewTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    conf := testdb.Config{
        Host:     "localhost",
        Port:     "5432",
        User:     "postgres",
        Password: "postgres",
        Database: "postgres",
        Options:  "sslmode=disable",
    }

    // Use atlas migrator — matches the project's migration tooling
    migrator := atlasmigrator.New("../../gen/atlas/schema.hcl", atlasmigrator.Config{})

    db := testdb.New(t, conf, migrator)
    // db is *sql.DB; convert to pgxpool.Pool for project usage
    // Template creation: ~500ms (one-time), clone: ~10ms per test
    return wrapToPgxPool(t, db)
}
```

### Pattern 7: forgetest.NewApp and PostDatastar Helper
**What:** Full HTTP test server + helper for submitting Datastar SSE requests
**When to use:** TEST-04 requirement — HTTP testing with Datastar form submissions
**Example:**
```go
// internal/forgetest/app.go

package forgetest

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

// NewApp creates a full HTTP test server using the application's router.
func NewApp(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
    t.Helper()
    router := buildRouter(pool)  // same router as production
    srv := httptest.NewServer(router)
    t.Cleanup(srv.Close)
    return srv
}

// PostDatastar sends a Datastar SSE form submission.
// Datastar sends signals as JSON body; response is text/event-stream.
func PostDatastar(t *testing.T, srv *httptest.Server, path string, signals any) *http.Response {
    t.Helper()
    body, err := json.Marshal(signals)
    if err != nil {
        t.Fatalf("PostDatastar marshal: %v", err)
    }

    req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(body))
    if err != nil {
        t.Fatalf("PostDatastar request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "text/event-stream")  // Signal Datastar client

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("PostDatastar do: %v", err)
    }
    t.Cleanup(func() { resp.Body.Close() })
    return resp
}
```

### Pattern 8: Tailwind Compilation in Dev Watcher
**What:** Tailwind standalone CLI invoked as subprocess during `forge dev` and `forge generate`
**When to use:** CSS compilation — no npm, no node_modules
**Example:**
```go
// internal/watcher/dev.go (extend existing watcher)

// RunTailwind runs the Tailwind CLI to compile CSS.
// The binary is in .forge/bin/tailwindcss (downloaded by toolsync).
func RunTailwind(projectRoot string) error {
    binPath := filepath.Join(projectRoot, ".forge", "bin", "tailwindcss")
    cmd := exec.Command(binPath,
        "-i", filepath.Join(projectRoot, "resources", "css", "input.css"),
        "-o", filepath.Join(projectRoot, "public", "css", "output.css"),
    )
    cmd.Dir = projectRoot
    return cmd.Run()
}

// Tailwind input.css (scaffolded into resources/css/):
// @import "tailwindcss";  (v4 syntax)
// OR
// @tailwind base;          (v3 syntax - current registry version)
// @tailwind components;
// @tailwind utilities;
```

### Anti-Patterns to Avoid
- **Putting business logic in HTML handlers:** Handlers must call action layer, not DB directly. Same rule as API handlers.
- **Using CDN for Tailwind or Datastar JS:** Violates single-binary philosophy. Tailwind is compiled ahead of time. Datastar JS must be embedded in the base layout templ component.
- **Storing sessions in cookies only:** Violates AUTH-03. Use SCS + pgxstore.
- **Calling `templ generate` manually in templates or generators:** The Go generator (`forge generate`) must invoke `templ generate` as a subprocess on the generated `.templ` files so `.go` counterparts exist for compilation.
- **Writing scaffold files to `gen/`:** Scaffold-once files belong in `resources/<name>/`. The `gen/` directory is always regenerated and would overwrite developer changes.
- **Using regular string comparison for session tokens:** Use `crypto/subtle.ConstantTimeCompare` (same rule from Phase 5).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Session management | Custom cookie + DB session table | alexedwards/scs v2 + pgxstore | OWASP compliance, rotation, lifetime, cleanup goroutine; session security has many subtle failure modes |
| OAuth2 provider integration | Raw golang.org/x/oauth2 per provider | markbates/goth | Each provider (Google, GitHub) has different token exchange, scopes, user JSON shapes; Goth normalizes all into `goth.User` |
| PostgreSQL test isolation | Schema-per-test with manual setup | peterldowns/testdb | Template DB + clone pattern is ~10ms; manual setup is complex, brittle, slow (500ms+) per test |
| Diff output between files | Custom line-by-line comparison | sergi/go-diff/diffmatchpatch | Diff algorithm handles line moves, block changes; custom implementations miss edge cases |
| Password hashing | Custom salt + hash | golang.org/x/crypto/bcrypt | bcrypt handles salt internally, cost factor for future hardening; custom crypto is always wrong |
| Templ compilation | Parsing .templ files in Go | templ CLI binary (`templ generate`) | Templ has complex syntax handling (escaping, imports, context); the compiler is part of the library |
| Datastar SSE framing | Manual `data:` / `event:` SSE writing | starfederation/datastar-go SDK | SSE framing rules, connection keepalive, proper content-type headers; SDK handles all of it |

**Key insight:** The HTML layer has more library surface area than the API layer because it encompasses session management, OAuth flows, CSS compilation, and SSE framing — each with its own correctness landmines. Use maintained libraries.

## Common Pitfalls

### Pitfall 1: Running `templ generate` Before vs. After Go Generation
**What goes wrong:** `forge generate` writes `.templ` files, but they need `templ generate` run on them before `go build` can compile the project. If the watcher or generate command doesn't invoke `templ generate`, the project won't compile.
**Why it happens:** `.templ` files are not Go files — they need the templ compiler to produce `_templ.go` counterparts.
**How to avoid:** `forge generate` must invoke `templ generate ./...` as a subprocess after writing scaffold `.templ` files. The dev watcher must re-run `templ generate` when `.templ` files change.
**Warning signs:** `undefined: views.ProductForm` or `cannot find package` errors during compilation.

### Pitfall 2: Scaffold-Once Collision on Re-Run
**What goes wrong:** Running `forge generate resource product` twice overwrites developer-modified `form.templ`, losing their changes.
**Why it happens:** Generator writes files unconditionally.
**How to avoid:** Check if file exists before writing. If it exists, skip and print a message like "Already exists: resources/products/views/form.templ (use --diff to see what would change)". The `--diff` flag shows the diff without overwriting.
**Warning signs:** Developer complains their form customizations are gone after re-running generate.

### Pitfall 3: Datastar Signal Naming Conflicts
**What goes wrong:** Signals named with underscores prefix (`$_field`) are treated as local (client-only) and not sent to the backend. Field-level signals with naming like `form.name` require nested struct unmarshaling.
**Why it happens:** Datastar has a convention: `$_` prefix = local only.
**How to avoid:** Use flat or dot-namespaced signal names. For form fields use `form.fieldName` pattern (nested JSON object). Backend struct must match:
```go
var signals struct {
    Form struct {
        Name string `json:"name"`
    } `json:"form"`
}
datastar.ReadSignals(r, &signals)
```
**Warning signs:** Backend receives empty signals; fields read as zero values.

### Pitfall 4: Auth Middleware Applied to Wrong Routes
**What goes wrong:** Auth middleware accidentally applied to `GET /auth/login` or OAuth callback routes, causing infinite redirect loops or 401 on the login page itself.
**Why it happens:** Blanket `router.Use(authMiddleware)` applied to all routes.
**How to avoid:** Use Chi route groups to scope auth middleware:
```go
// Public routes (no auth required)
router.Group(func(r chi.Router) {
    r.Get("/auth/login", loginPage)
    r.Post("/auth/login", loginHandler)
    r.Get("/auth/{provider}", gothic.BeginAuthHandler)
    r.Get("/auth/{provider}/callback", oauthCallback)
})

// Protected routes (auth required)
router.Group(func(r chi.Router) {
    r.Use(requireAuth(sessionManager))  // Auth middleware here only
    r.Get("/products", productList)
    r.Post("/products", createProduct)
})
```
**Warning signs:** Login page returns 401; OAuth flow fails with infinite redirect.

### Pitfall 5: OAuth State Not Persisting in Session
**What goes wrong:** OAuth provider sends a `state` parameter for CSRF protection. If sessions are not configured correctly, the callback handler can't verify the state and the OAuth flow fails.
**Why it happens:** `gothic.Store` defaults to a cookie-based store; if cookies are misconfigured (wrong domain, Secure flag in dev), the state parameter is lost.
**How to avoid:** In development, set `sessionManager.Cookie.Secure = false`. Ensure `gothic.Store` session key is consistent. Test OAuth flow in a real browser (httptest doesn't support cross-request cookies as cleanly).
**Warning signs:** "invalid state" or "state not found" errors in OAuth callback; OAuth works locally but not in staging.

### Pitfall 6: Tailwind CSS Not Scanning Templ Files
**What goes wrong:** Tailwind produces minimal CSS because it scans for class names at compile time and misses classes used inside `.templ` files.
**Why it happens:** Tailwind v3 scans by file pattern; default config may not include `.templ` extension.
**How to avoid:** Configure Tailwind to scan `.templ` files. For v3, add to `tailwind.config.js`:
```js
content: ["./resources/**/*.templ", "./internal/**/*.templ"]
```
For v4 (standalone), use `@source` in input.css or a `tailwind.config.css` config file.
**Warning signs:** Components use `class="text-red-600"` but the CSS file doesn't contain the rule; styles missing in production.

### Pitfall 7: forgetest.NewTestDB Not Finding Migrations
**What goes wrong:** The atlas migrator path is relative; in tests run from different packages, the path resolves incorrectly.
**Why it happens:** `testdb` atlasmigrator uses relative paths by default.
**How to avoid:** Use `testhelpers` in a fixed package with absolute path resolution using `runtime.Caller(0)` or a build-time path constant.
```go
func NewTestDB(t *testing.T) *pgxpool.Pool {
    _, filename, _, _ := runtime.Caller(0)
    repoRoot := filepath.Join(filepath.Dir(filename), "../..")
    schemaPath := filepath.Join(repoRoot, "gen/atlas/schema.hcl")
    // ...
}
```
**Warning signs:** "schema.hcl not found" in tests; tests pass from one directory but fail from another.

### Pitfall 8: Templ Not HTML-Escaping Dynamic Content
**What goes wrong:** Passing user input to `templ.Raw()` creates XSS vulnerabilities. Templ auto-escapes `{ variable }` expressions, but `@templ.Raw(str)` bypasses escaping.
**Why it happens:** Developers use `Raw()` for embedding HTML snippets without realizing the security implication.
**How to avoid:** Never pass user input to `@templ.Raw()`. Use `{ variable }` (auto-escaped) for all user-controlled content. Only use `@templ.Raw()` for trusted HTML (e.g., Datastar attribute helpers).
**Warning signs:** `@templ.Raw(userInput)` in templates.

### Pitfall 9: SCS Session Not Committed After OAuth Callback
**What goes wrong:** `sessionManager.Put(ctx, "userID", id)` doesn't persist without `sessionManager.LoadAndSave` middleware running. If the OAuth callback handler runs outside the middleware chain, the session isn't committed.
**Why it happens:** SCS commits the session at the end of `LoadAndSave` — it's a post-handler hook.
**How to avoid:** Ensure all routes that need session management are under the `router.Use(sessionManager.LoadAndSave)` middleware. The OAuth callback route must be inside this group.
**Warning signs:** Login "succeeds" (no error) but subsequent requests show user as unauthenticated; session data lost on redirect.

## Code Examples

Verified patterns from official sources:

### Templ Component with Datastar Form
```go
// Source: pkg.go.dev/github.com/a-h/templ@v0.3.977 + data-star.dev/guide
// resources/products/views/form.templ

package views

import (
    "myapp/gen/models"
    datastar "github.com/starfederation/datastar-go/datastar"
)

templ ProductForm(product *models.Product, errors map[string]string) {
    <div id="product-form">
        <form
            data-signals={ `{"name": "` + product.Name + `", "price": ""}` }
            data-on:submit__prevent={ datastar.PostSSE("/html/v1/products") }
        >
            <div>
                <label for="name">Name</label>
                <input
                    id="name"
                    type="text"
                    data-bind:name
                    class="border rounded px-2 py-1"
                />
                if err, ok := errors["name"]; ok {
                    <p class="text-red-600 text-sm">{ err }</p>
                }
            </div>
            <button
                type="submit"
                class="bg-blue-600 text-white px-4 py-2 rounded"
            >
                Save
            </button>
        </form>
    </div>
}
```

### Datastar SSE Handler with Validation Errors
```go
// Source: pkg.go.dev/github.com/starfederation/datastar-go@v1.1.0
// resources/products/handlers.go

func HandleCreate(acts actions.ProductActions) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var signals struct {
            Name string `json:"name"`
        }
        if err := datastar.ReadSignals(r, &signals); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        product, err := acts.Create(r.Context(), models.ProductCreate{Name: signals.Name})
        sse := datastar.NewSSE(w, r)

        if err != nil {
            // Re-render the form with validation errors from action layer
            errs := toFieldErrors(err)  // Convert forge errors to map[string]string
            sse.PatchElementTempl(views.ProductForm(&models.Product{Name: signals.Name}, errs))
            return
        }

        // Redirect to the new product's detail page
        sse.Redirectf("/products/%s", product.ID.String())
    }
}
```

### SCS Session Setup with pgxstore
```go
// Source: pkg.go.dev/github.com/alexedwards/scs/pgxstore, pkg.go.dev/github.com/alexedwards/scs/v2
package auth

func NewSessionManager(pool *pgxpool.Pool) *scs.SessionManager {
    sm := scs.New()
    sm.Store = pgxstore.New(pool)
    sm.Lifetime = 24 * time.Hour
    sm.Cookie.Name = "forge_session"
    sm.Cookie.HttpOnly = true
    sm.Cookie.SameSite = http.SameSiteLaxMode
    return sm
}

// Wire to Chi router:
router.Use(sm.LoadAndSave)

// Read user from session:
userID := sm.GetString(r.Context(), "userID")

// Write user to session (after login):
sm.Put(r.Context(), "userID", user.ID.String())
```

### Email/Password Authentication with bcrypt
```go
// Source: pkg.go.dev/golang.org/x/crypto/bcrypt
package auth

import "golang.org/x/crypto/bcrypt"

const DefaultCost = 12  // Higher than default 10, appropriate for 2026 hardware

func HashPassword(plaintext string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), DefaultCost)
    return string(hash), err
}

func CheckPassword(plaintext, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
    // Returns nil if match, bcrypt.ErrMismatchedHashAndPassword if not
}
// Note: bcrypt truncates passwords >72 bytes. Validate max 72 chars in form.
```

### Isolated Test DB with testdb
```go
// Source: pkg.go.dev/github.com/peterldowns/testdb
package forgetest

import (
    "testing"
    "runtime"
    "path/filepath"
    "github.com/peterldowns/testdb"
    "github.com/peterldowns/testdb/migrators/atlasmigrator"
)

func NewTestDB(t *testing.T) *sql.DB {
    t.Helper()
    _, filename, _, _ := runtime.Caller(0)
    repoRoot := filepath.Join(filepath.Dir(filename), "../../")

    migrator := atlasmigrator.New(
        filepath.Join(repoRoot, "gen/atlas/schema.hcl"),
        atlasmigrator.Config{},
    )
    return testdb.New(t, testdb.Config{
        Host:     "localhost",
        Port:     "5432",
        User:     "postgres",
        Password: "postgres",
        Database: "postgres",
        Options:  "sslmode=disable",
    }, migrator)
}
```

### Diff Output for --diff Flag
```go
// Source: pkg.go.dev/github.com/sergi/go-diff/diffmatchpatch
package scaffold

import "github.com/sergi/go-diff/diffmatchpatch"

func DiffScaffold(current, fresh string) string {
    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(current, fresh, true)
    dmp.DiffCleanupSemantic(diffs)  // Improve human readability
    return dmp.DiffPrettyText(diffs)
}
```

### forge generate resource <name> CLI Command
```go
// internal/cli/generate_resource.go
// New subcommand: forge generate resource <name> [--diff]

func newGenerateResourceCmd() *cobra.Command {
    var diffOnly bool

    cmd := &cobra.Command{
        Use:   "resource <name>",
        Short: "Scaffold HTML form, list, and detail views for a resource",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := args[0]
            // Parse schema for this resource
            // If --diff: render to buffer, diff against on-disk, print, exit
            // Otherwise: check if exists, if so abort with helpful message,
            //            if not: write scaffold files + run templ generate
            return runGenerateResource(name, diffOnly)
        },
    }
    cmd.Flags().BoolVar(&diffOnly, "diff", false, "Show diff between current and freshly scaffolded views")
    return cmd
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Server-side rendering with html/template | Type-safe compiled components with Templ | 2022-2024 | Compile-time errors, IDE completion, no runtime panics from template parsing |
| HTMX for HTML-over-wire | Datastar for SSE-driven reactivity | 2023-2024 | Unified signals model eliminates need for Alpine.js; SSE replaces hx-swap polling |
| gorilla/sessions with filesystem store | alexedwards/scs with PostgreSQL store | 2021-2023 | OWASP compliant, no filesystem state, multi-process safe |
| golang.org/x/oauth2 direct integration | markbates/goth multi-provider | 2016+ stable | One interface for 60+ providers vs. per-provider custom code |
| npm-based Tailwind build chain | Tailwind standalone CLI binary | 2021 (v2.0) | Zero Node.js dependency, single binary download |
| Tailwind v3 (JIT, config.js) | Tailwind v4 (CSS-first, @import, no JS config) | 2025 (v4.0) | v4 uses CSS config (`@import "tailwindcss"`), no `tailwind.config.js` needed |

**Deprecated/outdated:**
- **Tailwind v3 config.js**: v4 replaces JavaScript config with CSS-based config; both binaries still available but v4 is current
- **starfederation/datastar/sdk/go**: Moved to dedicated `starfederation/datastar-go` repo; use the new import path
- **htmx + Alpine.js**: Datastar replaces both with a unified model; still valid but no new investment needed here
- **gorilla/sessions only**: gorilla/sessions lacks server-side storage primitives that scs provides; pgstore adapters exist but scs is cleaner

## Open Questions

1. **Tailwind version: v3.4.17 (current registry) vs v4.x**
   - What we know: Toolsync registry currently pins Tailwind at v3.4.17. v4 has different config syntax (CSS-first, no `tailwind.config.js`). Both use the same binary URL pattern (`tailwindcss-<os>-<arch>`).
   - What's unclear: Phase 6 should decide: upgrade to v4 or stay on v3? v4 is current as of 2025, but v3 is still supported. If upgrading, the scaffolded `input.css` changes from `@tailwind` directives to `@import "tailwindcss"`.
   - Recommendation: Upgrade to v4. The scaffolded `input.css` is part of Phase 6 anyway. Update toolsync registry version from 3.4.17 to latest v4.x (currently v4.1.18).

2. **OAuth2 state storage: cookie-only vs. PostgreSQL for gothic.Store**
   - What we know: `gothic.Store` is gorilla/sessions compatible. goth uses it only for the OAuth state parameter (30-60 second window during the auth flow). Completed sessions go to SCS + pgxstore.
   - What's unclear: AUTH-03 says "sessions stored in PostgreSQL (no Redis)" — does this apply to OAuth state or just completed auth sessions?
   - Recommendation: Pragmatically, OAuth state (a 60-second CSRF token) can remain in a signed cookie (`sessions.NewCookieStore`). Completed auth sessions (hours/days) MUST go to SCS + pgxstore. Document this split clearly.

3. **forgetest.NewApp — how to build the router for testing**
   - What we know: `forgetest.NewApp` needs the same router setup as production (including auth middleware, session manager, action layer wired up). The generated project's router setup lives in user-space code (`cmd/<project>/main.go` or similar).
   - What's unclear: The `forgetest` package (part of forge framework, not user project) needs to call the user project's router builder. Should `NewApp` accept an `http.Handler` argument, or should it build its own router from a config?
   - Recommendation: `forgetest.NewApp(t, handler http.Handler) *httptest.Server` — accept the handler so user projects can pass their configured router. Tests wire everything up then pass to NewApp.

4. **Visibility/Mutability field rendering (HTML-07)**
   - What we know: HTML-07 requires scaffolded form components to conditionally render fields based on user's role (Visibility/Mutability annotations). The schema IR already has `ModifierIR` for field-level modifiers.
   - What's unclear: Whether role/permission checking in the template requires the current user to be available via context, and if that's a Phase 6 concern or deferred to Phase 7 (permissions).
   - Recommendation: Scaffold the `ProductForm(product, errors, role string)` signature now with a `role` parameter. Default to rendering all fields (no conditional logic in Phase 6). Phase 7 adds the actual visibility checks. This avoids a breaking signature change later.

5. **Form primitives library location**
   - What we know: HTML-04 requires FormField, TextInput, DecimalInput, SelectInput, RelationSelect components as a "form primitives library". These are reusable across all resources.
   - What's unclear: Should this live in `gen/html/primitives/` (generated-always) or `internal/html/primitives/` (framework-provided)?
   - Recommendation: `internal/html/primitives/` — these are framework primitives, not resource-specific generation. They don't change based on schema. Generated views import them from `github.com/forge-framework/forge/internal/html/primitives`.

## Sources

### Primary (HIGH confidence)
- [pkg.go.dev/github.com/a-h/templ@v0.3.977](https://pkg.go.dev/github.com/a-h/templ) - Component interface, Handler, streaming, fragments API
- [pkg.go.dev/github.com/starfederation/datastar-go/datastar](https://pkg.go.dev/github.com/starfederation/datastar-go/datastar) - v1.1.0 PatchElementTempl, ReadSignals, Redirect, NewSSE signatures
- [pkg.go.dev/github.com/alexedwards/scs/v2](https://pkg.go.dev/github.com/alexedwards/scs/v2) - v2.9.0 session manager API, LoadAndSave middleware
- [pkg.go.dev/github.com/alexedwards/scs/pgxstore](https://pkg.go.dev/github.com/alexedwards/scs/pgxstore) - PostgreSQL store with pgx driver, table schema
- [pkg.go.dev/github.com/markbates/goth](https://pkg.go.dev/github.com/markbates/goth) - v1.82.0 provider list, gothic.Store, UseProviders
- [pkg.go.dev/github.com/peterldowns/testdb](https://pkg.go.dev/github.com/peterldowns/testdb) - NewDB API, atlasmigrator, template clone pattern
- [templ.guide/server-side-rendering/datastar/](https://templ.guide/server-side-rendering/datastar/) - PatchElementTempl integration pattern, signal binding
- [tailwindcss.com/docs/installation/tailwind-cli](https://tailwindcss.com/docs/installation/tailwind-cli) - v4.1 current, standalone CLI, -i/-o flags
- Project source: `internal/toolsync/registry.go` — confirmed Tailwind binary URL pattern, current pinned version v3.4.17
- Project source: `internal/cli/routes.go` — confirmed `apiRoutes()` / Phase 6 `htmlRoutes()` separation already in place
- Project source: `internal/generator/templates/factory.go.tmpl` — confirmed builder pattern already generated (`BuildProduct().WithName("...").Build()`)

### Secondary (MEDIUM confidence)
- [data-star.dev/guide/backend_requests_sse_events](https://data-star.dev/guide/backend_requests_sse_events) - Signal transmission rules, GET vs POST body
- [pkg.go.dev/github.com/sergi/go-diff/diffmatchpatch](https://pkg.go.dev/github.com/sergi/go-diff/diffmatchpatch) - DiffMain, DiffPrettyText API
- [github.com/tailwindlabs/tailwindcss/releases](https://github.com/tailwindlabs/tailwindcss/releases) - v4.1.18 latest release, binary naming convention confirmed as `tailwindcss-<platform>-<arch>`

### Tertiary (LOW confidence, marked for validation)
- WebSearch results on Tailwind v4 config syntax changes — verify by running `tailwindcss --help` with v4 binary
- Gist: datastar system prompt for RC6 — patterns may have changed in v1.1.0 release

## Metadata

**Confidence breakdown:**
- Standard stack (Templ, Datastar, SCS, Goth, testdb): HIGH - All verified from official pkg.go.dev package pages with current versions
- Architecture patterns (scaffold-once, diff, route grouping): HIGH - Derived from reading existing Phase 5 code + requirements; the `apiRoutes()`/`htmlRoutes()` hook is already in production code
- Auth implementation (bcrypt, pgxstore, OAuth flow): HIGH - Standard Go patterns, verified from official docs
- Pitfalls (templ generate subprocess, scaffold collision, signal naming): MEDIUM-HIGH - Derived from library docs + known integration issues
- Tailwind v3 vs v4 decision: MEDIUM - Both work; recommendation to upgrade needs verification against actual v4 binary behavior

**Research date:** 2026-02-17
**Valid until:** 2026-03-19 (30 days — templ and Datastar are active but stabilizing; check for breaking changes in Datastar v1.x)
