package watcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/generator"
	"github.com/alternayte/forge/internal/migrate"
	"github.com/alternayte/forge/internal/parser"
	"github.com/alternayte/forge/internal/toolsync"
	"github.com/alternayte/forge/internal/ui"
)

// DevServer orchestrates the development workflow: watch + parse + generate.
type DevServer struct {
	ProjectRoot string
	Config      *config.Config
	watcher     *Watcher
}

// NewDevServer creates a new development server.
func NewDevServer(projectRoot string, cfg *config.Config) *DevServer {
	return &DevServer{
		ProjectRoot: projectRoot,
		Config:      cfg,
	}
}

// Start starts the development server and blocks until context is cancelled.
func (d *DevServer) Start(ctx context.Context) error {
	fmt.Println()
	fmt.Println(ui.Header("forge dev"))
	fmt.Println()

	// Run initial generation (with auto-migrate)
	fmt.Println(ui.Info("Running initial generation..."))
	if err := d.runGenerationAndMigrate(); err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Initial generation failed: %v", err)))
		// Continue watching even if initial generation fails
	} else {
		fmt.Println(ui.Success("Initial generation complete"))
	}
	fmt.Println()

	// Create watcher with onChange callback
	var err error
	d.watcher, err = New(d.onFileChange, 300*time.Millisecond)
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer d.watcher.Close()

	// Add directories to watch
	resourcesDir := filepath.Join(d.ProjectRoot, "resources")
	internalDir := filepath.Join(d.ProjectRoot, "internal")

	if err := d.addRecursive(resourcesDir); err != nil {
		return fmt.Errorf("watch resources: %w", err)
	}

	// Internal dir may not exist yet
	if _, err := os.Stat(internalDir); err == nil {
		if err := d.addRecursive(internalDir); err != nil {
			return fmt.Errorf("watch internal: %w", err)
		}
	}

	// Print watching message
	fmt.Println("Watching for changes...")
	fmt.Println("  " + ui.FilePathStyle.Render("resources/") + " " + ui.DimStyle.Render("— schema files (.go)"))
	if _, err := os.Stat(internalDir); err == nil {
		fmt.Println("  " + ui.FilePathStyle.Render("internal/") + "  " + ui.DimStyle.Render("— application code (.go, .templ)"))
	}
	fmt.Println()
	fmt.Println(ui.DimStyle.Render("Press Ctrl+C to stop"))
	fmt.Println()

	// Block until context is cancelled
	<-ctx.Done()

	fmt.Println()
	fmt.Println(ui.Info("Shutting down..."))
	return nil
}

// onFileChange is called when relevant files change.
func (d *DevServer) onFileChange() {
	timestamp := time.Now().Format("15:04:05")
	fmt.Println()
	fmt.Println(ui.DimStyle.Render(fmt.Sprintf("[%s]", timestamp)) + " " + ui.Info("Change detected, regenerating..."))

	if err := d.runGenerationAndMigrate(); err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Generation failed: %v", err)))
	} else {
		fmt.Println(ui.Success("Regeneration complete"))
		fmt.Println(ui.DimStyle.Render("Run 'go run .' to start your application"))
	}
	fmt.Println()
}

