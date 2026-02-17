package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"github.com/forge-framework/forge/internal/config"
)

// RateLimitMiddleware returns an HTTP middleware that enforces per-IP token
// bucket rate limiting using go-limiter. If config.Enabled is false, a no-op
// pass-through middleware is returned.
//
// Phase 5 note: All requests are rated against the Default tier. Tiered
// enforcement (authenticated vs API key vs anonymous) will be wired once the
// full server is assembled in Plan 03, which can inspect auth context values.
//
// The middleware automatically sets X-RateLimit-Limit, X-RateLimit-Remaining,
// and X-RateLimit-Reset response headers. Requests that exceed the limit
// receive HTTP 429 Too Many Requests.
func RateLimitMiddleware(cfg config.RateLimitConfig) (func(http.Handler) http.Handler, error) {
	if !cfg.Enabled {
		noop := func(next http.Handler) http.Handler { return next }
		return noop, nil
	}

	interval, err := parseDuration(cfg.Default.Interval)
	if err != nil {
		return nil, fmt.Errorf("rate_limit.default.interval: %w", err)
	}

	store, err := memorystore.New(&memorystore.Config{
		Tokens:   cfg.Default.Tokens,
		Interval: interval,
	})
	if err != nil {
		return nil, fmt.Errorf("rate limit store: %w", err)
	}

	m, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
	if err != nil {
		return nil, fmt.Errorf("rate limit middleware: %w", err)
	}

	return m.Handle, nil
}

// parseDuration wraps time.ParseDuration with a friendlier error message for
// configuration values. It accepts any duration string accepted by the
// standard library (e.g. "1m", "30s", "2h").
func parseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return d, nil
}
