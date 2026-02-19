package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-framework/forge/internal/parser"
)

// TestGenerateAtlasSchema_ProductTable tests end-to-end Atlas schema generation.
func TestGenerateAtlasSchema_ProductTable(t *testing.T) {
	// Create a sample Product resource
	product := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			{Name: "Description", Type: "Text"},
			{Name: "Price", Type: "Decimal", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			{Name: "Stock", Type: "Int"},
			{Name: "IsActive", Type: "Bool"},
			{Name: "SKU", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Unique"}}},
			{Name: "Category", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Index"}}},
		},
		HasTimestamps: true,
		Options: parser.ResourceOptionsIR{
			SoftDelete: true,
		},
	}

	// Create temp directory
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	// Generate Atlas schema
	err := GenerateAtlasSchema([]parser.ResourceIR{product}, outputDir)
	if err != nil {
		t.Fatalf("GenerateAtlasSchema failed: %v", err)
	}

	// Verify file exists
	schemaPath := filepath.Join(outputDir, "atlas", "schema.hcl")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify DO NOT EDIT header
	if !strings.Contains(contentStr, "DO NOT EDIT") {
		t.Error("Generated file missing 'DO NOT EDIT' header")
	}

	// Verify schema declaration
	if !strings.Contains(contentStr, `schema "public"`) {
		t.Error("Generated file missing schema declaration")
	}

	// Verify table name is pluralized and snake_case
	if !strings.Contains(contentStr, `table "products"`) {
		t.Error("Generated file missing table 'products'")
	}

	// Verify id column with UUID type
	if !strings.Contains(contentStr, `column "id"`) {
		t.Error("Generated file missing id column")
	}
	if !strings.Contains(contentStr, "type    = uuid") {
		t.Error("Generated file missing uuid type for id")
	}
	if !strings.Contains(contentStr, `default = sql("gen_random_uuid()")`) {
		t.Error("Generated file missing default uuid generation")
	}

	// Verify field columns exist
	requiredColumns := []string{
		`column "name"`,
		`column "description"`,
		`column "price"`,
		`column "stock"`,
		`column "is_active"`,
		`column "sku"`,
		`column "category"`,
	}
	for _, col := range requiredColumns {
		if !strings.Contains(contentStr, col) {
			t.Errorf("Generated file missing: %s", col)
		}
	}

	// Verify null constraints (Required -> null = false)
	if !strings.Contains(contentStr, "null = false") {
		t.Error("Generated file missing null constraints")
	}

	// Verify primary key
	if !strings.Contains(contentStr, "primary_key") {
		t.Error("Generated file missing primary key")
	}
	if !strings.Contains(contentStr, "columns = [column.id]") {
		t.Error("Generated file primary key not on id column")
	}

	// Verify timestamps
	if !strings.Contains(contentStr, `column "created_at"`) {
		t.Error("Generated file missing created_at timestamp")
	}
	if !strings.Contains(contentStr, `column "updated_at"`) {
		t.Error("Generated file missing updated_at timestamp")
	}
	if !strings.Contains(contentStr, "timestamptz") {
		t.Error("Generated file missing timestamptz type for timestamps")
	}

	// Verify soft delete column
	if !strings.Contains(contentStr, `column "deleted_at"`) {
		t.Error("Generated file missing deleted_at column")
	}

	// Verify unique index on SKU — with SoftDelete=true, a partial unique index is generated
	// (WHERE deleted_at IS NULL) so unique rows are enforced only among non-deleted records.
	if !strings.Contains(contentStr, `index "products_sku_unique_active"`) {
		t.Error("Generated file missing partial unique index on SKU (soft-delete resource)")
	}
	if !strings.Contains(contentStr, "unique  = true") {
		t.Error("Generated file missing unique = true for SKU index")
	}
	if !strings.Contains(contentStr, `where   = "deleted_at IS NULL"`) {
		t.Error("Generated file missing WHERE clause on soft-delete unique index")
	}

	// Verify regular index on Category
	if !strings.Contains(contentStr, `index "products_category_idx"`) {
		t.Error("Generated file missing regular index on Category")
	}
}

// TestAtlasTypeMapping verifies all 14 field types map to correct PostgreSQL types.
func TestAtlasTypeMapping(t *testing.T) {
	tests := []struct {
		fieldType  string
		atlasType  string
	}{
		{"UUID", "uuid"},
		{"String", "varchar(255)"},
		{"Text", "text"},
		{"Int", "integer"},
		{"BigInt", "bigint"},
		{"Decimal", "numeric(10,2)"},
		{"Bool", "boolean"},
		{"DateTime", "timestamptz"},
		{"Date", "date"},
		{"Enum", "text"},
		{"JSON", "jsonb"},
		{"Slug", "varchar(255)"},
		{"Email", "varchar(255)"},
		{"URL", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			result := atlasType(tt.fieldType)
			if result != tt.atlasType {
				t.Errorf("atlasType(%q) = %q, want %q", tt.fieldType, result, tt.atlasType)
			}
		})
	}
}

