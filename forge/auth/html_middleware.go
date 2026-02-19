package auth

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// RequireSession returns a Chi-compatible middleware that enforces session-based
// authentication for HTML routes. Unlike the API AuthMiddleware — which returns
// a 401 JSON response — this middleware redirects unauthenticated users to the
// login page, matching the UX expectation for browser-rendered pages.
func RequireSession(sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sm.GetString(r.Context(), SessionKeyUserID)
			if userID == "" {
				http.Redirect(w, r, "/auth/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetSessionUserID returns the authenticated user's ID string from the session.
// Returns an empty string when no user is logged in.
func GetSessionUserID(sm *scs.SessionManager, r *http.Request) string {
	return sm.GetString(r.Context(), SessionKeyUserID)
}

// GetSessionUserEmail returns the authenticated user's email address from the
// session. Returns an empty string when no user is logged in.
func GetSessionUserEmail(sm *scs.SessionManager, r *http.Request) string {
	return sm.GetString(r.Context(), SessionKeyUserEmail)
}

// LoginUser writes the user's ID and email into the session. It calls
// sm.RenewToken first to rotate the session ID and prevent session fixation
// attacks. Must be called after the session middleware has loaded the session
// (i.e. inside an HTTP handler, not before SessionMiddleware in the chain).
func LoginUser(sm *scs.SessionManager, r *http.Request, userID, email string) error {
	// RenewToken rotates the session token on the wire, invalidating any
	// pre-login session token that an attacker may have already obtained.
	if err := sm.RenewToken(r.Context()); err != nil {
		return err
	}
	sm.Put(r.Context(), SessionKeyUserID, userID)
	sm.Put(r.Context(), SessionKeyUserEmail, email)
	return nil
}

// LogoutUser destroys the entire session, removing all stored data and
// invalidating the session cookie on the client side.
func LogoutUser(sm *scs.SessionManager, r *http.Request) error {
	return sm.Destroy(r.Context())
}
