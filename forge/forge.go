package forge

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alternayte/forge/forge/auth"
	internalapi "github.com/alternayte/forge/internal/api"
	"github.com/alternayte/forge/internal/config"
)

// App is the forge application lifecycle manager.
// Create with New(), configure with builder methods, start with Listen().
type App struct {
	cfg          *config.Config
	router       chi.Router
	pool         *pgxpool.Pool
	apiRoutesFn  func(huma.API)
	htmlRoutesFn func(chi.Router)
	recoveryMw   func(http.Handler) http.Handler
	tokenStore   auth.TokenStore
	apiKeyStore  auth.APIKeyStore
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

// Listen starts the HTTP server and blocks until SIGTERM/SIGINT.
// On shutdown: stops accepting connections, drains in-flight requests,
// closes the DB pool.
func (a *App) Listen(addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to database
	pool, err := pgxpool.New(ctx, a.cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("forge: connect to database: %w", err)
	}
	defer pool.Close()
	a.pool = pool

	// Set up session manager.
	// NOTE: auth.NewSessionManager already sets sm.Store = pgxstore.New(pool) internally.
	isDev := a.cfg.Server.Host == "localhost"
	sm := auth.NewSessionManager(pool, isDev)

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
	if a.htmlRoutesFn != nil {
		err := internalapi.SetupHTML(a.router, internalapi.HTMLServerConfig{
			SessionManager: sm,
			RegisterRoutes: a.htmlRoutesFn,
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

// Pool returns the database connection pool.
// Only valid after Listen() has been called (for test infrastructure, use forge/forgetest).
func (a *App) Pool() *pgxpool.Pool {
	return a.pool
}
