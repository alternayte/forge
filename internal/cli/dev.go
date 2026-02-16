package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/forge-framework/forge/internal/config"
	"github.com/forge-framework/forge/internal/ui"
	"github.com/forge-framework/forge/internal/watcher"
	"github.com/spf13/cobra"
)

func newDevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development server with file watching",
		Long: `Starts a development server that watches for file changes in resources/
and internal/ directories. When schema files change, code is automatically
regenerated. Watches .go, .templ, .sql, and .css files.

Ctrl+C to stop.`,
		RunE: runDev,
	}
	return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
	// Find project root (forge.toml)
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Println(ui.Error("Cannot find forge.toml - are you in a Forge project?"))
		return err
	}

	// Load config
	configPath := filepath.Join(projectRoot, "forge.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Failed to load config: %v", err)))
		return err
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create and start dev server
	server := watcher.NewDevServer(projectRoot, cfg)
	if err := server.Start(ctx); err != nil {
		// Only return error if it's not a context cancellation
		if ctx.Err() == nil {
			return err
		}
	}

	return nil
}
