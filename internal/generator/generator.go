package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateConfig holds configuration for the generator.
type GenerateConfig struct {
	OutputDir     string // Output directory for generated files
	ProjectModule string // Go module path of the generated project
	ProjectRoot   string // Project root directory (parent of OutputDir)
}

// Generate orchestrates all code generation from parsed resources.
func Generate(resources []parser.ResourceIR, cfg GenerateConfig) error {
	// Generate model types
	if err := GenerateModels(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate Atlas HCL schema
	if err := GenerateAtlasSchema(resources, cfg.OutputDir); err != nil {
		return err
	}

	// Generate test factories
	if err := GenerateFactories(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate validation functions
	if err := GenerateValidation(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate query builder mods
	if err := GenerateQueries(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate pagination utilities
	if err := GeneratePagination(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate transaction wrapper
	if err := GenerateTransaction(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate SQLC configuration
	if err := GenerateSQLCConfig(resources, cfg.OutputDir, cfg.ProjectModule, cfg.ProjectRoot); err != nil {
		return err
	}

	// Generate error types and DB error mapping
	if err := GenerateErrors(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate action interfaces and default implementations
	if err := GenerateActions(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	// Generate middleware (panic recovery, error rendering)
	if err := GenerateMiddleware(resources, cfg.OutputDir, cfg.ProjectModule); err != nil {
		return err
	}

	return nil
}

// renderTemplate parses and executes a template from TemplatesFS.
func renderTemplate(tmplName string, data interface{}) ([]byte, error) {
	// Read template content from embedded filesystem
	content, err := TemplatesFS.ReadFile(tmplName)
	if err != nil {
		return nil, err
	}

	// Parse template with FuncMap
	tmpl, err := template.New(filepath.Base(tmplName)).Funcs(BuildFuncMap()).Parse(string(content))
	if err != nil {
		return nil, err
	}

	// Execute template to buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// writeGoFile writes a Go file with formatting and import management.
func writeGoFile(outputPath string, raw []byte) error {
	// Format the source
	formatted, err := FormatGoSource(outputPath, raw)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := ensureDir(filepath.Dir(outputPath)); err != nil {
		return err
	}

	// Write to disk
	return os.WriteFile(outputPath, formatted, 0644)
}

// writeRawFile writes a non-Go file directly without formatting.
func writeRawFile(outputPath string, raw []byte) error {
	// Ensure parent directory exists
	if err := ensureDir(filepath.Dir(outputPath)); err != nil {
		return err
	}

	// Write to disk
	return os.WriteFile(outputPath, raw, 0644)
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
