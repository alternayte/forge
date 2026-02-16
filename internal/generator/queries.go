package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateQueries generates query builder mod functions for all resources.
func GenerateQueries(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create queries directory
	queriesDir := filepath.Join(outputDir, "queries")
	if err := ensureDir(queriesDir); err != nil {
		return err
	}

	// Generate a query file for each resource
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
		raw, err := renderTemplate("templates/queries.go.tmpl", data)
		if err != nil {
			return err
		}

		// Write formatted Go file
		outputPath := filepath.Join(queriesDir, snake(resource.Name)+"_queries.go")
		if err := writeGoFile(outputPath, raw); err != nil {
			return err
		}
	}

	return nil
}
