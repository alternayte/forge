package generator

import (
	"path/filepath"

	"github.com/alternayte/forge/internal/parser"
)

// AtlasTemplateData holds data for rendering the Atlas schema template.
type AtlasTemplateData struct {
	Resources []parser.ResourceIR
}

// GenerateAtlasSchema generates an Atlas HCL schema file from parsed resources.
// The schema file represents the desired database state for Atlas CLI to diff against.
func GenerateAtlasSchema(resources []parser.ResourceIR, outputDir string) error {
	// Prepare template data
	data := AtlasTemplateData{
		Resources: resources,
	}

	// Render the Atlas HCL template
	content, err := renderTemplate("templates/atlas_schema.hcl.tmpl", data)
	if err != nil {
		return err
	}

	// Write to output directory (atlas/schema.hcl)
	outputPath := filepath.Join(outputDir, "atlas", "schema.hcl")
	return writeRawFile(outputPath, content)
}
