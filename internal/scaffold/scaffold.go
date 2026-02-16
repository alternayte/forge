package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

// ProjectData holds template variables for project scaffolding
type ProjectData struct {
	Name                 string
	Module               string
	GoVersion            string
	ExampleResource      string // lowercase (e.g., "product")
	ExampleResourceTitle string // titlecase (e.g., "Product")
}

// CreateProject scaffolds a new Forge project at the given path
func CreateProject(projectPath string, data ProjectData) error {
	// Create directory structure
	dirs := []string{
		filepath.Join(projectPath, "resources", data.ExampleResource),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		// Fix permissions (umask may restrict)
		if err := os.Chmod(dir, 0755); err != nil {
			return fmt.Errorf("failed to chmod directory %s: %w", dir, err)
		}
	}

	// Render templates
	templates := []struct {
		src  string
		dest string
	}{
		{"templates/forge.toml.tmpl", "forge.toml"},
		{"templates/main.go.tmpl", "main.go"},
		{"templates/go.mod.tmpl", "go.mod"},
		{"templates/gitignore.tmpl", ".gitignore"},
		{"templates/README.md.tmpl", "README.md"},
		{"templates/schema.go.tmpl", filepath.Join("resources", data.ExampleResource, "schema.go")},
	}

	for _, t := range templates {
		destPath := filepath.Join(projectPath, t.dest)
		if err := renderTemplate(t.src, destPath, data); err != nil {
			return fmt.Errorf("failed to render template %s: %w", t.src, err)
		}
	}

	return nil
}

// renderTemplate reads an embedded template, executes it, and writes to dest
func renderTemplate(src, dest string, data ProjectData) error {
	// Read template from embedded FS
	content, err := TemplatesFS.ReadFile(src)
	if err != nil {
		return err
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(src)).Parse(string(content))
	if err != nil {
		return err
	}

	// Create destination file
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// InferModule generates a module path from the project name
func InferModule(projectName string) string {
	// Try to detect git user config for generating github.com/user/project
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		username := strings.TrimSpace(string(output))
		// Convert to lowercase and replace spaces with hyphens for GitHub username
		username = strings.ToLower(strings.ReplaceAll(username, " ", "-"))
		return fmt.Sprintf("github.com/%s/%s", username, projectName)
	}

	// Fallback to just project name
	return projectName
}

// GetGoVersion returns the current Go version (e.g., "1.23")
func GetGoVersion() string {
	version := runtime.Version() // e.g., "go1.23.1"
	version = strings.TrimPrefix(version, "go")
	// Extract major.minor only
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}
