package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

// sampleProduct returns a ResourceIR representing a Product with Name, Price, and Quantity fields.
func sampleProduct() parser.ResourceIR {
	return parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{
				{Type: "Required"},
			}},
			{Name: "Price", Type: "Decimal"},
			{Name: "Quantity", Type: "Int"},
		},
	}
}

// sampleCategory returns a ResourceIR for a Category resource.
func sampleCategory() parser.ResourceIR {
	return parser.ResourceIR{
		Name: "Category",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Title", Type: "String"},
		},
	}
}

// TestScaffoldResource verifies that ScaffoldResource writes all 5 scaffold files
// with correct content into a fresh temp directory.
func TestScaffoldResource(t *testing.T) {
	dir := t.TempDir()
	resource := sampleProduct()
	const module = "github.com/example/myapp"

	result, err := ScaffoldResource(resource, dir, module)
	if err != nil {
		t.Fatalf("ScaffoldResource failed: %v", err)
	}

	// All 5 files should be created
	if len(result.Created) != 5 {
		t.Errorf("expected 5 created files, got %d: %v", len(result.Created), result.Created)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped files, got %d: %v", len(result.Skipped), result.Skipped)
	}

	resourceDir := filepath.Join(dir, "resources", "product")

	// Verify form.templ content
	formContent := readFile(t, filepath.Join(resourceDir, "views/form.templ"))
	for _, check := range []string{"ProductForm", "data-bind", "errors"} {
		if !strings.Contains(formContent, check) {
			t.Errorf("form.templ missing %q", check)
		}
	}

	// Verify list.templ content
	listContent := readFile(t, filepath.Join(resourceDir, "views/list.templ"))
	for _, check := range []string{"ProductList", "table"} {
		if !strings.Contains(listContent, check) {
			t.Errorf("list.templ missing %q", check)
		}
	}

	// Verify detail.templ content
	detailContent := readFile(t, filepath.Join(resourceDir, "views/detail.templ"))
	if !strings.Contains(detailContent, "ProductDetail") {
		t.Errorf("detail.templ missing %q", "ProductDetail")
	}

	// Verify handlers.go content
	handlersContent := readFile(t, filepath.Join(resourceDir, "handlers.go"))
	for _, check := range []string{"HandleCreate", "acts.Create", "datastar"} {
		if !strings.Contains(handlersContent, check) {
			t.Errorf("handlers.go missing %q", check)
		}
	}

	// Verify hooks.go content
	hooksContent := readFile(t, filepath.Join(resourceDir, "hooks.go"))
	for _, check := range []string{"BeforeCreate", "AfterCreate"} {
		if !strings.Contains(hooksContent, check) {
			t.Errorf("hooks.go missing %q", check)
		}
	}
}

// TestScaffoldResource_SkipsExisting verifies that ScaffoldResource does not overwrite
// files that already exist on disk.
func TestScaffoldResource_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	resource := sampleProduct()
	const module = "github.com/example/myapp"

	// Pre-create views/form.templ with custom content
	viewsDir := filepath.Join(dir, "resources", "product", "views")
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("creating views dir: %v", err)
	}
	customContent := "// my custom form implementation"
	if err := os.WriteFile(filepath.Join(viewsDir, "form.templ"), []byte(customContent), 0644); err != nil {
		t.Fatalf("pre-creating form.templ: %v", err)
	}

	result, err := ScaffoldResource(resource, dir, module)
	if err != nil {
		t.Fatalf("ScaffoldResource failed: %v", err)
	}

	// views/form.templ should be in Skipped
	if !containsPath(result.Skipped, "views/form.templ") {
		t.Errorf("expected views/form.templ in Skipped, got: %v", result.Skipped)
	}

	// Other 4 files should be in Created
	if len(result.Created) != 4 {
		t.Errorf("expected 4 created files, got %d: %v", len(result.Created), result.Created)
	}

	// Verify form.templ was NOT overwritten
	actualContent := readFile(t, filepath.Join(viewsDir, "form.templ"))
	if actualContent != customContent {
		t.Errorf("form.templ was overwritten; expected custom content, got: %s", actualContent)
	}
}

