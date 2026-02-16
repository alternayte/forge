package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateModels generates Go model files for all resources.
func GenerateModels(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create models directory
	modelsDir := filepath.Join(outputDir, "models")
	if err := ensureDir(modelsDir); err != nil {
		return err
	}

	// Generate a file for each resource
	for _, resource := range resources {
		// Render template
		raw, err := renderTemplate("templates/model.go.tmpl", resource)
		if err != nil {
			return err
		}

		// Write formatted Go file
		outputPath := filepath.Join(modelsDir, snake(resource.Name)+".go")
		if err := writeGoFile(outputPath, raw); err != nil {
			return err
		}
	}

	return nil
}
