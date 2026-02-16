package parser

import (
	"testing"

	"github.com/forge-framework/forge/internal/errors"
)

// TestParseSimpleResource tests parsing a basic resource with string fields
func TestParseSimpleResource(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Post = schema.Define("Post",
	schema.String("Title").Required(),
	schema.Text("Body"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(result.Resources))
	}

	res := result.Resources[0]
	if res.Name != "Post" {
		t.Errorf("Expected resource name 'Post', got '%s'", res.Name)
	}

	if len(res.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(res.Fields))
	}

	// Check first field
	field0 := res.Fields[0]
	if field0.Name != "Title" {
		t.Errorf("Expected field[0].Name='Title', got '%s'", field0.Name)
	}
	if field0.Type != "String" {
		t.Errorf("Expected field[0].Type='String', got '%s'", field0.Type)
	}

	// Check for Required modifier
	hasRequired := false
	for _, mod := range field0.Modifiers {
		if mod.Type == "Required" {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		t.Errorf("Expected field[0] to have Required modifier")
	}

	// Check second field
	field1 := res.Fields[1]
	if field1.Name != "Body" {
		t.Errorf("Expected field[1].Name='Body', got '%s'", field1.Name)
	}
	if field1.Type != "Text" {
		t.Errorf("Expected field[1].Type='Text', got '%s'", field1.Type)
	}
}

// TestParseAllFieldTypes tests parsing a resource with all 14 field types
func TestParseAllFieldTypes(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var AllTypes = schema.Define("AllTypes",
	schema.UUID("ID"),
	schema.String("Name"),
	schema.Text("Description"),
	schema.Int("Count"),
	schema.BigInt("BigCount"),
	schema.Decimal("Price"),
	schema.Bool("Active"),
	schema.DateTime("CreatedAt"),
	schema.Date("BirthDate"),
	schema.Enum("Status", "draft", "published"),
	schema.JSON("Metadata"),
	schema.Slug("UrlSlug"),
	schema.Email("EmailAddress"),
	schema.URL("Website"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(result.Resources))
	}

	res := result.Resources[0]
	if len(res.Fields) != 14 {
		t.Fatalf("Expected 14 fields, got %d", len(res.Fields))
	}

	expectedTypes := []string{
		"UUID", "String", "Text", "Int", "BigInt", "Decimal", "Bool",
		"DateTime", "Date", "Enum", "JSON", "Slug", "Email", "URL",
	}

	for i, expected := range expectedTypes {
		if res.Fields[i].Type != expected {
			t.Errorf("Expected field[%d].Type='%s', got '%s'", i, expected, res.Fields[i].Type)
		}
	}
}

