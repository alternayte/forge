package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/stringutil"
	"github.com/alternayte/forge/internal/generator"
	"github.com/alternayte/forge/internal/parser"
	"github.com/alternayte/forge/internal/toolsync"
	"github.com/alternayte/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newGenerateResourceCmd() *cobra.Command {
	var diffFlag bool

	cmd := &cobra.Command{
		Use:   "resource <name>",
		Short: "Scaffold HTML form, list, and detail views for a resource",
		Long: `Scaffold a resource's views and handlers into resources/<name>/.

Only files that do not already exist are written — existing developer
customizations are never overwritten. Use --diff to preview what would
change without writing any files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateResource(cmd, args, diffFlag)
		},
	}

	cmd.Flags().BoolVar(&diffFlag, "diff", false, "Show diff between current and freshly scaffolded views")

	return cmd
}

func runGenerateResource(cmd *cobra.Command, args []string, diff bool) error {
	resourceName := args[0]

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

	// Parse schemas from resources/ directory
	resourcesDir := filepath.Join(projectRoot, "resources")
	result, err := parser.ParseDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}

	// Check for parse errors
	if len(result.Errors) > 0 {
		for _, parseErr := range result.Errors {
			fmt.Println(ui.Error(parseErr.Error()))
		}
		return fmt.Errorf("schema errors found")
	}

	// Find the matching resource by name (case-insensitive)
	var target *parser.ResourceIR
	var available []string
	for i := range result.Resources {
		available = append(available, result.Resources[i].Name)
		if strings.EqualFold(result.Resources[i].Name, resourceName) {
			r := result.Resources[i]
			target = &r
		}
	}

	if target == nil {
		if len(available) == 0 {
			return fmt.Errorf("resource %q not found: no resources defined in resources/", resourceName)
		}
		return fmt.Errorf("resource %q not found. Available resources: %s", resourceName, strings.Join(available, ", "))
	}

	if diff {
		// Show diff without writing
		output, err := generator.DiffResource(*target, projectRoot, cfg.Project.Module)
		if err != nil {
			return fmt.Errorf("diff failed: %w", err)
		}
		fmt.Print(output)
		return nil
	}

	// Scaffold the resource
	scaffoldResult, err := generator.ScaffoldResource(*target, projectRoot, cfg.Project.Module)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}

	// Print created/skipped files
	for _, f := range scaffoldResult.Created {
		fmt.Println(ui.Success(fmt.Sprintf("created  resources/%s/%s", strings.ToLower(target.Name), f)))
	}
	for _, f := range scaffoldResult.Skipped {
		fmt.Println(ui.Info(fmt.Sprintf("skipped  resources/%s/%s (already exists)", strings.ToLower(target.Name), f)))
	}

	// Run templ generate on the resource views directory to compile .templ files to _templ.go
	if len(scaffoldResult.Created) > 0 {
		viewsDir := filepath.Join(projectRoot, "resources", strings.ToLower(target.Name), "views")
		if err := runTemplGenerate(projectRoot, viewsDir); err != nil {
			// Non-fatal: templ may not be installed; print warning and continue
			fmt.Println(ui.Info(fmt.Sprintf("Note: templ generate skipped: %v", err)))
		}
	}

	// Print summary
	created := len(scaffoldResult.Created)
	skipped := len(scaffoldResult.Skipped)
	fmt.Println()
	fmt.Printf("  %s\n", ui.DimStyle.Render(fmt.Sprintf(
		"Scaffolded %d %s for %s (%d skipped)",
		created, stringutil.Pluralize("file", created), target.Name, skipped,
	)))
	fmt.Println()

	return nil
}

// runTemplGenerate runs "templ generate <dir>" to compile .templ files to _templ.go.
// It tries the forge-managed binary first, then falls back to PATH.
func runTemplGenerate(projectRoot, viewsDir string) error {
	// Try forge-managed templ binary first
	forgeBin := filepath.Join(projectRoot, ".forge", "bin")
	forgeTemplBin := toolsync.ToolBinPath(forgeBin, "templ")

	var templBin string
	if toolsync.IsToolInstalled(forgeBin, "templ") {
		templBin = forgeTemplBin
	} else {
		// Fall back to PATH
		var err error
		templBin, err = exec.LookPath("templ")
		if err != nil {
			return fmt.Errorf("templ binary not found in .forge/bin or PATH. Run 'forge tool sync' first")
		}
	}

	// Check views directory exists
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		// No views directory (no .templ files created) — nothing to compile
		return nil
	}

	cmd := exec.Command(templBin, "generate", viewsDir)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
