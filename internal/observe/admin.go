package observe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alternayte/forge/internal/cli"
	"github.com/alternayte/forge/internal/config"
)

// StartAdminServer creates and starts a dedicated admin/ops HTTP server on the
// port specified in cfg (default 9090). It exposes three operational endpoints:
//
//   - /metrics    — Prometheus metrics (via OTel-Prometheus bridge, OTEL-02)
//   - /healthz    — JSON health check with build version info (OTEL-03)
//   - /debug/pprof/ — Go pprof profiling endpoints (admin-only, not on public mux)
//
// The server runs in a goroutine. The caller is responsible for calling
// Shutdown on the returned *http.Server during application teardown.
func StartAdminServer(ctx context.Context, cfg config.AdminConfig) *http.Server {
	mux := http.NewServeMux()

	// /metrics — Prometheus exposition format via OTel bridge
	mux.Handle("/metrics", promhttp.Handler())

	// /healthz — JSON health/build-info response
	mux.HandleFunc("/healthz", healthzHandler)

	// /debug/pprof/ — pprof endpoints registered on the admin mux only.
	// Do NOT use http.DefaultServeMux to avoid exposing pprof on the public
	// server. Each handler is registered explicitly.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	port := cfg.Port
	if port == 0 {
		port = 9090
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Admin server errors are not fatal — log but don't crash.
			_ = err
		}
	}()

	return srv
}

// healthzResponse is the JSON body returned by /healthz.
type healthzResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Built   string `json:"built"`
}

// healthzHandler responds with a JSON payload confirming the service is alive
// and including build provenance from ldflags-injected variables (OTEL-03).
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := healthzResponse{
		Status:  "ok",
		Version: cli.Version,
		Commit:  cli.Commit,
		Built:   cli.Date,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
