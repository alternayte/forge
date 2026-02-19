package notify

// Event represents a notification received from PostgreSQL LISTEN/NOTIFY and
// fan-out to SSE subscribers.
//
// The Type field controls SSE event framing:
//   - Empty string (zero value): data-only event (standard SSE data frame).
//   - "close": signals graceful server shutdown; the SSE handler should write
//     "event: close\ndata: {}\n\n" and terminate the connection (SSE-03).
//   - "refresh": backpressure signal indicating the subscriber's buffer was full
//     and events were dropped; the client should reload the resource state (SSE-05).
type Event struct {
	// Channel is the logical event channel (e.g., "products", "refresh").
	Channel string

	// Payload is the raw JSON payload from the NOTIFY message, already extracted
	// from the outer notifyMessage envelope. May be nil for control events.
	Payload []byte

	// Type is the SSE event type. Empty means a standard data event. "close" and
	// "refresh" are sentinel values recognised by the SSE HTTP handler.
	Type string
}

// CloseEvent is a sentinel Event value that the SSE handler sends when the
// server is shutting down (ctx cancelled). Signals clients to reconnect later.
var CloseEvent = Event{Type: "close"}

// internalSub is the internal subscriber handle. It holds the writable side of
// the event channel; only the notify package should write to it.
type internalSub struct {
	ch chan Event
}

// Subscription is the consumer-facing handle returned by NotifyHub.Subscribe.
// Events is a read-only channel; call Close to unsubscribe and release resources.
type Subscription struct {
	// Events is the read-only channel the consumer reads events from.
	Events <-chan Event

	// cancel unsubscribes and closes the underlying channel.
	cancel func()
}

// Close unsubscribes from the hub and drains/closes the underlying channel.
// Calling Close more than once is safe.
func (s *Subscription) Close() {
	s.cancel()
}
