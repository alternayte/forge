package auth

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// SessionKeyUserID is the key used to store the authenticated user's ID in
	// the session. The value is a string UUID.
	SessionKeyUserID = "user_id"

	// SessionKeyUserEmail is the key used to store the authenticated user's
	// email address in the session.
	SessionKeyUserEmail = "user_email"
)

// NewSessionManager creates and configures an SCS session manager backed by
// PostgreSQL via pgxstore. Sessions are scoped to 24 hours. The cookie is
// named "forge_session" and is marked httpOnly with SameSite=Lax to prevent
// CSRF via cross-site requests.
//
// isDev controls Cookie.Secure: set false during local development (HTTP), true
// in production (HTTPS only).
func NewSessionManager(pool *pgxpool.Pool, isDev bool) *scs.SessionManager {
	sm := scs.New()
	sm.Store = pgxstore.New(pool)
	sm.Lifetime = 24 * time.Hour
	sm.Cookie.Name = "forge_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = !isDev
	return sm
}

// SessionMiddleware returns an http.Handler middleware that loads the session
// from the store on each incoming request and saves it back before the response
// is written. This is the Chi-compatible form of SCS's LoadAndSave middleware.
func SessionMiddleware(sm *scs.SessionManager) func(http.Handler) http.Handler {
	return sm.LoadAndSave
}
