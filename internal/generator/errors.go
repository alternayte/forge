package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateErrors generates error types and database error mapping.
// Signature matches other generators for consistency even though errors are not per-resource.
func GenerateErrors(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create errors directory
	errorsDir := filepath.Join(outputDir, "errors")
	if err := ensureDir(errorsDir); err != nil {
		return err
	}

	// Generate errors.go with Error type and constructors
	errorsRaw, err := renderTemplate("templates/errors.go.tmpl", nil)
	if err != nil {
		return err
	}

	errorsPath := filepath.Join(errorsDir, "errors.go")
	if err := writeGoFile(errorsPath, errorsRaw); err != nil {
		return err
	}

	// Generate db_mapping.go with MapDBError function
	dbMappingRaw, err := renderTemplate("templates/errors_db_mapping.go.tmpl", nil)
	if err != nil {
		return err
	}

	dbMappingPath := filepath.Join(errorsDir, "db_mapping.go")
	if err := writeGoFile(dbMappingPath, dbMappingRaw); err != nil {
		return err
	}

	return nil
}
