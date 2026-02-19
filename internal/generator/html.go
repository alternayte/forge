package generator

import (
	"path/filepath"

	"github.com/alternayte/forge/internal/parser"
)

// GenerateHTML generates the HTML primitives library, Datastar SSE helpers,
// and the HTML route registration dispatcher.
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

	// Generate gen/html/register_all.go (HTML route dispatcher)
	registerData := struct {
		Resources     []parser.ResourceIR
		ProjectModule string
	}{
		Resources:     resources,
		ProjectModule: projectModule,
	}

	registerRaw, err := renderTemplate("templates/html_register_all.go.tmpl", registerData)
	if err != nil {
		return err
	}

	registerPath := filepath.Join(outputDir, "html", "register_all.go")
	if err := writeGoFile(registerPath, registerRaw); err != nil {
		return err
	}

	return nil
}
