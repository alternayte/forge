package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGenerateQueries(t *testing.T) {
	// Create test resource with filterable and sortable fields
	resource := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{
				Name: "ID",
				Type: "UUID",
			},
			{
				Name: "Name",
				Type: "String",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
					{Type: "Sortable"},
				},
			},
			{
				Name: "Description",
				Type: "Text",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
				},
			},
			{
				Name: "Price",
				Type: "Int",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
					{Type: "Sortable"},
				},
			},
			{
				Name: "Stock",
				Type: "BigInt",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
				},
			},
			{
				Name: "Email",
				Type: "Email",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
				},
			},
			{
				Name: "CreatedAt",
				Type: "DateTime",
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
					{Type: "Sortable"},
				},
			},
		},
	}

	// Create temporary output directory
	tempDir := t.TempDir()

	// Run generator
	err := GenerateQueries([]parser.ResourceIR{resource}, filepath.Join(tempDir, "gen"), "example.com/testapp")
	if err != nil {
		t.Fatalf("GenerateQueries failed: %v", err)
	}

	// Verify product_queries.go exists
	queriesPath := filepath.Join(tempDir, "gen", "queries", "product_queries.go")
	if _, err := os.Stat(queriesPath); os.IsNotExist(err) {
		t.Fatalf("Expected product_queries.go to exist at %s", queriesPath)
	}

	// Read generated file
	content, err := os.ReadFile(queriesPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify file contains expected functions
	if !strings.Contains(contentStr, "func FilterMods") {
		t.Error("Generated file should contain FilterMods function")
	}

	if !strings.Contains(contentStr, "func SortMod") {
		t.Error("Generated file should contain SortMod function")
	}

	// Verify Filterable fields get EQ/NEQ methods
	if !strings.Contains(contentStr, "func (f ProductFilters) NameEQ") {
		t.Error("Generated file should contain NameEQ method")
	}

	if !strings.Contains(contentStr, "func (f ProductFilters) NameNEQ") {
		t.Error("Generated file should contain NameNEQ method")
	}

	// Verify numeric Filterable fields get GTE/LTE methods
	if !strings.Contains(contentStr, "func (f ProductFilters) PriceGTE") {
		t.Error("Generated file should contain PriceGTE method for numeric field")
	}

	if !strings.Contains(contentStr, "func (f ProductFilters) PriceLTE") {
		t.Error("Generated file should contain PriceLTE method for numeric field")
	}

	// Verify DateTime fields get GTE/LTE methods
	if !strings.Contains(contentStr, "func (f ProductFilters) CreatedAtGTE") {
		t.Error("Generated file should contain CreatedAtGTE method for DateTime field")
	}

	if !strings.Contains(contentStr, "func (f ProductFilters) CreatedAtLTE") {
		t.Error("Generated file should contain CreatedAtLTE method for DateTime field")
	}

	// Verify string Filterable fields get Contains method
	if !strings.Contains(contentStr, "func (f ProductFilters) NameContains") {
		t.Error("Generated file should contain NameContains method for string field")
	}

	if !strings.Contains(contentStr, "func (f ProductFilters) DescriptionContains") {
		t.Error("Generated file should contain DescriptionContains method for text field")
	}

	if !strings.Contains(contentStr, "func (f ProductFilters) EmailContains") {
		t.Error("Generated file should contain EmailContains method for email field")
	}

	// Verify BigInt fields get GTE/LTE methods
	if !strings.Contains(contentStr, "func (f ProductFilters) StockGTE") {
		t.Error("Generated file should contain StockGTE method for BigInt field")
	}

	// Verify Bob imports
	if !strings.Contains(contentStr, `"github.com/stephenafamo/bob/dialect/psql"`) {
		t.Error("Generated file should import Bob psql package")
	}

	if !strings.Contains(contentStr, `"github.com/stephenafamo/bob/dialect/psql/sm"`) {
		t.Error("Generated file should import Bob sm package")
	}

	// Verify generated code compiles (can be formatted)
	_, err = format.Source(content)
	if err != nil {
		t.Errorf("Generated code does not compile: %v", err)
	}
}
