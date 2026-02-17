package forgetest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewApp creates an httptest.Server wrapping the given handler and registers
// automatic cleanup to close the server when the test completes.
//
// The handler should be the fully-assembled application router. NewApp is
// intentionally thin so that forgetest stays generic across any forge app.
//
//	srv := forgetest.NewApp(t, myRouter)
//	pool := forgetest.NewTestPool(t)
func NewApp(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return srv
}

// AppURL returns the base URL of the test server combined with the given path.
// This is a convenience helper for building request URLs in tests.
//
//	url := forgetest.AppURL(srv, "/api/v1/users")
func AppURL(srv *httptest.Server, path string) string {
	return srv.URL + path
}
