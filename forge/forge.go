package forge

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alternayte/forge/forge/auth"
	internalapi "github.com/alternayte/forge/internal/api"
	"github.com/alternayte/forge/internal/config"
)

// ConnectDB creates a pgxpool.Pool from the given Config.
// Call this early in main() so the pool is available for registry wiring.
func ConnectDB(cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("forge: connect to database: %w", err)
	}
	return pool, nil
}

// App is the forge application lifecycle manager.
// Create with New(), configure with builder methods, start with Listen().
type App struct {
	cfg              *config.Config
	router           chi.Router
	pool             *pgxpool.Pool
	apiRoutesFn      func(huma.API)
	htmlRoutesFn     func(chi.Router)
	recoveryMw       func(http.Handler) http.Handler
	tokenStore       auth.TokenStore
	apiKeyStore      auth.APIKeyStore
	oauthConfig      *auth.OAuthConfig
	findOrCreateUser auth.UserFinder
	authenticateUser auth.PasswordAuthenticator
	requireAuth      bool
	publicRoutesFn   func(chi.Router)
}

// New creates a new App from configuration loaded via LoadConfig.
func New(cfg Config) *App {
	return &App{
		cfg:    cfg.cfg,
		router: chi.NewRouter(),
	}
}

// RegisterAPIRoutes sets the function that registers all API routes on the Huma API.
// The function receives a huma.API and should call genapi.RegisterAllRoutes(api, registry).
func (a *App) RegisterAPIRoutes(fn func(huma.API)) *App {
	a.apiRoutesFn = fn
	return a
}

// RegisterHTMLRoutes sets the function that registers all HTML routes on a chi.Router.
// The function receives a session-protected chi.Router group.
func (a *App) RegisterHTMLRoutes(fn func(chi.Router)) *App {
	a.htmlRoutesFn = fn
	return a
}

// UseRecovery sets a custom panic recovery middleware. If not called,
// a default passthrough middleware is used.
func (a *App) UseRecovery(mw func(http.Handler) http.Handler) *App {
	a.recoveryMw = mw
	return a
}

// UseTokenStore sets the bearer token store for API authentication.
func (a *App) UseTokenStore(ts auth.TokenStore) *App {
	a.tokenStore = ts
	return a
}

// UseAPIKeyStore sets the API key store for API authentication.
func (a *App) UseAPIKeyStore(ks auth.APIKeyStore) *App {
	a.apiKeyStore = ks
	return a
}

// UsePool sets an externally-created database connection pool.
// When set, Listen() will skip internal pool creation and use this pool instead.
// The caller is responsible for closing the pool after Listen() returns.
func (a *App) UsePool(pool *pgxpool.Pool) *App {
	a.pool = pool
	return a
}

// UseOAuth configures OAuth2 providers (Google/GitHub) for HTML session auth.
func (a *App) UseOAuth(cfg auth.OAuthConfig, findOrCreateUser auth.UserFinder) *App {
	a.oauthConfig = &cfg
	a.findOrCreateUser = findOrCreateUser
	return a
}

// UsePasswordAuth configures email/password authentication for HTML session auth.
func (a *App) UsePasswordAuth(authenticateUser auth.PasswordAuthenticator) *App {
	a.authenticateUser = authenticateUser
	return a
}

// RequireAuth enables session enforcement on HTML routes.
// Unauthenticated users are redirected to /auth/login.
func (a *App) RequireAuth() *App {
	a.requireAuth = true
	return a
}

// RegisterPublicRoutes sets a function that registers routes outside the
// RequireSession middleware group. Use this for custom unauthenticated pages.
func (a *App) RegisterPublicRoutes(fn func(chi.Router)) *App {
	a.publicRoutesFn = fn
	return a
}

// Listen starts the HTTP server and blocks until SIGTERM/SIGINT.
// On shutdown: stops accepting connections, drains in-flight requests,
// closes the DB pool (if created internally).
func (a *App) Listen(addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to database if no external pool was provided via UsePool().
	if a.pool == nil {
		pool, err := pgxpool.New(ctx, a.cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("forge: connect to database: %w", err)
		}
		defer pool.Close()
		a.pool = pool
	}

	// Set up session manager.
	// NOTE: auth.NewSessionManager already sets sm.Store = pgxstore.New(pool) internally.
	isDev := a.cfg.Server.Host == "localhost"
	sm := auth.NewSessionManager(a.pool, isDev)

	// Set up OAuth providers if configured.
	if a.oauthConfig != nil {
		auth.SetupOAuth(*a.oauthConfig)
	}

	// Serve static files from public/ directory.
	// Files in public/ are served at the root path (e.g., public/css/output.css -> /css/output.css).
	// If the file doesn't exist, the request passes through to application routes.
	a.router.Use(staticFiles("public"))

	// Recovery middleware fallback
	recoveryMw := a.recoveryMw
	if recoveryMw == nil {
		recoveryMw = func(next http.Handler) http.Handler { return next }
	}

	// Wire API routes
	if a.apiRoutesFn != nil {
		_, err := internalapi.SetupAPI(
			a.router,
			a.cfg.API,
			a.tokenStore,
			a.apiKeyStore,
			recoveryMw,
			a.apiRoutesFn,
		)
		if err != nil {
			return fmt.Errorf("forge: setup API: %w", err)
		}
	}

	// Wire HTML routes
	if a.htmlRoutesFn != nil || a.requireAuth || a.findOrCreateUser != nil || a.authenticateUser != nil {
		err := internalapi.SetupHTML(a.router, internalapi.HTMLServerConfig{
			SessionManager:       sm,
			RegisterRoutes:       a.htmlRoutesFn,
			RequireAuth:          a.requireAuth,
			RegisterPublicRoutes: a.buildPublicRoutesFn(sm),
		})
		if err != nil {
			return fmt.Errorf("forge: setup HTML: %w", err)
		}
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: a.router,
	}

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		slog.Info("forge: shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("forge: shutdown error", "err", err)
		}
	}()

	slog.Info("forge: listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("forge: server error: %w", err)
	}

	return nil
}

// buildPublicRoutesFn composes auth routes and custom public routes into a
// single function for the public (unauthenticated) route group.
func (a *App) buildPublicRoutesFn(sm *scs.SessionManager) func(chi.Router) {
	if a.findOrCreateUser == nil && a.authenticateUser == nil && a.publicRoutesFn == nil {
		return nil
	}
	return func(r chi.Router) {
		if a.findOrCreateUser != nil || a.authenticateUser != nil {
			auth.RegisterOAuthRoutes(r, sm, a.findOrCreateUser, a.authenticateUser)
		}
		if a.publicRoutesFn != nil {
			a.publicRoutesFn(r)
		}
	}
}

// Pool returns the database connection pool.
// Only valid after Listen() has been called (for test infrastructure, use forge/forgetest).
func (a *App) Pool() *pgxpool.Pool {
	return a.pool
}

// staticFiles returns middleware that serves files from the given directory.
// If the requested path matches a file on disk, it's served directly.
// Otherwise the request passes through to the next handler.
func staticFiles(dir string) func(http.Handler) http.Handler {
	absDir, _ := filepath.Abs(dir)
	fs := http.FileServer(http.Dir(absDir))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only serve GET/HEAD for static files
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Check if the file exists in the public directory
			filePath := filepath.Join(absDir, filepath.Clean(r.URL.Path))
			info, err := os.Stat(filePath)
			if err != nil || info.IsDir() {
				next.ServeHTTP(w, r)
				return
			}

			fs.ServeHTTP(w, r)
		})
	}
}