// TestParseFieldModifiersWithValues tests parsing field modifiers with their values
func TestParseFieldModifiersWithValues(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Article = schema.Define("Article",
	schema.String("Title").Required().MaxLen(200).MinLen(3).Default("Untitled").Label("Title").Placeholder("Enter title"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	res := result.Resources[0]
	field := res.Fields[0]

	expectedModifiers := map[string]interface{}{
		"Required":    nil,
		"MaxLen":      200,
		"MinLen":      3,
		"Default":     "Untitled",
		"Label":       "Title",
		"Placeholder": "Enter title",
	}

	if len(field.Modifiers) != len(expectedModifiers) {
		t.Fatalf("Expected %d modifiers, got %d", len(expectedModifiers), len(field.Modifiers))
	}

	for _, mod := range field.Modifiers {
		expectedVal, exists := expectedModifiers[mod.Type]
		if !exists {
			t.Errorf("Unexpected modifier: %s", mod.Type)
			continue
		}

		if mod.Type == "Required" {
			// Required has no value
			if mod.Value != nil {
				t.Errorf("Expected Required modifier to have nil value, got %v", mod.Value)
			}
		} else {
			if mod.Value != expectedVal {
				t.Errorf("Expected modifier %s to have value %v, got %v", mod.Type, expectedVal, mod.Value)
			}
		}
	}
}

// TestParseRelationships tests parsing relationship definitions
func TestParseRelationships(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Post = schema.Define("Post",
	schema.String("Title"),
	schema.BelongsTo("Category", "categories").Optional().OnDelete(schema.SetNull),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	res := result.Resources[0]
	if len(res.Relationships) != 1 {
		t.Fatalf("Expected 1 relationship, got %d", len(res.Relationships))
	}

	rel := res.Relationships[0]
	if rel.Name != "Category" {
		t.Errorf("Expected relationship.Name='Category', got '%s'", rel.Name)
	}
	if rel.Type != "BelongsTo" {
		t.Errorf("Expected relationship.Type='BelongsTo', got '%s'", rel.Type)
	}
	if rel.Table != "categories" {
		t.Errorf("Expected relationship.Table='categories', got '%s'", rel.Table)
	}
	if !rel.Optional {
		t.Errorf("Expected relationship.Optional=true, got false")
	}
	if rel.OnDelete != "SetNull" {
		t.Errorf("Expected relationship.OnDelete='SetNull', got '%s'", rel.OnDelete)
	}
}

// TestParseResourceOptions tests parsing resource-level options
func TestParseResourceOptions(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Post = schema.Define("Post",
	schema.SoftDelete(),
	schema.Auditable(),
	schema.Timestamps(),
	schema.String("Title"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	res := result.Resources[0]
	if !res.Options.SoftDelete {
		t.Errorf("Expected Options.SoftDelete=true, got false")
	}
	if !res.Options.Auditable {
		t.Errorf("Expected Options.Auditable=true, got false")
	}
	if !res.HasTimestamps {
		t.Errorf("Expected HasTimestamps=true, got false")
	}
	if len(res.Fields) != 1 {
		t.Fatalf("Expected 1 field (excluding timestamps), got %d", len(res.Fields))
	}
}

// TestParseEnumWithValues tests parsing Enum fields with values and default
func TestParseEnumWithValues(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Post = schema.Define("Post",
	schema.Enum("Status", "draft", "published", "archived").Default("draft"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	res := result.Resources[0]
	field := res.Fields[0]

	if field.Type != "Enum" {
		t.Errorf("Expected field.Type='Enum', got '%s'", field.Type)
	}

	expectedValues := []string{"draft", "published", "archived"}
	if len(field.EnumValues) != len(expectedValues) {
		t.Fatalf("Expected %d enum values, got %d", len(expectedValues), len(field.EnumValues))
	}

	for i, expected := range expectedValues {
		if field.EnumValues[i] != expected {
			t.Errorf("Expected EnumValues[%d]='%s', got '%s'", i, expected, field.EnumValues[i])
		}
	}

	// Check Default modifier
	hasDefault := false
	for _, mod := range field.Modifiers {
		if mod.Type == "Default" && mod.Value == "draft" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Errorf("Expected field to have Default modifier with value 'draft'")
	}
}

// TestRejectDynamicValues tests that dynamic values produce rich error diagnostics
func TestRejectDynamicValues(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var maxLen = 200

var Post = schema.Define("Post",
	schema.String("Title").MaxLen(maxLen),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Fatalf("Expected error for dynamic value, got none")
	}

	// Check that we got a diagnostic error
	diag, ok := result.Errors[0].(errors.Diagnostic)
	if !ok {
		t.Fatalf("Expected Diagnostic error, got %T", result.Errors[0])
	}

	if diag.Code != errors.ErrDynamicValue {
		t.Errorf("Expected error code E001, got %s", diag.Code)
	}

	if diag.Line == 0 {
		t.Errorf("Expected diagnostic to have line number, got 0")
	}

	if diag.Hint == "" {
		t.Errorf("Expected diagnostic to have hint, got empty string")
	}
}

// TestCollectMultipleErrors tests that multiple errors are collected in single pass
func TestCollectMultipleErrors(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var maxLen = 200
var minLen = 3

var Post = schema.Define("Post",
	schema.String("Title").MaxLen(maxLen).MinLen(minLen),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) < 2 {
		t.Fatalf("Expected at least 2 errors, got %d", len(result.Errors))
	}

	// Verify both are diagnostics
	for i, e := range result.Errors {
		if _, ok := e.(errors.Diagnostic); !ok {
			t.Errorf("Expected result.Errors[%d] to be Diagnostic, got %T", i, e)
		}
	}
}

// TestParseFileWithNoSchemaDefine tests graceful handling of files without schema.Define
func TestParseFileWithNoSchemaDefine(t *testing.T) {
	source := `package resources

import "fmt"

func SomeHelper() {
	fmt.Println("This file has no schema.Define calls")
}
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors for file without schema.Define, got %d: %v", len(result.Errors), result.Errors)
	}

	if len(result.Resources) != 0 {
		t.Fatalf("Expected 0 resources, got %d", len(result.Resources))
	}
}

// TestParseAllRelationshipTypes tests all 4 relationship types
func TestParseAllRelationshipTypes(t *testing.T) {
	source := `package resources

import "github.com/forge-framework/forge/internal/schema"

var Post = schema.Define("Post",
	schema.BelongsTo("User", "users"),
	schema.HasMany("Comment", "comments"),
	schema.HasOne("Setting", "settings"),
	schema.ManyToMany("Tag", "tags"),
)
`
	result, err := ParseString(source, "test.go")
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	res := result.Resources[0]
	if len(res.Relationships) != 4 {
		t.Fatalf("Expected 4 relationships, got %d", len(res.Relationships))
	}

	expectedTypes := []string{"BelongsTo", "HasMany", "HasOne", "ManyToMany"}
	for i, expected := range expectedTypes {
		if res.Relationships[i].Type != expected {
			t.Errorf("Expected relationship[%d].Type='%s', got '%s'", i, expected, res.Relationships[i].Type)
		}
	}
}
