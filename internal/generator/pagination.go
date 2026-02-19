package generator

import (
	"path/filepath"

	"github.com/alternayte/forge/internal/parser"
)

// GeneratePagination generates pagination utilities for cursor and offset-based pagination.
// This generates a single file (not per-resource) since pagination logic is generic.
func GeneratePagination(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create queries directory
	queriesDir := filepath.Join(outputDir, "queries")
	if err := ensureDir(queriesDir); err != nil {
		return err
	}

	// Render template (no resource-specific data needed)
	raw, err := renderTemplate("templates/pagination.go.tmpl", nil)
	if err != nil {
		return err
	}

	// Write formatted Go file
	outputPath := filepath.Join(queriesDir, "pagination.go")
	if err := writeGoFile(outputPath, raw); err != nil {
		return err
	}

	return nil
}
