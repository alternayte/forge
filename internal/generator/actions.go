package generator

import (
	"path/filepath"

	"github.com/alternayte/forge/internal/parser"
)

// GenerateActions generates action interfaces and default implementations for all resources.
func GenerateActions(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create actions directory
	actionsDir := filepath.Join(outputDir, "actions")
	if err := ensureDir(actionsDir); err != nil {
		return err
	}

	// Generate shared types.go file first
	typesData := struct {
		ProjectModule string
	}{
		ProjectModule: projectModule,
	}

	typesRaw, err := renderTemplate("templates/actions_types.go.tmpl", typesData)
	if err != nil {
		return err
	}

	typesPath := filepath.Join(actionsDir, "types.go")
	if err := writeGoFile(typesPath, typesRaw); err != nil {
		return err
	}

	// Generate defaults.go — NewDefaultRegistry(db) pre-populated with all defaults
	defaultsData := struct {
		Resources     []parser.ResourceIR
		ProjectModule string
	}{
		Resources:     resources,
		ProjectModule: projectModule,
	}

	defaultsRaw, err := renderTemplate("templates/actions_defaults.go.tmpl", defaultsData)
	if err != nil {
		return err
	}

	defaultsPath := filepath.Join(actionsDir, "defaults.go")
	if err := writeGoFile(defaultsPath, defaultsRaw); err != nil {
		return err
	}

	// Generate an actions file for each resource
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
		raw, err := renderTemplate("templates/actions.go.tmpl", data)
		if err != nil {
			return err
		}

		// Write formatted Go file
		outputPath := filepath.Join(actionsDir, snake(resource.Name)+".go")
		if err := writeGoFile(outputPath, raw); err != nil {
			return err
		}
	}

	return nil
}
