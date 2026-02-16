package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateSQLCConfig generates a sqlc.yaml configuration file for the project.
// This provides an escape hatch for developers to write custom SQL queries.
func GenerateSQLCConfig(resources []parser.ResourceIR, outputDir string, projectModule string, projectRoot string) error {
	// Render template (static YAML, no resource-specific data)
	raw, err := renderTemplate("templates/sqlc.yaml.tmpl", nil)
	if err != nil {
		return err
	}

	// Write to project root (not gen/)
	outputPath := filepath.Join(projectRoot, "sqlc.yaml")
	if err := writeRawFile(outputPath, raw); err != nil {
		return err
	}

	return nil
}

// GenerateTransaction generates transaction wrapper functions with River support.
// This generates a single file (not per-resource) in gen/forge/.
func GenerateTransaction(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create forge directory
	forgeDir := filepath.Join(outputDir, "forge")
	if err := ensureDir(forgeDir); err != nil {
		return err
	}

	// Render template (no resource-specific data needed)
	raw, err := renderTemplate("templates/transaction.go.tmpl", nil)
	if err != nil {
		return err
	}

	// Write formatted Go file
	outputPath := filepath.Join(forgeDir, "transaction.go")
	if err := writeGoFile(outputPath, raw); err != nil {
		return err
	}

	return nil
}
