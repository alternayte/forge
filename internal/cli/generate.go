package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/errors"
	"github.com/alternayte/forge/internal/generator"
	"github.com/alternayte/forge/internal/parser"
	"github.com/alternayte/forge/internal/ui"
	"github.com/alternayte/forge/internal/watcher"
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

	// Compile generated templ files in gen/html/ (layout, primitives)
	genHTMLDir := filepath.Join(genDir, "html")
	if err := runTemplGenerate(projectRoot, genHTMLDir); err != nil {
		fmt.Println(ui.Info(fmt.Sprintf("Note: templ generate skipped for gen/html: %v", err)))
	}

	// Scaffold each resource's views and handlers (scaffold-once: skips existing files)
	for _, resource := range result.Resources {
		scaffoldResult, err := generator.ScaffoldResource(resource, projectRoot, cfg.Project.Module)
		if err != nil {
			return fmt.Errorf("scaffold %s failed: %w", resource.Name, err)
		}
		for _, f := range scaffoldResult.Created {
			fmt.Println(ui.Success(fmt.Sprintf("resources/%s/%s", strings.ToLower(resource.Name), f)))
		}
		if len(scaffoldResult.Created) > 0 {
			viewsDir := filepath.Join(projectRoot, "resources", strings.ToLower(resource.Name), "views")
			if err := runTemplGenerate(projectRoot, viewsDir); err != nil {
				fmt.Println(ui.Info(fmt.Sprintf("Note: templ generate skipped for %s: %v", resource.Name, err)))
			}
		}
	}

	// Scaffold Tailwind input CSS (scaffold-once: skips if exists)
	if err := watcher.ScaffoldTailwindInput(projectRoot); err != nil {
		fmt.Println(ui.Info(fmt.Sprintf("Note: could not scaffold Tailwind input CSS: %v", err)))
	}

	// Compile Tailwind CSS if the binary is installed
	tailwindBin := filepath.Join(projectRoot, ".forge", "bin", "tailwindcss")
	if fileExists(tailwindBin) {
		if err := watcher.RunTailwind(projectRoot); err != nil {
			fmt.Println(ui.Info(fmt.Sprintf("Note: Tailwind CSS compilation failed: %v", err)))
		}
	} else {
		fmt.Println(ui.Info("Tailwind CSS binary not found. Run 'forge tool sync' to download it."))
	}

	// Inject replace directive for local forge development and tidy modules
	if err := injectForgeReplace(projectRoot); err != nil {
		fmt.Println(ui.Info(fmt.Sprintf("Note: could not inject replace directive: %v", err)))
	}
	if err := runGoModTidy(projectRoot); err != nil {
		fmt.Println(ui.Info(fmt.Sprintf("Note: go mod tidy failed: %v", err)))
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

// injectForgeReplace adds a replace directive for the forge module if running
// from source. This enables `go build` in generated projects during development
// before the forge module is published to a module proxy.
func injectForgeReplace(projectRoot string) error {
	forgeRoot, err := findForgeSourceRoot()
	if err != nil {
		// Not running from source — skip silently
		return nil
	}

	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	if bytes.Contains(content, []byte("replace github.com/alternayte/forge")) {
		return nil // already present
	}

	cmd := exec.Command("go", "mod", "edit",
		"-replace", fmt.Sprintf("github.com/alternayte/forge=%s", forgeRoot))
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod edit -replace: %s: %w", out, err)
	}
	return nil
}

// findForgeSourceRoot locates the forge module root by walking up from the
// current executable. Returns an error if the binary isn't inside a forge
// source tree (e.g., installed via `go install`).
func findForgeSourceRoot() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(exe)
	for {
		goMod := filepath.Join(dir, "go.mod")
		if content, err := os.ReadFile(goMod); err == nil {
			if bytes.Contains(content, []byte("module github.com/alternayte/forge")) {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("forge source root not found")
		}
		dir = parent
	}
}

// runGoModTidy runs `go mod tidy` in the project directory to resolve
// transitive dependencies after code generation.
func runGoModTidy(projectRoot string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %s: %w", out, err)
	}
	return nil
}
