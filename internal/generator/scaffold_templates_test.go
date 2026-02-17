package generator

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/forge-framework/forge/internal/parser"
)

// TestScaffoldTemplates verifies the three scaffold templ templates render
// correctly for a sample resource with various field types and modifiers.
func TestScaffoldTemplates(t *testing.T) {
	resource := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{
				{Type: "Required"},
				{Type: "Sortable"},
				{Type: "Filterable"},
			}},
			{Name: "Price", Type: "Decimal"},
			{Name: "Active", Type: "Bool"},
			{Name: "Quantity", Type: "Int"},
			{Name: "Status", Type: "Enum", EnumValues: []string{"active", "inactive"}},
			{Name: "AdminNotes", Type: "Text", Modifiers: []parser.ModifierIR{
				{Type: "Visibility", Value: "admin"},
			}},
			{Name: "InternalCode", Type: "String", Modifiers: []parser.ModifierIR{
				{Type: "Mutability", Value: "admin"},
			}},
		},
	}

	data := struct {
		Resource      parser.ResourceIR
		ProjectModule string
	}{
		Resource:      resource,
		ProjectModule: "github.com/example/myapp",
	}

	tests := []struct {
		name   string
		tmpl   string
		checks []string
	}{
		{
			name: "form",
			tmpl: "templates/scaffold_form.templ.tmpl",
			checks: []string{
				// Package and imports
				"package views",
				"gen/html/primitives",
				"gen/models",
				// Component signature with role parameter
				"ProductForm",
				"*models.Product",
				"errors map[string]string",
				"role string",
				// Datastar SSE form element
				"product-form",
				"data-on:submit__prevent",
				"@post('/products')",
				// data-signals for non-ID fields
				"data-signals",
				"name: ''",
				"price: ''",
				"active: ''",
				// data-bind on inline inputs (Bool, Int generate inline inputs with data-bind)
				"data-bind",
				// FormField primitive usage
				"primitives.FormField",
				// Visibility modifier generates role guard
				`role == "" || role == "admin"`,
				// fmt.Sprint used for read-only fallback and DecimalInput
				"fmt.Sprint",
				// Submit button with Tailwind styling
				"bg-blue-600",
				"Save",
			},
		},
		{
			name: "detail",
			tmpl: "templates/scaffold_detail.templ.tmpl",
			checks: []string{
				// Package and imports
				"package views",
				"gen/models",
				// Component signature (read-only, no role param)
				"ProductDetail",
				"*models.Product",
				// dl/dt/dd layout for field display
				"<dl",
				"<dt",
				"<dd",
				// Field values rendered with fmt.Sprint
				"fmt.Sprint",
				"product.Name",
				"product.ID",
				// Action buttons
				"Edit",
				"Back to list",
				"/products",
			},
		},
		{
			name: "list",
			tmpl: "templates/scaffold_list.templ.tmpl",
			checks: []string{
				// Package and imports
				"package views",
				"gen/models",
				// Component signature with sort/pagination params
				"ProductList",
				"[]models.Product",
				"currentSort string",
				"currentDir string",
				"page int",
				"totalPages int",
				// Table structure
				"<table",
				"<thead",
				"<tbody",
				// New resource button
				"New Product",
				"/products/new",
				// Sort header with Datastar click handler (Name is Sortable)
				"data-on:click",
				"@get('/products')",
				// Filter input for filterable field (Name is Filterable)
				"filter_name",
				// Pagination controls
				"Previous",
				"Next",
				"totalPages",
				// Action links per row
				"View",
				"Edit",
				// fmt.Sprint for field value rendering
				"fmt.Sprint",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.tmpl)
			if err != nil {
				t.Fatalf("reading template: %v", err)
			}

			tmpl, err := template.New("test").Funcs(BuildFuncMap()).Parse(string(content))
			if err != nil {
				t.Fatalf("parsing template: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("executing template: %v", err)
			}

			output := buf.String()
			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("output missing %q\nFirst 800 chars of output:\n%s",
						check, output[:minLen(800, len(output))])
				}
			}
		})
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
