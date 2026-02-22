package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

// TestGoldenPath generates a complete project for a Post resource and verifies
// that it compiles with `go build ./...`. This catches template bugs that prevent
// any generated project from building.
//
// Requires network access for `go mod tidy` on first run. Skip with -short.
func TestGoldenPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping golden path test in short mode (requires network for go mod tidy)")
	}

	// Locate the forge-go module root (two dirs up from internal/generator/)
	_, thisFile, _, _ := runtime.Caller(0)
	forgeRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

	// Create a temp directory as the project root
	projectRoot := t.TempDir()
	genDir := filepath.Join(projectRoot, "gen")

	// Define a Post resource: Title (String, Required+Filterable+Sortable), Body (Text)
	post := parser.ResourceIR{
		Name:          "Post",
		HasTimestamps: true,
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{
				Name: "Title",
				Type: "String",
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
					{Type: "Filterable"},
					{Type: "Sortable"},
				},
			},
			{
				Name: "Body",
				Type: "Text",
			},
		},
	}

	resources := []parser.ResourceIR{post}
	projectModule := "example.com/fgtester"

	// Step 1: Generate all code
	err := Generate(resources, GenerateConfig{
		OutputDir:     genDir,
		ProjectModule: projectModule,
		ProjectRoot:   projectRoot,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Step 2: Scaffold resource handlers and hooks
	_, err = ScaffoldResource(post, projectRoot, projectModule)
	if err != nil {
		t.Fatalf("ScaffoldResource failed: %v", err)
	}

	// Step 3: Create stub _templ.go files for views (templ binary not available in tests).
	// These stubs export the functions that scaffold_handlers.go.tmpl calls.
	viewsDir := filepath.Join(projectRoot, "resources", "post", "views")
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("mkdir views: %v", err)
	}

	viewsStub := `package views

import (
	"context"
	"io"

	"example.com/fgtester/gen/models"
)

// stubComponent satisfies the datastar.TemplComponent interface.
type stubComponent struct{}

func (stubComponent) Render(_ context.Context, _ io.Writer) error { return nil }

func PostList(_ []models.Post, _ string, _ string, _ int, _ int) stubComponent {
	return stubComponent{}
}

func PostForm(_ *models.Post, _ map[string]string, _ string) stubComponent {
	return stubComponent{}
}

func PostDetail(_ *models.Post) stubComponent {
	return stubComponent{}
}

func PostError(_ string) stubComponent {
	return stubComponent{}
}
`
	if err := os.WriteFile(filepath.Join(viewsDir, "views_templ.go"), []byte(viewsStub), 0644); err != nil {
		t.Fatalf("write views stub: %v", err)
	}

	// Create layout stub (templ binary not available in tests)
	layoutDir := filepath.Join(genDir, "html", "layout")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		t.Fatalf("mkdir layout: %v", err)
	}

	layoutStub := `package layout

import (
	"context"
	"io"
)

type TemplComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

type stubComponent struct{}

func (stubComponent) Render(_ context.Context, _ io.Writer) error { return nil }

func Page(_ string, _ TemplComponent) stubComponent {
	return stubComponent{}
}
`
	if err := os.WriteFile(filepath.Join(layoutDir, "layout_templ.go"), []byte(layoutStub), 0644); err != nil {
		t.Fatalf("write layout stub: %v", err)
	}

	// Step 4: Write go.mod with replace directive pointing to local forge-go module
	goMod := fmt.Sprintf(`module example.com/fgtester

go 1.25.6

require github.com/alternayte/forge v0.0.0

replace github.com/alternayte/forge => %s
`, forgeRoot)

	if err := os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// Step 5: Write forge.toml (needed by forge.LoadConfig in main.go)
	forgeToml := `[project]
name = "fgtester"
module = "example.com/fgtester"

[server]
port = 8080
`
	if err := os.WriteFile(filepath.Join(projectRoot, "forge.toml"), []byte(forgeToml), 0644); err != nil {
		t.Fatalf("write forge.toml: %v", err)
	}

	// Step 6: go mod tidy (fetches deps) then go build ./...
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = projectRoot
	tidy.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	tidyOut, err := tidy.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed:\n%s\n%v", tidyOut, err)
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = projectRoot
	buildOut, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed:\n%s\n%v", buildOut, err)
	}
}
