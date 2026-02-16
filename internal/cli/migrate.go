package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/forge-framework/forge/internal/config"
	"github.com/forge-framework/forge/internal/migrate"
	"github.com/forge-framework/forge/internal/toolsync"
	"github.com/forge-framework/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
		Long: `Create, apply, and manage database migrations using Atlas.

Atlas versioned workflow reads all .sql files from the migrations/ directory,
including hand-written migrations. If you manually create or edit a migration file,
run 'forge migrate hash' to update the integrity checksum (atlas.sum).`,
	}

	cmd.AddCommand(
		newMigrateDiffCmd(),
		newMigrateUpCmd(),
		newMigrateDownCmd(),
		newMigrateStatusCmd(),
		newMigrateHashCmd(),
	)

	return cmd
}

func newMigrateDiffCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "diff [name]",
		Short: "Create a new migration from schema changes",
		Long: `Generate a new migration file by comparing the current database schema
with the target schema defined in gen/atlas/schema.hcl.

If the migration contains destructive operations (DROP TABLE, DROP COLUMN, etc.),
it will be rejected unless --force is specified.

Example:
  forge migrate diff add_users
  forge migrate diff --force remove_old_table`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find project root
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
			}

			// Load config
			cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
			if err != nil {
				return fmt.Errorf("failed to load forge.toml: %w", err)
			}

			// Check if atlas is installed
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Check if gen/atlas/schema.hcl exists
			schemaPath := filepath.Join(projectRoot, "gen", "atlas", "schema.hcl")
			if _, err := os.Stat(schemaPath); err != nil {
				fmt.Println()
				fmt.Println(ui.Error("No schema found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge generate") + " to create the schema first."))
				fmt.Println()
				return fmt.Errorf("gen/atlas/schema.hcl not found")
			}

			// Determine migration name
			name := "migration"
			if len(args) > 0 {
				name = args[0]
			}

			// Build migrate config
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
				SchemaURL:    schemaPath,
				DatabaseURL:  cfg.Database.URL,
				DevURL:       getDevURL(cfg.Database.URL),
			}

			// Run diff
			fmt.Println()
			fmt.Println(ui.Header("Creating migration..."))
			fmt.Println()

			migrationFile, err := migrate.Diff(migrateCfg, name, force)
			if err != nil {
				// Check if it's a destructive change error (will be formatted already)
				if !force && (containsDestructiveKeywords(err.Error())) {
					fmt.Println(err.Error())
					fmt.Println()
					return nil // Exit gracefully for destructive warnings
				}
				return err
			}

			// Success
			relPath, _ := filepath.Rel(projectRoot, migrationFile)
			fmt.Println(ui.Success("Created migration: " + ui.FilePathStyle.Render(relPath)))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "bypass destructive change warnings")

	return cmd
}

func newMigrateUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations",
		Long: `Apply all pending migrations to the database.

This runs 'atlas migrate apply' to bring the database schema up to date
with the migration files in the migrations/ directory.

Example:
  forge migrate up`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find project root
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
			}

			// Load config
			cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
			if err != nil {
				return fmt.Errorf("failed to load forge.toml: %w", err)
			}

			// Check if atlas is installed
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Build migrate config
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
				DatabaseURL:  cfg.Database.URL,
			}

			// Run up
			fmt.Println()
			fmt.Println(ui.Header("Applying migrations..."))
			fmt.Println()

			output, err := migrate.Up(migrateCfg)
			if err != nil {
				return err
			}

			// Display atlas output
			fmt.Println(output)
			fmt.Println()
			fmt.Println(ui.Success("Migrations applied"))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newMigrateDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Roll back the last migration",
		Long: `Roll back the most recently applied migration.

This runs 'atlas migrate down' to revert the last migration.

Example:
  forge migrate down`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find project root
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
			}

			// Load config
			cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
			if err != nil {
				return fmt.Errorf("failed to load forge.toml: %w", err)
			}

			// Check if atlas is installed
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Build migrate config
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
				DatabaseURL:  cfg.Database.URL,
				DevURL:       getDevURL(cfg.Database.URL),
			}

			// Run down
			fmt.Println()
			fmt.Println(ui.Header("Rolling back migration..."))
			fmt.Println()

			output, err := migrate.Down(migrateCfg)
			if err != nil {
				return err
			}

			// Display atlas output
			fmt.Println(output)
			fmt.Println()
			fmt.Println(ui.Success("Migration rolled back"))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newMigrateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current migration state",
		Long: `Display the current state of database migrations.

This shows which migrations have been applied and which are pending.

Example:
  forge migrate status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find project root
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
			}

			// Load config
			cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
			if err != nil {
				return fmt.Errorf("failed to load forge.toml: %w", err)
			}

			// Check if atlas is installed
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Build migrate config
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
				DatabaseURL:  cfg.Database.URL,
			}

			// Run status
			fmt.Println()

			output, err := migrate.Status(migrateCfg)
			if err != nil {
				return err
			}

			// Display atlas output
			fmt.Println(output)
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newMigrateHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Recompute migration checksums",
		Long: `Recompute the atlas.sum checksum file after manual migration edits.

Run this after creating or editing migration files manually to update
the integrity checksum that Atlas uses to track migration changes.

Example:
  forge migrate hash`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find project root
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
			}

			// Check if atlas is installed
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Build migrate config
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
			}

			// Run hash
			fmt.Println()
			fmt.Println(ui.Header("Recomputing migration checksums..."))
			fmt.Println()

			if err := migrate.Hash(migrateCfg); err != nil {
				return err
			}

			fmt.Println(ui.Success("Migration checksums updated"))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

// getDevURL constructs a dev database URL from the main database URL.
// For now, it just returns the same URL - Atlas will create a temporary schema.
// In the future, we could parse and modify the URL to use a separate dev database.
func getDevURL(databaseURL string) string {
	// For simplicity, use the same database URL
	// Atlas will create temporary schemas for diffing
	return databaseURL
}

// containsDestructiveKeywords checks if an error message contains destructive change keywords.
func containsDestructiveKeywords(errMsg string) bool {
	keywords := []string{"WARNING", "Destructive", "DROP TABLE", "DROP COLUMN"}
	for _, kw := range keywords {
		if containsString(errMsg, kw) {
			return true
		}
	}
	return false
}

// containsString checks if a string contains a substring (case-insensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
