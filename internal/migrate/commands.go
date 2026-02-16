package migrate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// Config holds the configuration for Atlas CLI commands.
type Config struct {
	AtlasBin     string // Path to atlas binary
	MigrationDir string // Path to migrations/ directory
	SchemaURL    string // Path to gen/atlas/schema.hcl (file:// URL)
	DatabaseURL  string // PostgreSQL connection URL
	DevURL       string // Dev database URL for atlas diff
}

// Diff generates a new migration by comparing the current schema with the target schema.
// If force is false and the migration contains destructive changes, returns an error.
func Diff(cfg Config, name string, force bool) (string, error) {
	// Check atlas binary exists
	if _, err := os.Stat(cfg.AtlasBin); err != nil {
		return "", fmt.Errorf("atlas binary not found. Run 'forge tool sync' to download it")
	}

	// Ensure migration directory exists
	if err := os.MkdirAll(cfg.MigrationDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migration directory: %w", err)
	}

	// Build atlas command
	args := []string{
		"migrate", "diff", name,
		"--dir", "file://" + cfg.MigrationDir,
		"--to", "file://" + cfg.SchemaURL,
		"--dev-url", cfg.DevURL,
	}

	// Execute atlas command
	cmd := exec.Command(cfg.AtlasBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Combine stdout and stderr for error context
		output := stdout.String() + stderr.String()
		return "", fmt.Errorf("atlas migrate diff failed: %w\nOutput: %s", err, output)
	}

	// Find the newly created migration file (newest .sql file in migrations/)
	migrationFile, err := findNewestMigration(cfg.MigrationDir)
	if err != nil {
		return "", fmt.Errorf("failed to find generated migration: %w", err)
	}

	// If force flag not set, check for destructive changes
	if !force {
		sqlContent, err := os.ReadFile(migrationFile)
		if err != nil {
			return "", fmt.Errorf("failed to read migration file: %w", err)
		}

		if ContainsDestructiveChange(string(sqlContent)) {
			changes := FindDestructiveChanges(string(sqlContent))
			warning := DestructiveWarning(changes)

			// Delete the generated migration file since it's rejected
			os.Remove(migrationFile)

			return "", fmt.Errorf("%s", warning)
		}
	}

	return migrationFile, nil
}

// Up applies pending migrations to the database.
func Up(cfg Config) (string, error) {
	// Check atlas binary exists
	if _, err := os.Stat(cfg.AtlasBin); err != nil {
		return "", fmt.Errorf("atlas binary not found. Run 'forge tool sync' to download it")
	}

	// Build atlas command
	args := []string{
		"migrate", "apply",
		"--dir", "file://" + cfg.MigrationDir,
		"--url", cfg.DatabaseURL,
	}

	// Execute atlas command
	cmd := exec.Command(cfg.AtlasBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		output := stdout.String() + stderr.String()
		return "", fmt.Errorf("atlas migrate apply failed: %w\nOutput: %s", err, output)
	}

	return stdout.String() + stderr.String(), nil
}

// Down rolls back the last applied migration.
func Down(cfg Config) (string, error) {
	// Check atlas binary exists
	if _, err := os.Stat(cfg.AtlasBin); err != nil {
		return "", fmt.Errorf("atlas binary not found. Run 'forge tool sync' to download it")
	}

	// Build atlas command
	args := []string{
		"migrate", "down",
		"--dir", "file://" + cfg.MigrationDir,
		"--url", cfg.DatabaseURL,
		"--dev-url", cfg.DevURL,
	}

	// Execute atlas command
	cmd := exec.Command(cfg.AtlasBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		output := stdout.String() + stderr.String()
		return "", fmt.Errorf("atlas migrate down failed: %w\nOutput: %s", err, output)
	}

	return stdout.String() + stderr.String(), nil
}

// Status shows the current migration state.
func Status(cfg Config) (string, error) {
	// Check atlas binary exists
	if _, err := os.Stat(cfg.AtlasBin); err != nil {
		return "", fmt.Errorf("atlas binary not found. Run 'forge tool sync' to download it")
	}

	// Build atlas command
	args := []string{
		"migrate", "status",
		"--dir", "file://" + cfg.MigrationDir,
		"--url", cfg.DatabaseURL,
	}

	// Execute atlas command
	cmd := exec.Command(cfg.AtlasBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		output := stdout.String() + stderr.String()
		return "", fmt.Errorf("atlas migrate status failed: %w\nOutput: %s", err, output)
	}

	return stdout.String() + stderr.String(), nil
}

// Hash recomputes the atlas.sum checksum file after manual migration edits.
func Hash(cfg Config) error {
	// Check atlas binary exists
	if _, err := os.Stat(cfg.AtlasBin); err != nil {
		return fmt.Errorf("atlas binary not found. Run 'forge tool sync' to download it")
	}

	// Build atlas command
	args := []string{
		"migrate", "hash",
		"--dir", "file://" + cfg.MigrationDir,
	}

	// Execute atlas command
	cmd := exec.Command(cfg.AtlasBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("atlas migrate hash failed: %w\nOutput: %s", err, stderr.String())
	}

	return nil
}

// findNewestMigration returns the path to the most recently created .sql file in the migrations directory.
func findNewestMigration(migrationDir string) (string, error) {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return "", err
	}

	var sqlFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, entry)
		}
	}

	if len(sqlFiles) == 0 {
		return "", fmt.Errorf("no .sql files found in migrations directory")
	}

	// Sort by modification time (newest first)
	sort.Slice(sqlFiles, func(i, j int) bool {
		infoI, _ := sqlFiles[i].Info()
		infoJ, _ := sqlFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return filepath.Join(migrationDir, sqlFiles[0].Name()), nil
}
