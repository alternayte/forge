package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateValidation generates validation functions for all resources.
func GenerateValidation(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create validation directory
	validationDir := filepath.Join(outputDir, "validation")
	if err := ensureDir(validationDir); err != nil {
		return err
	}

	// Generate shared types.go file first
	typesRaw, err := renderTemplate("templates/validation_types.go.tmpl", nil)
	if err != nil {
		return err
	}

	typesPath := filepath.Join(validationDir, "types.go")
	if err := writeGoFile(typesPath, typesRaw); err != nil {
		return err
	}

	// Generate a validation file for each resource
	for _, resource := range resources {
		// Prepare template data with ProjectModule
		data := struct {
			parser.ResourceIR
			ProjectModule string
		}{
			ResourceIR:    resource,
			ProjectModule: projectModule,
		}

		// Render template
		raw, err := renderTemplate("templates/validation.go.tmpl", data)
		if err != nil {
			return err
		}

		// Write formatted Go file
		outputPath := filepath.Join(validationDir, snake(resource.Name)+"_validation.go")
		if err := writeGoFile(outputPath, raw); err != nil {
			return err
		}
	}

	return nil
}