// TestAtlasUniqueIndex verifies Unique modifier generates unique index.
func TestAtlasUniqueIndex(t *testing.T) {
	resource := parser.ResourceIR{
		Name: "User",
		Fields: []parser.FieldIR{
			{Name: "Email", Type: "Email", Modifiers: []parser.ModifierIR{{Type: "Unique"}, {Type: "Required"}}},
		},
		HasTimestamps: false,
	}

	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	err := GenerateAtlasSchema([]parser.ResourceIR{resource}, outputDir)
	if err != nil {
		t.Fatalf("GenerateAtlasSchema failed: %v", err)
	}

	schemaPath := filepath.Join(outputDir, "atlas", "schema.hcl")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify unique index generated
	if !strings.Contains(contentStr, `index "users_email_unique"`) {
		t.Error("Generated file missing unique index on Email")
	}
	if !strings.Contains(contentStr, `columns = [column.email]`) {
		t.Error("Generated file unique index not on email column")
	}
	if !strings.Contains(contentStr, "unique  = true") {
		t.Error("Generated file missing unique = true for Email index")
	}
}

// TestAtlasNull verifies null constraints based on Required modifier.
func TestAtlasNull(t *testing.T) {
	tests := []struct {
		name      string
		modifiers []parser.ModifierIR
		expected  string
	}{
		{"Required field", []parser.ModifierIR{{Type: "Required"}}, "false"},
		{"Optional field", []parser.ModifierIR{}, "true"},
		{"With other modifiers", []parser.ModifierIR{{Type: "MaxLen", Value: 100}}, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := atlasNull(tt.modifiers)
			if result != tt.expected {
				t.Errorf("atlasNull(%+v) = %q, want %q", tt.modifiers, result, tt.expected)
			}
		})
	}
}

// TestAtlasTypeWithModifiers verifies MaxLen modifier overrides varchar length.
func TestAtlasTypeWithModifiers(t *testing.T) {
	tests := []struct {
		name     string
		field    parser.FieldIR
		expected string
	}{
		{
			"String with MaxLen",
			parser.FieldIR{Type: "String", Modifiers: []parser.ModifierIR{{Type: "MaxLen", Value: 100}}},
			"varchar(100)",
		},
		{
			"String without MaxLen",
			parser.FieldIR{Type: "String", Modifiers: []parser.ModifierIR{}},
			"varchar(255)",
		},
		{
			"Email with MaxLen",
			parser.FieldIR{Type: "Email", Modifiers: []parser.ModifierIR{{Type: "MaxLen", Value: 200}}},
			"varchar(200)",
		},
		{
			"Text (no MaxLen support)",
			parser.FieldIR{Type: "Text", Modifiers: []parser.ModifierIR{}},
			"text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := atlasTypeWithModifiers(tt.field)
			if result != tt.expected {
				t.Errorf("atlasTypeWithModifiers(%+v) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}

// TestAtlasDefault verifies default value formatting.
func TestAtlasDefault(t *testing.T) {
	tests := []struct {
		name     string
		field    parser.FieldIR
		expected string
	}{
		{
			"String default",
			parser.FieldIR{Type: "String", Modifiers: []parser.ModifierIR{{Type: "Default", Value: "active"}}},
			`"active"`,
		},
		{
			"Bool default true",
			parser.FieldIR{Type: "Bool", Modifiers: []parser.ModifierIR{{Type: "Default", Value: true}}},
			"true",
		},
		{
			"Bool default false",
			parser.FieldIR{Type: "Bool", Modifiers: []parser.ModifierIR{{Type: "Default", Value: false}}},
			"false",
		},
		{
			"Int default",
			parser.FieldIR{Type: "Int", Modifiers: []parser.ModifierIR{{Type: "Default", Value: 42}}},
			"42",
		},
		{
			"No default",
			parser.FieldIR{Type: "String", Modifiers: []parser.ModifierIR{}},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := atlasDefault(tt.field)
			if result != tt.expected {
				t.Errorf("atlasDefault(%+v) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}
