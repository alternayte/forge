package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateHTML generates the HTML primitives library and Datastar SSE helpers.
// These are generated-always files in gen/html/ (not scaffolded-once).
func GenerateHTML(resources []parser.ResourceIR, outputDir, projectModule string) error {
	// Template data with project module for imports
	data := struct {
		ProjectModule string
	}{
		ProjectModule: projectModule,
	}

	// Generate gen/html/primitives/primitives.templ
	primitivesDir := filepath.Join(outputDir, "html", "primitives")
	if err := ensureDir(primitivesDir); err != nil {
		return err
	}

	primitivesRaw, err := renderTemplate("templates/html_primitives.templ.tmpl", data)
	if err != nil {
		return err
	}

	primitivesPath := filepath.Join(primitivesDir, "primitives.templ")
	if err := writeRawFile(primitivesPath, primitivesRaw); err != nil {
		return err
	}

	// Generate gen/html/sse/sse.go
	sseDir := filepath.Join(outputDir, "html", "sse")
	if err := ensureDir(sseDir); err != nil {
		return err
	}

	sseRaw, err := renderTemplate("templates/html_sse.go.tmpl", data)
	if err != nil {
		return err
	}

	ssePath := filepath.Join(sseDir, "sse.go")
	if err := writeGoFile(ssePath, sseRaw); err != nil {
		return err
	}

	return nil
}
