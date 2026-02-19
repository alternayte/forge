package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/toolsync"
	"github.com/alternayte/forge/internal/ui"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var outputFlag string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build production binary",
		Long:  `Runs the full build pipeline: generate -> templ generate -> tailwind build -> go build with production optimizations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(cmd, args, outputFlag)
		},
	}

	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output binary path (default: ./bin/<project-name>)")

	return cmd
}

func runBuild(cmd *cobra.Command, args []string, outputFlag string) error {
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

	// Determine output path
	outputPath := outputFlag
	if outputPath == "" {
		name := cfg.Project.Name
		if name == "" {
			name = "app"
		}
		outputPath = filepath.Join(projectRoot, "bin", name)
	}

	fmt.Println()
	fmt.Println(ui.Header("Building production binary..."))
	fmt.Println()

	// Resolve tool binary paths
	forgeBin := filepath.Join(projectRoot, ".forge", "bin")

	// Resolve templ binary
	templBin, err := resolveToolBinary(forgeBin, "templ")
	if err != nil {
		return fmt.Errorf("templ not found: %w", err)
	}

	// Resolve tailwind binary
	tailwindBin, err := resolveToolBinary(forgeBin, "tailwindcss")
	if err != nil {
		return fmt.Errorf("tailwindcss not found: %w", err)
	}

	// Ensure bin output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Ensure migrations/ and static/ directories exist (required for embed.go directives)
	for _, dir := range []string{"migrations", "static"} {
		dirPath := filepath.Join(projectRoot, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("creating %s directory: %w", dir, err)
		}
	}

	// Write embed.go if it doesn't exist
	embedPath := filepath.Join(projectRoot, "embed.go")
	if _, err := os.Stat(embedPath); os.IsNotExist(err) {
		if err := writeEmbedGo(projectRoot); err != nil {
			return fmt.Errorf("writing embed.go: %w", err)
		}
		fmt.Println(ui.Success("Generated embed.go"))
	}

	// Step 1: forge generate
	if err := runStep(projectRoot, "Generating code", "forge", []string{"generate"}); err != nil {
		return fmt.Errorf("forge generate failed: %w", err)
	}

	// Step 2: templ generate ./...
	if err := runStep(projectRoot, "Compiling templates", templBin, []string{"generate", "./..."}); err != nil {
		return fmt.Errorf("templ generate failed: %w", err)
	}

	// Step 3: tailwind build with --minify
	tailwindInput := filepath.Join(projectRoot, "static", "css", "input.css")
	tailwindOutput := filepath.Join(projectRoot, "static", "css", "app.css")
	if _, err := os.Stat(tailwindInput); err == nil {
		// Only run tailwind if input file exists
		if err := runStep(projectRoot, "Building CSS", tailwindBin, []string{
			"--input", tailwindInput,
			"--output", tailwindOutput,
			"--minify",
		}); err != nil {
			return fmt.Errorf("tailwind build failed: %w", err)
		}
	} else {
		fmt.Println(ui.Info("Skipping CSS build (static/css/input.css not found)"))
	}

	// Step 4: go build with production flags
	ldflags, err := buildLdflags(projectRoot)
	if err != nil {
		return fmt.Errorf("building ldflags: %w", err)
	}

	if err := runStep(projectRoot, "Compiling binary", "go", []string{
		"build",
		"-trimpath",
		fmt.Sprintf("-ldflags=%s", ldflags),
		"-o", outputPath,
		".",
	}); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Print binary size
	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("reading binary info: %w", err)
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	fmt.Println()
	fmt.Println(ui.Success(fmt.Sprintf("Binary: %s (%.1f MB)", outputPath, sizeMB)))

	if sizeMB > 30 {
		fmt.Fprintf(os.Stderr, "%s Binary size (%.1f MB) exceeds 30MB target\n", ui.WarnIcon, sizeMB)
	}

	fmt.Println()

	return nil
}

// resolveToolBinary finds a tool binary in .forge/bin first, then PATH.
func resolveToolBinary(forgeBin, toolName string) (string, error) {
	if toolsync.IsToolInstalled(forgeBin, toolName) {
		return toolsync.ToolBinPath(forgeBin, toolName), nil
	}

	// Fall back to PATH
	path, err := exec.LookPath(toolName)
	if err != nil {
		return "", fmt.Errorf("%s not found in .forge/bin or PATH. Run 'forge tool sync' first", toolName)
	}
	return path, nil
}

// runStep runs a pipeline step, printing its name and streaming output.
func runStep(projectRoot, name string, binary string, args []string) error {
	fmt.Print(ui.Info(fmt.Sprintf("  %s...", name)))
	fmt.Println()

	cmd := exec.Command(binary, args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildLdflags constructs the -ldflags value with version, commit, and date injected.
func buildLdflags(projectRoot string) (string, error) {
	// Get git version tag
	version := runGitCmd(projectRoot, "describe", "--tags", "--always")
	if version == "" {
		version = "dev"
	}
	version = strings.TrimSpace(version)

	// Get short commit hash
	commit := runGitCmd(projectRoot, "rev-parse", "--short", "HEAD")
	if commit == "" {
		commit = "none"
	}
	commit = strings.TrimSpace(commit)

	// Get build date in RFC3339 format
	date := time.Now().UTC().Format(time.RFC3339)

	const pkg = "github.com/alternayte/forge/internal/cli"

	ldflags := fmt.Sprintf(
		"-s -w -X %s.Version=%s -X %s.Commit=%s -X %s.Date=%s",
		pkg, version,
		pkg, commit,
		pkg, date,
	)

	return ldflags, nil
}

// runGitCmd runs a git command and returns trimmed stdout, or empty string on error.
func runGitCmd(projectRoot string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// writeEmbedGo writes an embed.go file to the project root with //go:embed directives.
func writeEmbedGo(projectRoot string) error {
	content := `package main

import "embed"

//go:embed migrations
var MigrationsFS embed.FS

//go:embed static
var StaticFS embed.FS
`
	return os.WriteFile(filepath.Join(projectRoot, "embed.go"), []byte(content), 0644)
}