// TestScaffoldResource_MultipleResources verifies that scaffolding multiple resources
// creates separate directories with no cross-contamination.
func TestScaffoldResource_MultipleResources(t *testing.T) {
	dir := t.TempDir()
	const module = "github.com/example/myapp"

	product := sampleProduct()
	category := sampleCategory()

	_, err := ScaffoldResource(product, dir, module)
	if err != nil {
		t.Fatalf("ScaffoldResource(product) failed: %v", err)
	}
	_, err = ScaffoldResource(category, dir, module)
	if err != nil {
		t.Fatalf("ScaffoldResource(category) failed: %v", err)
	}

	// Verify separate directories exist
	productDir := filepath.Join(dir, "resources", "product")
	categoryDir := filepath.Join(dir, "resources", "category")

	for _, d := range []string{productDir, categoryDir} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("expected directory to exist: %s", d)
		}
	}

	// No cross-contamination: product files should not reference Category
	productHandlers := readFile(t, filepath.Join(productDir, "handlers.go"))
	if strings.Contains(productHandlers, "Category") {
		t.Errorf("product handlers.go should not reference Category, but does")
	}

	// No cross-contamination: category files should not reference Product
	categoryHandlers := readFile(t, filepath.Join(categoryDir, "handlers.go"))
	if strings.Contains(categoryHandlers, "Product") {
		t.Errorf("category handlers.go should not reference Product, but does")
	}
}

// TestDiffResource verifies that DiffResource produces a non-empty diff when a scaffolded
// file has been modified on disk.
func TestDiffResource(t *testing.T) {
	dir := t.TempDir()
	resource := sampleProduct()
	const module = "github.com/example/myapp"

	// First scaffold the resource
	_, err := ScaffoldResource(resource, dir, module)
	if err != nil {
		t.Fatalf("ScaffoldResource failed: %v", err)
	}

	// Modify form.templ to introduce a difference
	formPath := filepath.Join(dir, "resources", "product", "views", "form.templ")
	original := readFile(t, formPath)
	modified := original + "\n// developer customization\n"
	if err := os.WriteFile(formPath, []byte(modified), 0644); err != nil {
		t.Fatalf("modifying form.templ: %v", err)
	}

	diff, err := DiffResource(resource, dir, module)
	if err != nil {
		t.Fatalf("DiffResource failed: %v", err)
	}

	if diff == "" {
		t.Error("expected non-empty diff output, got empty string")
	}

	// The diff should reference the modified text
	if !strings.Contains(diff, "developer customization") {
		t.Errorf("diff output does not contain changed text %q\nDiff:\n%s", "developer customization", diff)
	}
}

// TestDiffResource_NoExistingFiles verifies that DiffResource correctly reports files
// that would be created when no scaffold has been run yet.
func TestDiffResource_NoExistingFiles(t *testing.T) {
	dir := t.TempDir()
	resource := sampleProduct()
	const module = "github.com/example/myapp"

	diff, err := DiffResource(resource, dir, module)
	if err != nil {
		t.Fatalf("DiffResource failed: %v", err)
	}

	if diff == "" {
		t.Error("expected non-empty diff output for non-existent files, got empty string")
	}

	// Should mention that files would be created
	if !strings.Contains(diff, "would be created") {
		t.Errorf("expected diff to mention 'would be created', got:\n%s", diff)
	}
}

// readFile is a test helper that reads a file and fails the test on error.
func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file %s: %v", path, err)
	}
	return string(content)
}

// containsPath checks whether a slice of paths contains the given path.
func containsPath(paths []string, target string) bool {
	for _, p := range paths {
		if p == target {
			return true
		}
	}
	return false
}
