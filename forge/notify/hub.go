package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgxlisten"
)

// maxNotifyPayloadBytes is the PostgreSQL pg_notify payload size limit.
// Exceeding this limit causes pg_notify to error or silently truncate.
const maxNotifyPayloadBytes = 8000

// Executor is the minimal interface required by PostgresHub.Publish.
// It is satisfied by *pgx.Conn, *pgxpool.Pool, and pgx.Tx — any type that
// can execute SQL. The notify package does not import the generated actions
// package to avoid circular dependencies.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// NotifyHub is the interface for publishing and subscribing to real-time events
// via PostgreSQL LISTEN/NOTIFY. The interface allows swapping the backend to
// Redis, NATS, or any other pub/sub system (SSE-06).
type NotifyHub interface {
	// Subscribe registers a subscriber for events on the given channel scoped
	// to the given tenant. Returns a Subscription whose Events channel receives
	// fan-out notifications. Call Subscription.Close to unsubscribe.
	Subscribe(channel string, tenantID uuid.UUID) *Subscription

	// Publish sends a pg_notify payload to all subscribers on the given channel
	// scoped to the given tenant. The payload must be < 8000 bytes (PostgreSQL limit).
	Publish(ctx context.Context, channel string, tenantID uuid.UUID, payload []byte) error

	// Start begins listening for PostgreSQL notifications. It blocks until ctx is
	// cancelled (SSE-03: graceful shutdown via context propagation). Any error
	// returned indicates a fatal configuration problem; reconnection errors are
	// handled internally by pgxlisten.
	Start(ctx context.Context) error
}

// notifyMessage is the JSON envelope written to pg_notify payloads.
// All PostgreSQL NOTIFY calls use the single "forge_events" channel; routing
// is performed by the channel + tenant_id fields inside the payload.
type notifyMessage struct {
	Channel    string          `json:"channel"`
	TenantID   string          `json:"tenant_id"`
	RawPayload json.RawMessage `json:"payload"`
}

// PostgresHub implements NotifyHub using a single dedicated PostgreSQL
// LISTEN connection (via pgxlisten) that fans out events to per-subscriber
// buffered Go channels.
//
// One connection handles all channels — routing is done in the payload JSON.
// This avoids per-client LISTEN connections which do not scale (SSE-04).
type PostgresHub struct {
	connConfig *pgx.ConnConfig
	db         Executor
	bufferSize int

	mu   sync.RWMutex
	subs map[string][]*internalSub // key = "channel:tenantID"
}

// NewPostgresHub creates a PostgresHub.
//
//   - connConfig: pgx connection config for the dedicated LISTEN connection.
//     Must be a separate connection from the application pool — LISTEN state
//     is connection-scoped and cannot be shared with query traffic (SSE-04).
//   - db: Executor for pg_notify calls in Publish (e.g., the application pool).
//   - bufferSize: per-subscriber channel capacity. Defaults to 32 when 0.
func NewPostgresHub(connConfig *pgx.ConnConfig, db Executor, bufferSize int) *PostgresHub {
	if bufferSize <= 0 {
		bufferSize = 32
	}
	return &PostgresHub{
		connConfig: connConfig,
		db:         db,
		bufferSize: bufferSize,
		subs:       make(map[string][]*internalSub),
	}
}

// Subscribe registers a new subscriber for events on the given channel scoped
// to the given tenant. Returns a Subscription with a buffered Events channel.
// The subscription is active immediately; any in-flight pg_notify dispatches
// that arrive after Subscribe returns will be delivered to the channel.
func (h *PostgresHub) Subscribe(channel string, tenantID uuid.UUID) *Subscription {
	key := channel + ":" + tenantID.String()

	ch := make(chan Event, h.bufferSize)
	sub := &internalSub{ch: ch}

	h.mu.Lock()
	h.subs[key] = append(h.subs[key], sub)
	h.mu.Unlock()

	cancel := func() { h.unsubscribe(key, sub) }

	return &Subscription{
		Events: ch,
		cancel: cancel,
	}
}

