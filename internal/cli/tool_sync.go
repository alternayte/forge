package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alternayte/forge/internal/stringutil"
	"github.com/alternayte/forge/internal/toolsync"
	"github.com/alternayte/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newToolSyncCmd() *cobra.Command {
	var (
		toolsFlag []string
		forceFlag bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Download required tool binaries",
		Long: `Download required tool binaries to .forge/bin/ directory.

Forge relies on external tools (templ, sqlc, tailwind, atlas) for code generation
and development. This command downloads platform-appropriate binaries for your OS
and architecture.

By default, tools that are already installed are skipped. Use --force to
re-download all tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Detect platform
			platform := toolsync.DetectPlatform()
			if err := platform.Validate(); err != nil {
				return fmt.Errorf("unsupported platform: %w", err)
			}

			// Determine project root (find forge.toml walking up directories)
			projectRoot, err := findProjectRoot()
			if err != nil {
				// If no forge.toml found, use current directory
				projectRoot, _ = os.Getwd()
			}

			// Set destination directory
			destDir := filepath.Join(projectRoot, ".forge", "bin")

			// Get tool registry
			registry := toolsync.DefaultRegistry()

			// Filter tools if --tools flag provided
			var tools []toolsync.ToolDef
			if len(toolsFlag) > 0 {
				toolSet := make(map[string]bool)
				for _, name := range toolsFlag {
					toolSet[name] = true
				}

				for _, tool := range registry {
					if toolSet[tool.Name] {
						tools = append(tools, tool)
					}
				}

				// Validate all requested tools exist
				if len(tools) != len(toolsFlag) {
					return fmt.Errorf("some requested tools not found in registry")
				}
			} else {
				tools = registry
			}

			// Print header
			fmt.Println()
			fmt.Println(ui.Header("Syncing tools..."))
			fmt.Println()

			// Track results
			var (
				results  []string
				failures []string
			)

			// Process each tool
			for _, tool := range tools {
				alreadyInstalled := toolsync.IsToolInstalled(destDir, tool.BinaryName)

				if alreadyInstalled && !forceFlag {
					// Already installed, skip
					msg := fmt.Sprintf("%s %s", tool.Name, ui.DimStyle.Render("v"+tool.Version))
					results = append(results, ui.Success(msg+" "+ui.DimStyle.Render("(already installed)")))
					continue
				}

				// Download tool
				fmt.Print(ui.Info(fmt.Sprintf("Downloading %s v%s...", tool.Name, tool.Version)))
				fmt.Print("\r") // Prepare to overwrite progress line

				err := toolsync.DownloadTool(tool, platform, destDir, func(pct float64) {
					// Progress callback - could update spinner here
					// For now, we just let it download silently
				})

				if err != nil {
					errMsg := fmt.Sprintf("%s: %s", tool.Name, err.Error())
					failures = append(failures, errMsg)
					results = append(results, ui.Error(tool.Name+" "+ui.DimStyle.Render("(download failed)")))
					continue
				}

				// Success
				msg := fmt.Sprintf("%s %s", tool.Name, ui.DimStyle.Render("v"+tool.Version))
				results = append(results, ui.Success(msg))
			}

			// Print results
			for _, result := range results {
				fmt.Println(result)
			}

			fmt.Println()

			// Print summary
			if len(failures) > 0 {
				fmt.Println(ui.Error(fmt.Sprintf("%d tools failed to download", len(failures))))
				for _, failure := range failures {
					fmt.Println("    " + failure)
				}
				return fmt.Errorf("tool sync failed")
			}

			synced := len(tools) - len(failures)
			summary := fmt.Sprintf("%d %s synced", synced, stringutil.Pluralize("tool", synced))
			fmt.Println("  " + ui.DimStyle.Render(summary))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&toolsFlag, "tools", nil, "specific tools to sync (comma-separated)")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "re-download even if already installed")

	return cmd
}

// findProjectRoot walks up directories looking for forge.toml
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, "forge.toml")
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", fmt.Errorf("forge.toml not found")
		}
		dir = parent
	}
}

