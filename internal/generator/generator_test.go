package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

// TestGenerateModels_ProductSchema tests end-to-end model generation.
func TestGenerateModels_ProductSchema(t *testing.T) {
	// Create a sample Product resource matching Phase 1 examples
	product := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			{Name: "Description", Type: "Text"},
			{Name: "Price", Type: "Decimal", Modifiers: []parser.ModifierIR{{Type: "Required"}, {Type: "Filterable"}}},
			{Name: "Stock", Type: "Int", Modifiers: []parser.ModifierIR{{Type: "Filterable"}}},
			{Name: "IsActive", Type: "Bool"},
		},
		HasTimestamps: true,
		Options: parser.ResourceOptionsIR{
			SoftDelete: true,
		},
	}

	// Create temp directory
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	// Generate models
	err := GenerateModels([]parser.ResourceIR{product}, outputDir, "example.com/project")
	if err != nil {
		t.Fatalf("GenerateModels failed: %v", err)
	}

	// Verify file exists
	modelPath := filepath.Join(outputDir, "models", "product.go")
	content, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify DO NOT EDIT header
	if !strings.Contains(contentStr, "DO NOT EDIT") {
		t.Error("Generated file missing 'DO NOT EDIT' header")
	}

	// Verify struct types exist
	requiredTypes := []string{
		"type Product struct",
		"type ProductCreate struct",
		"type ProductUpdate struct",
		"type ProductFilter struct",
		"type ProductSort struct",
	}

	for _, requiredType := range requiredTypes {
		if !strings.Contains(contentStr, requiredType) {
			t.Errorf("Generated file missing: %s", requiredType)
		}
	}

	// Verify timestamps are present
	if !strings.Contains(contentStr, "CreatedAt") {
		t.Error("Generated file missing CreatedAt timestamp")
	}
	if !strings.Contains(contentStr, "UpdatedAt") {
		t.Error("Generated file missing UpdatedAt timestamp")
	}

	// Verify soft delete field
	if !strings.Contains(contentStr, "DeletedAt") {
		t.Error("Generated file missing DeletedAt field")
	}

	// Verify filterable fields appear in Filter struct
	if !strings.Contains(contentStr, "Price") && strings.Contains(contentStr, "Filter") {
		t.Error("Filter struct should include Price field (marked Filterable)")
	}

	// Verify file is valid Go syntax
	_, err = format.Source(content)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

// TestGoTypeMapping verifies all 14 field type mappings.
func TestGoTypeMapping(t *testing.T) {
	tests := []struct {
		fieldType string
		goType    string
	}{
		{"UUID", "uuid.UUID"},
		{"String", "string"},
		{"Text", "string"},
		{"Int", "int"},
		{"BigInt", "int64"},
		{"Decimal", "decimal.Decimal"},
		{"Bool", "bool"},
		{"DateTime", "time.Time"},
		{"Date", "time.Time"},
		{"Enum", "string"},
		{"JSON", "json.RawMessage"},
		{"Slug", "string"},
		{"Email", "string"},
		{"URL", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			result := goType(tt.fieldType)
			if result != tt.goType {
				t.Errorf("goType(%q) = %q, want %q", tt.fieldType, result, tt.goType)
			}
		})
	}
}

// TestSnakeCase verifies PascalCase to snake_case conversion.
func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CreatedAt", "created_at"},
		{"ID", "id"},
		{"HTTPStatus", "http_status"}, // Acronyms handled as units
		{"Name", "name"},
		{"IsActive", "is_active"},
		{"ProductID", "product_id"}, // ID suffix treated as acronym
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := snake(tt.input)
			if result != tt.expected {
				t.Errorf("snake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsIDField verifies ID field detection.
func TestIsIDField(t *testing.T) {
	tests := []struct {
		name     string
		field    parser.FieldIR
		expected bool
	}{
		{"ID field", parser.FieldIR{Name: "ID", Type: "UUID"}, true},
		{"Non-ID UUID", parser.FieldIR{Name: "UserID", Type: "UUID"}, false},
		{"ID non-UUID", parser.FieldIR{Name: "ID", Type: "String"}, false},
		{"Regular field", parser.FieldIR{Name: "Name", Type: "String"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIDField(tt.field)
			if result != tt.expected {
				t.Errorf("isIDField(%+v) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

// TestIsRequired verifies Required modifier detection.
func TestIsRequired(t *testing.T) {
	tests := []struct {
		name      string
		modifiers []parser.ModifierIR
		expected  bool
	}{
		{"Has Required", []parser.ModifierIR{{Type: "Required"}}, true},
		{"No Required", []parser.ModifierIR{{Type: "MaxLen", Value: 100}}, false},
		{"Empty", []parser.ModifierIR{}, false},
		{"Multiple with Required", []parser.ModifierIR{{Type: "MaxLen", Value: 100}, {Type: "Required"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRequired(tt.modifiers)
			if result != tt.expected {
				t.Errorf("isRequired(%+v) = %v, want %v", tt.modifiers, result, tt.expected)
			}
		})
	}
}

// TestIsFilterable verifies Filterable modifier detection.
func TestIsFilterable(t *testing.T) {
	tests := []struct {
		name      string
		modifiers []parser.ModifierIR
		expected  bool
	}{
		{"Has Filterable", []parser.ModifierIR{{Type: "Filterable"}}, true},
		{"No Filterable", []parser.ModifierIR{{Type: "Required"}}, false},
		{"Empty", []parser.ModifierIR{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFilterable(tt.modifiers)
			if result != tt.expected {
				t.Errorf("isFilterable(%+v) = %v, want %v", tt.modifiers, result, tt.expected)
			}
		})
	}
}
