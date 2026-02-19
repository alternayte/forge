package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/errors"
	"github.com/alternayte/forge/internal/generator"
	"github.com/alternayte/forge/internal/parser"
	"github.com/alternayte/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate code from resource schemas",
		Long:  `Parses schema definitions from resources/ and generates Go models, Atlas HCL schemas, and test factories into gen/.`,
		RunE:  runGenerate,
	}
	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Find project root by looking for forge.toml
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
	}

	// Load config from forge.toml
	cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
	if err != nil {
		return fmt.Errorf("failed to load forge.toml: %w", err)
	}

	// Check that module is set
	if cfg.Project.Module == "" {
		return fmt.Errorf("project module not set in forge.toml. Add a module path under [project]")
	}

	// Parse schemas from resources/ directory
	resourcesDir := filepath.Join(projectRoot, "resources")
	result, err := parser.ParseDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}

	// Check for parse errors
	if len(result.Errors) > 0 {
		// Format and display all errors
		fmt.Println()
		for _, parseErr := range result.Errors {
			// Check if error is a Diagnostic for rich formatting
			if diag, ok := parseErr.(errors.Diagnostic); ok {
				fmt.Println(errors.Format(diag))
			} else {
				// Fallback to simple error message
				fmt.Println(ui.Error(parseErr.Error()))
			}
		}
		fmt.Println()
		os.Exit(1)
	}

	// Check if any resources were found
	if len(result.Resources) == 0 {
		fmt.Println()
		fmt.Println(ui.Info("No schema definitions found in resources/. Create a schema file first."))
		fmt.Println()
		return nil
	}

	// Clean gen/ directory before generating
	genDir := filepath.Join(projectRoot, "gen")
	if err := os.RemoveAll(genDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean gen/ directory: %w", err)
	}

	// Print header
	fmt.Println()
	fmt.Println(ui.Header("Generating code..."))
	fmt.Println()

	// Run generators
	err = generator.Generate(result.Resources, generator.GenerateConfig{
		OutputDir:     genDir,
		ProjectModule: cfg.Project.Module,
		ProjectRoot:   projectRoot,
	})
	if err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	// Count generated files and display results
	modelsDir := filepath.Join(genDir, "models")
	if dirExists(modelsDir) {
		modelCount, _ := countFilesInDir(modelsDir)
		fmt.Println(ui.Success(fmt.Sprintf("gen/models/ — %d model %s", modelCount, pluralize("file", modelCount))))
	}

	atlasSchemaPath := filepath.Join(genDir, "atlas", "schema.hcl")
	if fileExists(atlasSchemaPath) {
		fmt.Println(ui.Success("gen/atlas/schema.hcl — database schema"))
	}

	factoriesDir := filepath.Join(genDir, "factories")
	if dirExists(factoriesDir) {
		factoryCount, _ := countFilesInDir(factoriesDir)
		fmt.Println(ui.Success(fmt.Sprintf("gen/factories/ — %d factory %s", factoryCount, pluralize("file", factoryCount))))
	}

	fmt.Println()

	// Print summary with timing
	duration := time.Since(startTime)
	summary := fmt.Sprintf("Generated %d %s in %dms", len(result.Resources), pluralize("resource", len(result.Resources)), duration.Milliseconds())
	fmt.Println("  " + ui.DimStyle.Render(summary))
	fmt.Println()

	return nil
}

// countFilesInDir counts files in a directory (non-recursive).
func countFilesInDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count, nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
