package generator

import (
	"bytes"
	"os"
	"path/filepath"
)

// GenerateMain scaffolds a main.go in the project root if one doesn't already exist.
// This follows the scaffold-once pattern: if the user has already edited main.go to
// include forge framework wiring, we don't overwrite it.
// Use `forge generate --diff` to show what a fresh main.go would look like.
func GenerateMain(projectRoot, projectModule string) error {
	mainPath := filepath.Join(projectRoot, "main.go")

	// Scaffold-once: skip if main.go already has forge framework wiring
	if _, err := os.Stat(mainPath); err == nil {
		// File exists — check if it already uses forge.App
		content, err := os.ReadFile(mainPath)
		if err != nil {
			return err
		}
		// If main.go already imports forge, it's been generated before — skip
		if bytes.Contains(content, []byte("forge.New")) || bytes.Contains(content, []byte("forge.LoadConfig")) {
			return nil
		}
	}

	// Render the main.go template
	data := struct {
		ProjectModule string
	}{
		ProjectModule: projectModule,
	}

	raw, err := renderTemplate("templates/main.go.tmpl", data)
	if err != nil {
		return err
	}

	return writeGoFile(mainPath, raw)
}
