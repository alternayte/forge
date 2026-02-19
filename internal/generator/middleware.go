package generator

import (
	"path/filepath"

	"github.com/alternayte/forge/internal/parser"
)

// GenerateMiddleware generates panic recovery and error rendering middleware.
// This generates two files (not per-resource) since middleware is global.
func GenerateMiddleware(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create middleware directory
	middlewareDir := filepath.Join(outputDir, "middleware")
	if err := ensureDir(middlewareDir); err != nil {
		return err
	}

	// Template data with project module for imports
	data := struct {
		ProjectModule string
	}{
		ProjectModule: projectModule,
	}

	// Generate recovery.go with panic recovery middleware
	recoveryRaw, err := renderTemplate("templates/middleware_recovery.go.tmpl", data)
	if err != nil {
		return err
	}

	recoveryPath := filepath.Join(middlewareDir, "recovery.go")
	if err := writeGoFile(recoveryPath, recoveryRaw); err != nil {
		return err
	}

	// Generate errors.go with error rendering helpers
	errorsRaw, err := renderTemplate("templates/middleware_errors.go.tmpl", data)
	if err != nil {
		return err
	}

	errorsPath := filepath.Join(middlewareDir, "errors.go")
	if err := writeGoFile(errorsPath, errorsRaw); err != nil {
		return err
	}

	return nil
}
