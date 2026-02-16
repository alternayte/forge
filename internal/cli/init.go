package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/forge-framework/forge/internal/scaffold"
	"github.com/forge-framework/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new Forge project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectName, projectPath string
			createNewDir := false

			// Determine project name and path
			if len(args) > 0 {
				// Argument provided: create new directory
				projectName = args[0]
				projectPath = args[0]
				createNewDir = true
			} else {
				// No argument: use current directory
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				projectName = filepath.Base(cwd)
				projectPath = "."
			}

			// Check if forge.toml already exists
			forgeTomlPath := filepath.Join(projectPath, "forge.toml")
			if _, err := os.Stat(forgeTomlPath); err == nil {
				return fmt.Errorf("project already initialized (forge.toml exists)")
			}

			// Prepare project data
			data := scaffold.ProjectData{
				Name:                 projectName,
				Module:               scaffold.InferModule(projectName),
				GoVersion:            scaffold.GetGoVersion(),
				ExampleResource:      "product",
				ExampleResourceTitle: "Product",
			}

			// Create project structure
			if err := scaffold.CreateProject(projectPath, data); err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			// Initialize git repository if new directory was created
			if createNewDir {
				cmd := exec.Command("git", "init")
				cmd.Dir = projectPath
				if err := cmd.Run(); err != nil {
					// Non-fatal error - continue
					fmt.Fprintf(os.Stderr, "%s Failed to initialize git repository\n", ui.WarnIcon)
				}
			}

			// Print styled success output
			fmt.Println(ui.Success(fmt.Sprintf("Created project %s", projectName)))
			fmt.Println()

			// List created files in grouped format
			files := []string{
				"forge.toml",
				"main.go",
				"go.mod",
				".gitignore",
				"README.md",
				filepath.Join("resources", "product", "schema.go"),
			}

			fmt.Println(ui.Grouped("Files:", files))
			fmt.Println()

			// Show next steps
			nextSteps := []string{}
			if createNewDir {
				nextSteps = append(nextSteps, fmt.Sprintf("cd %s", projectName))
			}
			nextSteps = append(nextSteps, "forge generate")

			fmt.Println(ui.Grouped("Next steps:", nextSteps))

			return nil
		},
	}
}
