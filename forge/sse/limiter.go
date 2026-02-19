package sse

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrTooManyConnections is returned when the global SSE connection cap is exceeded.
var ErrTooManyConnections = errors.New("too many SSE connections")

// ErrTooManyConnectionsForUser is returned when the per-user SSE connection cap is exceeded.
var ErrTooManyConnectionsForUser = errors.New("too many SSE connections for user")

// SSELimiter enforces global and per-user SSE connection limits using atomic counters.
// Acquire returns a release function on success, or an error when the cap is exceeded.
//
// SSE-01: Global connection cap via active atomic counter.
// SSE-02: Per-user cap via sync.Map keyed by userID.
// SSE-03: Graceful shutdown is handled by the SSE HTTP handler detecting ctx.Done()
// and writing a close event frame before returning — the limiter itself only tracks
// counts and provides the release callback.
type SSELimiter struct {
	maxTotal   int64
	maxPerUser int64
	active     atomic.Int64
	perUser    sync.Map // string(userID) -> *atomic.Int64
}

// NewSSELimiter creates a new SSELimiter with the given global and per-user limits.
// Callers pass SSEConfig.MaxTotalConnections and SSEConfig.MaxPerUser directly;
// the limiter does not import config to avoid circular dependencies.
func NewSSELimiter(maxTotal, maxPerUser int) *SSELimiter {
	return &SSELimiter{
		maxTotal:   int64(maxTotal),
		maxPerUser: int64(maxPerUser),
	}
}

// Acquire attempts to reserve an SSE connection slot for the given user.
// Returns a release function that must be called (typically via defer) when the
// connection ends, and nil error on success.
// Returns ErrTooManyConnections when the global cap is reached.
// Returns ErrTooManyConnectionsForUser when the per-user cap is reached.
func (l *SSELimiter) Acquire(userID string) (release func(), err error) {
	// SSE-01: Check global cap.
	if l.active.Load() >= l.maxTotal {
		return nil, ErrTooManyConnections
	}

	// SSE-02: Get or create per-user atomic counter.
	val, _ := l.perUser.LoadOrStore(userID, &atomic.Int64{})
	userCount := val.(*atomic.Int64)

	// Check per-user cap.
	if userCount.Load() >= l.maxPerUser {
		return nil, ErrTooManyConnectionsForUser
	}

	// Increment both counters — slot is now reserved.
	l.active.Add(1)
	userCount.Add(1)

	release = func() {
		l.active.Add(-1)
		userCount.Add(-1)
	}
	return release, nil
}

// ActiveConnections returns the current total number of active SSE connections.
// Useful for Prometheus metrics.
func (l *SSELimiter) ActiveConnections() int64 {
	return l.active.Load()
}

// ActiveForUser returns the number of active SSE connections for the given user.
// Useful for debugging and per-user metrics.
func (l *SSELimiter) ActiveForUser(userID string) int64 {
	val, ok := l.perUser.Load(userID)
	if !ok {
		return 0
	}
	return val.(*atomic.Int64).Load()
}
