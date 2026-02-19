package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// FactoryTemplateData holds data for rendering the factory template.
type FactoryTemplateData struct {
	Name          string
	Fields        []parser.FieldIR
	Options       parser.ResourceOptionsIR
	HasTimestamps bool
	ProjectModule string
}

// GenerateFactories generates test factory files for each resource.
// Factories provide builder-pattern helpers for creating test data.
func GenerateFactories(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	for _, resource := range resources {
		// Prepare template data
		data := FactoryTemplateData{
			Name:          resource.Name,
			Fields:        resource.Fields,
			Options:       resource.Options,
			HasTimestamps: resource.HasTimestamps,
			ProjectModule: projectModule,
		}

		// Render the factory template
		content, err := renderTemplate("templates/factory.go.tmpl", data)
		if err != nil {
			return err
		}

		// Write to output directory (factories/{snake_name}.go)
		outputPath := filepath.Join(outputDir, "factories", snake(resource.Name)+".go")
		if err := writeGoFile(outputPath, content); err != nil {
			return err
		}
	}

	return nil
}
