package forgetest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// SSEEvent represents a single server-sent event with a type and data payload.
type SSEEvent struct {
	// Type is the event name from the "event:" line (e.g. "datastar-merge-fragments").
	Type string
	// Data is the payload from the "data:" line.
	Data string
}

// PostDatastar sends a Datastar SSE form submission to the test server.
// It marshals signals to JSON, sets Content-Type: application/json and
// Accept: text/event-stream, then performs the POST.
//
// The response body is registered for cleanup. Callers can call ReadSSEEvents
// on the response to inspect the returned Datastar events.
//
//	resp := forgetest.PostDatastar(t, srv, "/users", map[string]any{"name": "Alice"})
//	events := forgetest.ReadSSEEvents(t, resp)
func PostDatastar(t *testing.T, srv *httptest.Server, path string, signals any) *http.Response {
	t.Helper()

	body, err := json.Marshal(signals)
	if err != nil {
		t.Fatalf("forgetest.PostDatastar: failed to marshal signals: %v", err)
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("forgetest.PostDatastar: failed to create request: %v", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("forgetest.PostDatastar: request failed: %v", err)
		return nil
	}

	t.Cleanup(func() {
		resp.Body.Close()
	})

	return resp
}

// GetDatastar sends a GET request with Accept: text/event-stream to the test server.
// This simulates a Datastar client requesting a streaming SSE endpoint.
//
//	resp := forgetest.GetDatastar(t, srv, "/users/1")
//	events := forgetest.ReadSSEEvents(t, resp)
func GetDatastar(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("forgetest.GetDatastar: failed to create request: %v", err)
		return nil
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("forgetest.GetDatastar: request failed: %v", err)
		return nil
	}

	t.Cleanup(func() {
		resp.Body.Close()
	})

	return resp
}

// ReadSSEEvents reads a response body and parses it as server-sent events.
// It returns a slice of SSEEvent, one per "event:" block in the response.
//
// SSE format parsed:
//
//	event: datastar-merge-fragments
//	data: <div id="main">...</div>
//
// Blank lines separate events. Lines without "event:" or "data:" prefixes are ignored.
func ReadSSEEvents(t *testing.T, resp *http.Response) []SSEEvent {
	t.Helper()

	var events []SSEEvent
	var current SSEEvent
	hasEvent := false

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event:"):
			current.Type = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			hasEvent = true
		case strings.HasPrefix(line, "data:"):
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if current.Data == "" {
				current.Data = data
			} else {
				// Multi-line data: join with newline
				current.Data = current.Data + "\n" + data
			}
			hasEvent = true
		case line == "":
			// Blank line signals end of an event block
			if hasEvent {
				events = append(events, current)
				current = SSEEvent{}
				hasEvent = false
			}
		}
	}

	// Flush any final event that wasn't followed by a blank line
	if hasEvent {
		events = append(events, current)
	}

	if err := scanner.Err(); err != nil {
		t.Logf("forgetest.ReadSSEEvents: scanner error: %v", err)
	}

	return events
}