// runGenerationAndMigrate runs code generation and auto-migrates if the schema changed.
func (d *DevServer) runGenerationAndMigrate() error {
	// 1. Hash schema before generation
	beforeHash := d.hashSchemaFile()

	// 2. Run generation
	if err := d.runGeneration(); err != nil {
		return err
	}

	// 3. Hash schema after generation
	afterHash := d.hashSchemaFile()
	if beforeHash == afterHash && beforeHash != "" {
		return nil // No schema changes
	}

	// 4. Auto-create DB if needed (non-fatal)
	if err := d.ensureDatabase(); err != nil {
		fmt.Println(ui.Warn(fmt.Sprintf("Could not auto-create database: %v", err)))
		// Non-fatal: continue to migration attempt
	}

	// 5. Auto-migrate if atlas is available
	atlasBin := toolsync.ToolBinPath(filepath.Join(d.ProjectRoot, ".forge", "bin"), "atlas")
	if !toolsync.IsToolInstalled(filepath.Join(d.ProjectRoot, ".forge", "bin"), "atlas") {
		// Atlas not installed — skip auto-migrate silently
		return nil
	}

	schemaPath := filepath.Join(d.ProjectRoot, "gen", "atlas", "schema.hcl")
	if _, err := os.Stat(schemaPath); err != nil {
		// Schema file not generated yet — skip
		return nil
	}

	fmt.Println(ui.Info("Schema changed, running migration..."))

	migrateCfg := migrate.Config{
		AtlasBin:     atlasBin,
		MigrationDir: filepath.Join(d.ProjectRoot, "migrations"),
		SchemaURL:    schemaPath,
		DatabaseURL:  d.Config.Database.URL,
		DevURL:       d.Config.Database.URL,
	}

	// Generate migration diff (force=true skips destructive warnings in dev)
	_, err := migrate.Diff(migrateCfg, "auto_dev", true)
	if err != nil {
		fmt.Println(ui.Warn(fmt.Sprintf("Migration diff failed: %v", err)))
		return nil // Non-fatal in dev
	}

	// Apply pending migrations
	output, err := migrate.Up(migrateCfg)
	if err != nil {
		fmt.Println(ui.Warn(fmt.Sprintf("Migration apply failed: %v", err)))
		return nil // Non-fatal in dev
	}

	if strings.TrimSpace(output) != "" {
		fmt.Println(output)
	}
	fmt.Println(ui.Success("Migrations applied"))

	return nil
}

// runGeneration parses resources and generates code.
func (d *DevServer) runGeneration() error {
	resourcesDir := filepath.Join(d.ProjectRoot, "resources")

	// Parse resources
	result, err := parser.ParseDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Check for parse errors
	if len(result.Errors) > 0 {
		// Print errors but don't crash the watcher
		for _, parseErr := range result.Errors {
			fmt.Println(ui.Error(parseErr.Error()))
		}
		return fmt.Errorf("schema errors found")
	}

	// Check if any resources were found
	if len(result.Resources) == 0 {
		return fmt.Errorf("no resources found in resources/")
	}

	// Generate code
	genCfg := generator.GenerateConfig{
		OutputDir:     filepath.Join(d.ProjectRoot, "gen"),
		ProjectModule: d.Config.Project.Module,
		ProjectRoot:   d.ProjectRoot,
	}

	if err := generator.Generate(result.Resources, genCfg); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	return nil
}

// hashSchemaFile returns a SHA-256 hex hash of gen/atlas/schema.hcl.
// Returns empty string if the file doesn't exist or can't be read.
func (d *DevServer) hashSchemaFile() string {
	path := filepath.Join(d.ProjectRoot, "gen", "atlas", "schema.hcl")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ensureDatabase creates the application database if it doesn't exist.
// This is a dev-mode convenience — production apps use explicit database management.
func (d *DevServer) ensureDatabase() error {
	dbURL := d.Config.Database.URL
	if dbURL == "" {
		return nil // No database configured
	}

	// Parse the URL to extract the database name
	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("parse database URL: %w", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return nil
	}

	// Connect to the maintenance database (postgres) to create the app database
	maintURL := *u
	maintURL.Path = "/postgres"

	conn, err := pgx.Connect(context.Background(), maintURL.String())
	if err != nil {
		return fmt.Errorf("connect to maintenance DB: %w", err)
	}
	defer conn.Close(context.Background())

	// Check if database exists
	var exists bool
	err = conn.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		// CREATE DATABASE doesn't support parameters — use quoted identifier
		// (dbName comes from the user's own forge.toml, not external input)
		_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{dbName}.Sanitize()))
		if err != nil {
			return err
		}
		fmt.Println(ui.Success(fmt.Sprintf("Created database %s", dbName)))
	}

	return nil
}

// addRecursive recursively adds directories to the watcher.
func (d *DevServer) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			return nil
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") && path != root {
			return filepath.SkipDir
		}

		// Skip gen/ directory
		if entry.Name() == "gen" {
			return filepath.SkipDir
		}

		// Skip node_modules
		if entry.Name() == "node_modules" {
			return filepath.SkipDir
		}

		// Add directory to watcher
		return d.watcher.Add(path)
	})
}