// unsubscribe removes sub from the subs map and closes its channel.
// Called by the Subscription.cancel closure.
func (h *PostgresHub) unsubscribe(key string, target *internalSub) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs := h.subs[key]
	for i, s := range subs {
		if s == target {
			// Remove by swapping with last element.
			subs[i] = subs[len(subs)-1]
			subs[len(subs)-1] = nil
			h.subs[key] = subs[:len(subs)-1]
			break
		}
	}
	close(target.ch)
}

// Publish sends a pg_notify message on the "forge_events" PostgreSQL channel.
// The payload is wrapped in a JSON envelope with channel and tenantID for
// in-process routing. Returns an error if the marshaled payload exceeds the
// 8000-byte PostgreSQL NOTIFY limit.
func (h *PostgresHub) Publish(ctx context.Context, channel string, tenantID uuid.UUID, payload []byte) error {
	msg := notifyMessage{
		Channel:    channel,
		TenantID:   tenantID.String(),
		RawPayload: json.RawMessage(payload),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	if len(data) > maxNotifyPayloadBytes {
		return fmt.Errorf("notify: payload too large (%d bytes, max %d): keep payloads to IDs only", len(data), maxNotifyPayloadBytes)
	}

	_, err = h.db.Exec(ctx, "SELECT pg_notify($1, $2)", "forge_events", string(data))
	if err != nil {
		return fmt.Errorf("notify: pg_notify: %w", err)
	}
	return nil
}

// Start begins listening on the "forge_events" PostgreSQL NOTIFY channel.
// It blocks until ctx is cancelled, enabling graceful shutdown (SSE-03).
// Non-fatal errors (e.g., transient connection failures) are logged; pgxlisten
// handles reconnection automatically with the configured ReconnectDelay.
func (h *PostgresHub) Start(ctx context.Context) error {
	listener := &pgxlisten.Listener{
		Connect: func(ctx context.Context) (*pgx.Conn, error) {
			return pgx.ConnectConfig(ctx, h.connConfig)
		},
		ReconnectDelay: 5 * time.Second,
		LogError: func(ctx context.Context, err error) {
			// Only log if the context is not already cancelled — avoid noise
			// from expected shutdown errors.
			if !errors.Is(err, context.Canceled) {
				slog.WarnContext(ctx, "notify hub: connection error", "err", err)
			}
		},
	}

	listener.Handle("forge_events", pgxlisten.HandlerFunc(h.handleNotification))

	return listener.Listen(ctx)
}

// handleNotification is called by pgxlisten for each "forge_events" notification.
// It unmarshals the envelope, locates subscribers for the (channel, tenantID) key,
// and fans out using non-blocking sends (SSE-05 backpressure).
func (h *PostgresHub) handleNotification(ctx context.Context, n *pgconn.Notification, conn *pgx.Conn) error {
	var msg notifyMessage
	if err := json.Unmarshal([]byte(n.Payload), &msg); err != nil {
		slog.WarnContext(ctx, "notify hub: invalid payload", "err", err, "channel", n.Channel)
		return nil // non-fatal — malformed payloads should not kill the listener
	}

	key := msg.Channel + ":" + msg.TenantID

	h.mu.RLock()
	subs := h.subs[key]
	h.mu.RUnlock()

	event := Event{
		Channel: msg.Channel,
		Payload: msg.RawPayload,
	}

	for _, sub := range subs {
		select {
		case sub.ch <- event:
			// Delivered successfully.
		default:
			// SSE-05: buffer full — drop the event and send a refresh signal so
			// the client knows to reload current state rather than miss an update.
			select {
			case sub.ch <- Event{Channel: "refresh"}:
			default:
				// Even the refresh signal cannot be sent; subscriber is severely
				// behind. Drop silently — the subscriber will eventually drain
				// and receive a future refresh.
			}
		}
	}

	return nil
}
