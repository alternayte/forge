package cli

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/migrate"
	"github.com/alternayte/forge/internal/toolsync"
	"github.com/alternayte/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage development database",
		Long: `Manage the development database with create, drop, console, seed, and reset commands.

These commands wrap PostgreSQL client tools (createdb, dropdb, psql) and integrate with
Atlas migrations and user-defined seed files for complete database lifecycle management.`,
	}

	cmd.AddCommand(
		newDBCreateCmd(),
		newDBDropCmd(),
		newDBConsoleCmd(),
		newDBSeedCmd(),
		newDBResetCmd(),
	)

	return cmd
}

func newDBCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create the development database",
		Long: `Create the PostgreSQL database specified in forge.toml.

This runs 'createdb' with connection parameters extracted from the database URL.
If the database already exists, this is a no-op.

Example:
  forge db create`,
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

			// Parse database URL
			host, port, dbName, err := parseDatabaseURL(cfg.Database.URL)
			if err != nil {
				return fmt.Errorf("failed to parse database URL: %w", err)
			}

			// Create database
			fmt.Println()
			fmt.Println(ui.Header("Creating database..."))
			fmt.Println()

			if err := createDatabase(host, port, dbName); err != nil {
				return err
			}

			fmt.Println(ui.Success("Database created: " + dbName))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newDBDropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drop the development database",
		Long: `Drop the PostgreSQL database specified in forge.toml.

This runs 'dropdb' with connection parameters extracted from the database URL.
If the database doesn't exist, this is a no-op.

WARNING: This permanently deletes all data in the database.

Example:
  forge db drop`,
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

			// Parse database URL
			host, port, dbName, err := parseDatabaseURL(cfg.Database.URL)
			if err != nil {
				return fmt.Errorf("failed to parse database URL: %w", err)
			}

			// Drop database
			fmt.Println()
			fmt.Println(ui.Header("Dropping database..."))
			fmt.Println()

			if err := dropDatabase(host, port, dbName); err != nil {
				return err
			}

			fmt.Println(ui.Success("Database dropped: " + dbName))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newDBConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open a psql session to the database",
		Long: `Open an interactive psql session connected to the database specified in forge.toml.

This runs 'psql' and hands off terminal control for interactive SQL queries.

Example:
  forge db console`,
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

			// Run psql with interactive terminal
			fmt.Println()
			fmt.Println(ui.Header("Connecting to database..."))
			fmt.Println()

			psqlCmd := exec.Command("psql", cfg.Database.URL)
			psqlCmd.Stdin = os.Stdin
			psqlCmd.Stdout = os.Stdout
			psqlCmd.Stderr = os.Stderr

			if err := psqlCmd.Run(); err != nil {
				// Check if psql is not installed
				if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
					fmt.Println(ui.Error("PostgreSQL client tools not found."))
					fmt.Println(ui.Info("Install postgresql-client or postgresql to get createdb, dropdb, and psql."))
					fmt.Println()
					return fmt.Errorf("psql command not found")
				}
				return fmt.Errorf("psql failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func newDBSeedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Run database seed file",
		Long: `Run the db/seed.go file to populate development data.

The seed file is a user-written Go file that uses generated factories or raw SQL
to create test data. Forge passes the DATABASE_URL environment variable so the
seed file knows which database to connect to.

Example:
  forge db seed`,
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

			// Run seed
			fmt.Println()
			fmt.Println(ui.Header("Running seed..."))
			fmt.Println()

			if err := runSeed(projectRoot, cfg); err != nil {
				return err
			}

			fmt.Println(ui.Success("Database seeded"))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newDBResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Drop, create, migrate, and seed the database",
		Long: `Reset the database to a clean state by running the full sequence:
1. Drop the existing database
2. Create a fresh database
3. Apply all migrations
4. Run seed file (if it exists)

This is useful for resetting your development environment to a known state.

WARNING: This permanently deletes all data in the database.

Example:
  forge db reset`,
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

			// Parse database URL
			host, port, dbName, err := parseDatabaseURL(cfg.Database.URL)
			if err != nil {
				return fmt.Errorf("failed to parse database URL: %w", err)
			}

			// Check if atlas is installed (needed for migrations)
			atlasBin := toolsync.ToolBinPath(filepath.Join(projectRoot, ".forge", "bin"), "atlas")
			if !toolsync.IsToolInstalled(filepath.Join(projectRoot, ".forge", "bin"), "atlas") {
				fmt.Println()
				fmt.Println(ui.Error("Atlas not found."))
				fmt.Println(ui.Info("Run " + ui.CommandStyle.Render("forge tool sync") + " to download required tools."))
				fmt.Println()
				return fmt.Errorf("atlas binary not installed")
			}

			// Execute reset sequence
			fmt.Println()
			fmt.Println(ui.Header("Resetting database..."))
			fmt.Println()

			// Step 1: Drop database
			fmt.Println(ui.Info("  Dropping database..."))
			if err := dropDatabase(host, port, dbName); err != nil {
				return err
			}
			fmt.Println(ui.Success("  Database dropped"))
			fmt.Println()

			// Step 2: Create database
			fmt.Println(ui.Info("  Creating database..."))
			if err := createDatabase(host, port, dbName); err != nil {
				return err
			}
			fmt.Println(ui.Success("  Database created"))
			fmt.Println()

			// Step 3: Apply migrations
			fmt.Println(ui.Info("  Applying migrations..."))
			migrateCfg := migrate.Config{
				AtlasBin:     atlasBin,
				MigrationDir: filepath.Join(projectRoot, "migrations"),
				DatabaseURL:  cfg.Database.URL,
			}

			output, err := migrate.Up(migrateCfg)
			if err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}
			// Display atlas output if not empty
			if strings.TrimSpace(output) != "" {
				fmt.Println(output)
			}
			fmt.Println(ui.Success("  Migrations applied"))
			fmt.Println()

			// Step 4: Run seed (if exists)
			seedPath := filepath.Join(projectRoot, "db", "seed.go")
			if _, err := os.Stat(seedPath); err == nil {
				fmt.Println(ui.Info("  Running seed..."))
				if err := runSeed(projectRoot, cfg); err != nil {
					return fmt.Errorf("failed to run seed: %w", err)
				}
				fmt.Println(ui.Success("  Database seeded"))
				fmt.Println()
			} else {
				fmt.Println(ui.Info("  No seed file found (skipping)"))
				fmt.Println()
			}

			fmt.Println(ui.Success("Database reset complete"))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

