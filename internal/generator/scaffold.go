package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alternayte/forge/internal/parser"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ScaffoldResult holds the outcome of a ScaffoldResource call.
type ScaffoldResult struct {
	Created []string // Files written to disk
	Skipped []string // Files that already existed (not overwritten)
}

// scaffoldFile describes a single file to be scaffolded.
type scaffoldFile struct {
	Template   string // Template name in TemplatesFS
	OutputPath string // Relative to resource dir (e.g., "views/form.templ")
	IsTempl    bool   // true for .templ files, false for .go files
}

// scaffoldFiles returns the list of scaffold files for a resource.
// It conditionally includes jobs.go when the resource declares Hooks.
func scaffoldFiles(resource parser.ResourceIR) []scaffoldFile {
	files := []scaffoldFile{
		{Template: "templates/scaffold_form.templ.tmpl", OutputPath: "views/form.templ", IsTempl: true},
		{Template: "templates/scaffold_list.templ.tmpl", OutputPath: "views/list.templ", IsTempl: true},
		{Template: "templates/scaffold_detail.templ.tmpl", OutputPath: "views/detail.templ", IsTempl: true},
		{Template: "templates/scaffold_error.templ.tmpl", OutputPath: "views/error.templ", IsTempl: true},
		{Template: "templates/scaffold_handlers.go.tmpl", OutputPath: "handlers.go", IsTempl: false},
		{Template: "templates/scaffold_hooks.go.tmpl", OutputPath: "hooks.go", IsTempl: false},
	}

	// Conditionally include jobs.go for resources with lifecycle hooks (JOBS-02).
	if len(resource.Options.Hooks.AfterCreate) > 0 || len(resource.Options.Hooks.AfterUpdate) > 0 {
		files = append(files, scaffoldFile{
			Template:   "templates/scaffold_jobs.go.tmpl",
			OutputPath: "jobs.go",
			IsTempl:    false,
		})
	}

	return files
}

// templateData is the data passed to scaffold templates.
type scaffoldTemplateData struct {
	Resource      parser.ResourceIR
	ProjectModule string
}

// ScaffoldResource writes scaffold-once files for a resource into resources/<name>/.
// It skips any file that already exists on disk (protecting developer customizations).
func ScaffoldResource(resource parser.ResourceIR, projectRoot, projectModule string) (*ScaffoldResult, error) {
	resourceDir := filepath.Join(projectRoot, "resources", snake(resource.Name))
	result := &ScaffoldResult{}

	rendered, err := renderScaffoldToMap(resource, projectModule)
	if err != nil {
		return nil, fmt.Errorf("rendering scaffold templates: %w", err)
	}

	for _, sf := range scaffoldFiles(resource) {
		fullPath := filepath.Join(resourceDir, sf.OutputPath)

		// Skip if file already exists
		if _, statErr := os.Stat(fullPath); statErr == nil {
			result.Skipped = append(result.Skipped, sf.OutputPath)
			continue
		}

		raw := rendered[sf.OutputPath]

		var writeErr error
		if sf.IsTempl {
			writeErr = writeRawFile(fullPath, raw)
		} else {
			writeErr = writeGoFile(fullPath, raw)
		}
		if writeErr != nil {
			return nil, fmt.Errorf("writing %s: %w", sf.OutputPath, writeErr)
		}

		result.Created = append(result.Created, sf.OutputPath)
	}

	return result, nil
}

// renderScaffoldToMap renders all scaffold templates into a map of outputPath -> rendered bytes.
// Used by both ScaffoldResource and DiffResource.
func renderScaffoldToMap(resource parser.ResourceIR, projectModule string) (map[string][]byte, error) {
	data := scaffoldTemplateData{
		Resource:      resource,
		ProjectModule: projectModule,
	}

	files := scaffoldFiles(resource)
	out := make(map[string][]byte, len(files))
	for _, sf := range files {
		rendered, err := renderTemplate(sf.Template, data)
		if err != nil {
			return nil, fmt.Errorf("rendering template %s: %w", sf.Template, err)
		}
		out[sf.OutputPath] = rendered
	}
	return out, nil
}

// DiffResource produces a unified diff between on-disk scaffold files and freshly-rendered
// scaffold output. Files that don't exist on disk are reported as "would be created".
func DiffResource(resource parser.ResourceIR, projectRoot, projectModule string) (string, error) {
	resourceDir := filepath.Join(projectRoot, "resources", snake(resource.Name))

	rendered, err := renderScaffoldToMap(resource, projectModule)
	if err != nil {
		return "", fmt.Errorf("rendering scaffold templates: %w", err)
	}

	dmp := diffmatchpatch.New()
	var sb strings.Builder

	for _, sf := range scaffoldFiles(resource) {
		fullPath := filepath.Join(resourceDir, sf.OutputPath)

		sb.WriteString(fmt.Sprintf("=== %s ===\n", sf.OutputPath))

		onDisk, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				sb.WriteString("File does not exist (would be created)\n\n")
				continue
			}
			return "", fmt.Errorf("reading %s: %w", fullPath, readErr)
		}

		fresh := rendered[sf.OutputPath]

		// Produce diff using diffmatchpatch
		diffs := dmp.DiffMain(string(onDisk), string(fresh), false)
		dmp.DiffCleanupSemantic(diffs)

		// Check if there are any actual differences
		hasChanges := false
		for _, d := range diffs {
			if d.Type != diffmatchpatch.DiffEqual {
				hasChanges = true
				break
			}
		}

		if !hasChanges {
			sb.WriteString("No changes.\n\n")
			continue
		}

		sb.WriteString(dmp.DiffPrettyText(diffs))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
