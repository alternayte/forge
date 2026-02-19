package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

// TestGenerateFactories_ProductSchema tests end-to-end factory generation.
func TestGenerateFactories_ProductSchema(t *testing.T) {
	// Create a sample Product resource
	product := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Name", Type: "String"},
			{Name: "Description", Type: "Text"},
			{Name: "Price", Type: "Decimal"},
			{Name: "Stock", Type: "Int"},
			{Name: "IsActive", Type: "Bool"},
		},
		HasTimestamps: true,
	}

	// Create temp directory
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	// Generate factories
	err := GenerateFactories([]parser.ResourceIR{product}, outputDir, "example.com/project")
	if err != nil {
		t.Fatalf("GenerateFactories failed: %v", err)
	}

	// Verify file exists
	factoryPath := filepath.Join(outputDir, "factories", "product.go")
	content, err := os.ReadFile(factoryPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify DO NOT EDIT header
	if !strings.Contains(contentStr, "DO NOT EDIT") {
		t.Error("Generated file missing 'DO NOT EDIT' header")
	}

	// Verify package declaration
	if !strings.Contains(contentStr, "package factories") {
		t.Error("Generated file missing 'package factories' declaration")
	}

	// Verify imports (should include models package)
	if !strings.Contains(contentStr, `"example.com/project/gen/models"`) {
		t.Error("Generated file missing models package import")
	}

	// Verify New function
	if !strings.Contains(contentStr, "func NewProduct()") {
		t.Error("Generated file missing NewProduct function")
	}
	if !strings.Contains(contentStr, "*models.Product") {
		t.Error("Generated file missing *models.Product return type")
	}

	// Verify Builder type
	if !strings.Contains(contentStr, "type ProductBuilder struct") {
		t.Error("Generated file missing ProductBuilder struct")
	}

	// Verify Build function
	if !strings.Contains(contentStr, "func BuildProduct()") {
		t.Error("Generated file missing BuildProduct function")
	}

	// Verify Build method
	if !strings.Contains(contentStr, "func (b *ProductBuilder) Build()") {
		t.Error("Generated file missing Build method")
	}

	// Verify With methods exist (excluding ID field)
	withMethods := []string{
		"func (b *ProductBuilder) WithName(",
		"func (b *ProductBuilder) WithDescription(",
		"func (b *ProductBuilder) WithPrice(",
		"func (b *ProductBuilder) WithStock(",
		"func (b *ProductBuilder) WithIsActive(",
	}
	for _, method := range withMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("Generated file missing: %s", method)
		}
	}

	// Verify ID field is NOT included in builder (auto-generated)
	if strings.Contains(contentStr, "func (b *ProductBuilder) WithID(") {
		t.Error("Generated file should not have WithID method (ID is auto-generated)")
	}

	// Verify file is valid Go syntax
	_, err = format.Source(content)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

// TestDefaultTestValues verifies all field types produce reasonable default values.
func TestDefaultTestValues(t *testing.T) {
	tests := []struct {
		name     string
		field    parser.FieldIR
		expected string
	}{
		{
			"UUID field",
			parser.FieldIR{Name: "ID", Type: "UUID"},
			"uuid.New()",
		},
		{
			"String field",
			parser.FieldIR{Name: "Name", Type: "String"},
			`"test-name"`,
		},
		{
			"Text field",
			parser.FieldIR{Name: "Description", Type: "Text"},
			`"Test Description content"`,
		},
		{
			"Int field",
			parser.FieldIR{Name: "Stock", Type: "Int"},
			"42",
		},
		{
			"BigInt field",
			parser.FieldIR{Name: "Count", Type: "BigInt"},
			"100000",
		},
		{
			"Decimal field",
			parser.FieldIR{Name: "Price", Type: "Decimal"},
			"decimal.NewFromFloat(9.99)",
		},
		{
			"Bool field",
			parser.FieldIR{Name: "IsActive", Type: "Bool"},
			"true",
		},
		{
			"DateTime field",
			parser.FieldIR{Name: "CreatedAt", Type: "DateTime"},
			"time.Now()",
		},
		{
			"Date field",
			parser.FieldIR{Name: "BirthDate", Type: "Date"},
			"time.Now()",
		},
		{
			"Enum field with values",
			parser.FieldIR{Name: "Status", Type: "Enum", EnumValues: []string{"active", "inactive"}},
			`"active"`,
		},
		{
			"Enum field without values",
			parser.FieldIR{Name: "Status", Type: "Enum", EnumValues: []string{}},
			`"default"`,
		},
		{
			"JSON field",
			parser.FieldIR{Name: "Metadata", Type: "JSON"},
			`json.RawMessage("{}")`,
		},
		{
			"Slug field",
			parser.FieldIR{Name: "Slug", Type: "Slug"},
			`"test-slug"`,
		},
		{
			"Email field",
			parser.FieldIR{Name: "Email", Type: "Email"},
			`"test@example.com"`,
		},
		{
			"URL field",
			parser.FieldIR{Name: "Website", Type: "URL"},
			`"https://example.com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultTestValue(tt.field)
			if result != tt.expected {
				t.Errorf("defaultTestValue(%+v) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}

// TestGenerateFactories_MultipleResources verifies factory generation for multiple resources.
func TestGenerateFactories_MultipleResources(t *testing.T) {
	resources := []parser.ResourceIR{
		{
			Name: "User",
			Fields: []parser.FieldIR{
				{Name: "Email", Type: "Email"},
				{Name: "Name", Type: "String"},
			},
		},
		{
			Name: "Product",
			Fields: []parser.FieldIR{
				{Name: "Name", Type: "String"},
				{Name: "Price", Type: "Decimal"},
			},
		},
	}

	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	err := GenerateFactories(resources, outputDir, "example.com/project")
	if err != nil {
		t.Fatalf("GenerateFactories failed: %v", err)
	}

	// Verify both files exist
	userPath := filepath.Join(outputDir, "factories", "user.go")
	productPath := filepath.Join(outputDir, "factories", "product.go")

	if _, err := os.Stat(userPath); err != nil {
		t.Errorf("User factory file not created: %v", err)
	}

	if _, err := os.Stat(productPath); err != nil {
		t.Errorf("Product factory file not created: %v", err)
	}

	// Verify user factory has correct content
	userContent, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatalf("Failed to read user factory: %v", err)
	}
	if !strings.Contains(string(userContent), "func NewUser()") {
		t.Error("User factory missing NewUser function")
	}

	// Verify product factory has correct content
	productContent, err := os.ReadFile(productPath)
	if err != nil {
		t.Fatalf("Failed to read product factory: %v", err)
	}
	if !strings.Contains(string(productContent), "func NewProduct()") {
		t.Error("Product factory missing NewProduct function")
	}
}