// Helper functions

// parseDatabaseURL extracts host, port, and database name from a PostgreSQL URL.
// Handles both postgres:// and postgresql:// schemes.
// Example: postgres://localhost:5432/my-app?sslmode=disable -> (localhost, 5432, my-app)
func parseDatabaseURL(rawURL string) (host, port, dbName string, err error) {
	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid database URL: %w", err)
	}

	// Check scheme
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", "", "", fmt.Errorf("unsupported database scheme: %s (expected postgres:// or postgresql://)", u.Scheme)
	}

	// Extract host and port
	host = u.Hostname()
	if host == "" {
		host = "localhost"
	}

	port = u.Port()
	if port == "" {
		port = "5432"
	}

	// Extract database name from path (strip leading slash)
	dbName = strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", "", "", fmt.Errorf("database name not found in URL path")
	}

	return host, port, dbName, nil
}

// createDatabase creates a PostgreSQL database using the createdb command.
func createDatabase(host, port, dbName string) error {
	cmd := exec.Command("createdb", "-h", host, "-p", port, dbName)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if database already exists (not an error)
		if strings.Contains(stderr.String(), "already exists") {
			fmt.Println(ui.Info("Database already exists: " + dbName))
			return nil
		}

		// Check if createdb is not installed
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			fmt.Println(ui.Error("PostgreSQL client tools not found."))
			fmt.Println(ui.Info("Install postgresql-client or postgresql to get createdb, dropdb, and psql."))
			fmt.Println()
			return fmt.Errorf("createdb command not found")
		}

		return fmt.Errorf("createdb failed: %w\nOutput: %s", err, stderr.String())
	}

	return nil
}

// dropDatabase drops a PostgreSQL database using the dropdb command.
func dropDatabase(host, port, dbName string) error {
	cmd := exec.Command("dropdb", "--if-exists", "-h", host, "-p", port, dbName)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if dropdb is not installed
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			fmt.Println(ui.Error("PostgreSQL client tools not found."))
			fmt.Println(ui.Info("Install postgresql-client or postgresql to get createdb, dropdb, and psql."))
			fmt.Println()
			return fmt.Errorf("dropdb command not found")
		}

		return fmt.Errorf("dropdb failed: %w\nOutput: %s", err, stderr.String())
	}

	return nil
}

// runSeed runs the db/seed.go file with the DATABASE_URL environment variable set.
func runSeed(projectRoot string, cfg *config.Config) error {
	seedPath := filepath.Join(projectRoot, "db", "seed.go")

	// Check if seed file exists
	if _, err := os.Stat(seedPath); err != nil {
		fmt.Println(ui.Info("No seed file found at db/seed.go. Create one to populate development data."))
		return nil
	}

	// Run seed file
	cmd := exec.Command("go", "run", "./db/seed.go")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "DATABASE_URL="+cfg.Database.URL)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("seed failed: %w", err)
	}

	return nil
}
