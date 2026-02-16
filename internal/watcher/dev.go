package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-framework/forge/internal/config"
	"github.com/forge-framework/forge/internal/generator"
	"github.com/forge-framework/forge/internal/parser"
	"github.com/forge-framework/forge/internal/ui"
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

	// Run initial generation
	fmt.Println(ui.Info("Running initial generation..."))
	if err := d.runGeneration(); err != nil {
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

	if err := d.runGeneration(); err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Generation failed: %v", err)))
	} else {
		fmt.Println(ui.Success("Regeneration complete"))
		fmt.Println(ui.DimStyle.Render("Run 'go run .' to start your application"))
	}
	fmt.Println()
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
