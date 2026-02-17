package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nyxstack/scalarui"
)

// RegisterDocsHandler registers the Scalar UI documentation handler at /api/docs.
// Scalar UI is embedded directly in the binary — no CDN dependency, consistent with
// the single-binary deployment philosophy.
//
// The specURL parameter is the URL path where the OpenAPI spec is served, typically
// "/api/openapi.json". Scalar UI will fetch this URL from the browser to render the
// interactive documentation.
//
// Huma's built-in docs handler (Stoplight Elements via CDN) is disabled in SetupAPI by
// setting DocsPath to "". This handler replaces it with a self-hosted Scalar UI that
// works offline and air-gapped environments.
func RegisterDocsHandler(router chi.Router, specURL string) {
	cfg := scalarui.NewConfig().
		WithTitle("Forge API Documentation").
		WithDescription("Production-ready REST API").
		WithURL(specURL).
		WithTheme("purple").
		WithDarkMode(true).
		WithSidebar(true).
		WithInteractive(true)

	ui := scalarui.New(cfg)

	router.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		html, err := ui.Render()
		if err != nil {
			http.Error(w, "failed to render API docs", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
}
