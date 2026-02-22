package parser

import (
	"os"
	"testing"
)

// TestParseProductSchemaExample tests parsing the Product example from CONTEXT.md
func TestParseProductSchemaExample(t *testing.T) {
	source := `package resources

import "github.com/alternayte/forge/schema"

var Product = schema.Define("Product",
	schema.SoftDelete(),
	schema.UUID("ID").PrimaryKey(),
	schema.String("Title").Required().MaxLen(200).Label("Product Title"),
	schema.Text("Description").Help("Full product description"),
	schema.Decimal("Price").Required().Filterable().Sortable(),
	schema.Enum("Status", "draft", "published", "archived").Default("draft"),
	schema.Bool("Featured").Default(false),
	schema.BelongsTo("Category", "categories").Optional().OnDelete(schema.SetNull),
	schema.HasMany("Reviews", "reviews"),
	schema.Timestamps(),
)
`

	result, err := ParseString(source, "product_schema.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d errors", len(result.Errors))
		for _, e := range result.Errors {
			t.Logf("Error: %v", e)
		}
	}

	if len(result.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(result.Resources))
	}

	product := result.Resources[0]

	// Verify resource name
	if product.Name != "Product" {
		t.Errorf("Expected name 'Product', got '%s'", product.Name)
	}

	// Verify options
	if !product.Options.SoftDelete {
		t.Errorf("Expected SoftDelete=true")
	}

	// Verify timestamps
	if !product.HasTimestamps {
		t.Errorf("Expected HasTimestamps=true")
	}

	// Verify fields
	expectedFieldCount := 6 // ID, Title, Description, Price, Status, Featured
	if len(product.Fields) != expectedFieldCount {
		t.Fatalf("Expected %d fields, got %d", expectedFieldCount, len(product.Fields))
	}

	// Verify field types
	fieldTypes := make(map[string]string)
	for _, field := range product.Fields {
		fieldTypes[field.Name] = field.Type
	}

	if fieldTypes["ID"] != "UUID" {
		t.Errorf("Expected ID to be UUID, got %s", fieldTypes["ID"])
	}
	if fieldTypes["Title"] != "String" {
		t.Errorf("Expected Title to be String, got %s", fieldTypes["Title"])
	}
	if fieldTypes["Status"] != "Enum" {
		t.Errorf("Expected Status to be Enum, got %s", fieldTypes["Status"])
	}

	// Verify relationships
	expectedRelCount := 2 // Category (BelongsTo), Reviews (HasMany)
	if len(product.Relationships) != expectedRelCount {
		t.Fatalf("Expected %d relationships, got %d", expectedRelCount, len(product.Relationships))
	}

	// Find Category relationship
	var categoryRel *RelationshipIR
	for i := range product.Relationships {
		if product.Relationships[i].Name == "Category" {
			categoryRel = &product.Relationships[i]
			break
		}
	}

	if categoryRel == nil {
		t.Fatal("Expected Category relationship not found")
	}

	if categoryRel.Type != "BelongsTo" {
		t.Errorf("Expected Category type BelongsTo, got %s", categoryRel.Type)
	}
	if categoryRel.Table != "categories" {
		t.Errorf("Expected Category table 'categories', got '%s'", categoryRel.Table)
	}
	if !categoryRel.Optional {
		t.Errorf("Expected Category to be Optional")
	}
	if categoryRel.OnDelete != "SetNull" {
		t.Errorf("Expected Category OnDelete='SetNull', got '%s'", categoryRel.OnDelete)
	}

	t.Logf("Successfully parsed Product schema:")
	t.Logf("  - %d fields", len(product.Fields))
	t.Logf("  - %d relationships", len(product.Relationships))
	t.Logf("  - SoftDelete: %v", product.Options.SoftDelete)
	t.Logf("  - Timestamps: %v", product.HasTimestamps)
}

// TestParseFileFromDisk tests parsing a real file from disk
func TestParseFileFromDisk(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "schema_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `package resources

import "github.com/alternayte/forge/schema"

var User = schema.Define("User",
	schema.String("Email").Required(),
	schema.String("Name"),
)
`

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// Parse the file
	result, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d", len(result.Errors))
	}

	if len(result.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(result.Resources))
	}

	if result.Resources[0].Name != "User" {
		t.Errorf("Expected resource name 'User', got '%s'", result.Resources[0].Name)
	}
}
