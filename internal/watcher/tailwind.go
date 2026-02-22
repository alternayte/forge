package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// tailwindInputCSS is the Tailwind v4 input CSS content.
// Tailwind v4 uses a single @import directive and auto-detects content sources.
const tailwindInputCSS = `@import "tailwindcss";
`

// tailwindBinPath returns the expected path to the Tailwind CSS standalone CLI binary.
// The binary is placed in .forge/bin/tailwindcss (downloaded by 'forge tool sync').
func tailwindBinPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".forge", "bin", "tailwindcss")
}

// RunTailwind compiles the project's Tailwind CSS once (single-shot, synchronous).
// It reads from resources/css/input.css and writes to public/css/output.css using
// the standalone Tailwind CLI binary at .forge/bin/tailwindcss (zero npm dependency).
//
// Returns an error if the binary is not installed or compilation fails.
// Run 'forge tool sync' to download the Tailwind CLI binary.
func RunTailwind(projectRoot string) error {
	binPath := tailwindBinPath(projectRoot)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("Tailwind CSS binary not found at %s. Run 'forge tool sync' first", binPath)
	}

	inputPath := filepath.Join(projectRoot, "resources", "css", "input.css")
	outputPath := filepath.Join(projectRoot, "public", "css", "output.css")

	// Ensure the output directory exists before asking Tailwind to write into it.
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating public/css directory: %w", err)
	}

	cmd := exec.Command(binPath, "-i", inputPath, "-o", outputPath)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunTailwindWatch starts the Tailwind CLI in --watch mode for continuous compilation
// during development. The returned *exec.Cmd can be used by the caller to stop the
// process (cmd.Process.Kill() or cmd.Wait() after cancellation).
//
// Used by 'forge dev' to recompile CSS whenever .templ files change.
func RunTailwindWatch(projectRoot string) (*exec.Cmd, error) {
	binPath := tailwindBinPath(projectRoot)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Tailwind CSS binary not found at %s. Run 'forge tool sync' first", binPath)
	}

	inputPath := filepath.Join(projectRoot, "resources", "css", "input.css")
	outputPath := filepath.Join(projectRoot, "public", "css", "output.css")

	// Ensure the output directory exists before starting the watcher.
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("creating public/css directory: %w", err)
	}

	cmd := exec.Command(binPath, "-i", inputPath, "-o", outputPath, "--watch")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting Tailwind watch: %w", err)
	}

	return cmd, nil
}

// ScaffoldTailwindInput creates the initial Tailwind CSS input file for a new project.
// Uses the scaffold-once pattern: existing files are never overwritten to preserve
// developer customizations.
//
// Creates:
//   - resources/css/input.css  — Tailwind v4 @import directive
//
// Tailwind v4 auto-detects content sources from all non-gitignored files,
// so no tailwind.config.js is needed.
func ScaffoldTailwindInput(projectRoot string) error {
	inputPath := filepath.Join(projectRoot, "resources", "css", "input.css")

	// Scaffold input.css (skip if already exists)
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(inputPath), 0755); err != nil {
			return fmt.Errorf("creating resources/css directory: %w", err)
		}
		if err := os.WriteFile(inputPath, []byte(tailwindInputCSS), 0644); err != nil {
			return fmt.Errorf("writing resources/css/input.css: %w", err)
		}
	}

	return nil
}
