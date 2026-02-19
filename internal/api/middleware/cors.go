package middleware

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/alternayte/forge/internal/config"
)

// CORSMiddleware returns an HTTP middleware that applies CORS headers based on
// the forge.toml CORSConfig. If config.Enabled is false, a no-op pass-through
// middleware is returned so the middleware slot can always be wired in.
//
// SAFETY: Combining AllowCredentials with a wildcard origin ("*") is forbidden
// by the CORS spec. If detected, credentials are disabled and a warning is
// logged so the API remains functional rather than breaking silently.
func CORSMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}

	allowCredentials := cfg.AllowCredentials

	// Guard against the wildcard + credentials pitfall (CORS spec violation).
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" && allowCredentials {
			log.Println("[forge/cors] WARNING: wildcard origin (*) cannot be combined with " +
				"AllowCredentials=true — disabling credentials to keep CORS functional. " +
				"Set specific allowed origins in forge.toml to re-enable credentials.")
			allowCredentials = false
			break
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           cfg.MaxAge,
	})

	return c.Handler
}
