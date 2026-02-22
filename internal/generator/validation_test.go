package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGenerateValidation(t *testing.T) {
	// Create test resource with various validation scenarios
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
					{Type: "Required"},
					{Type: "MaxLen", Value: 100},
					{Type: "MinLen", Value: 3},
				},
			},
			{
				Name: "Description",
				Type: "Text",
				Modifiers: []parser.ModifierIR{
					{Type: "MaxLen", Value: 500},
				},
			},
			{
				Name: "Status",
				Type: "Enum",
				EnumValues: []string{"draft", "published", "archived"},
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
				},
			},
			{
				Name: "Email",
				Type: "Email",
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
				},
			},
			{
				Name: "Price",
				Type: "Int",
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
				},
			},
		},
	}

	// Create temporary output directory
	tempDir := t.TempDir()

	// Run generator
	err := GenerateValidation([]parser.ResourceIR{resource}, filepath.Join(tempDir, "gen"), "example.com/testapp")
	if err != nil {
		t.Fatalf("GenerateValidation failed: %v", err)
	}

	// Verify types.go exists
	typesPath := filepath.Join(tempDir, "gen", "validation", "types.go")
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Errorf("Expected types.go to exist at %s", typesPath)
	}

	// Verify product_validation.go exists
	validationPath := filepath.Join(tempDir, "gen", "validation", "product_validation.go")
	if _, err := os.Stat(validationPath); os.IsNotExist(err) {
		t.Fatalf("Expected product_validation.go to exist at %s", validationPath)
	}

	// Read generated file
	content, err := os.ReadFile(validationPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify file contains expected functions
	if !strings.Contains(contentStr, "func ValidateProductCreate") {
		t.Error("Generated file should contain ValidateProductCreate function")
	}

	if !strings.Contains(contentStr, "func ValidateProductUpdate") {
		t.Error("Generated file should contain ValidateProductUpdate function")
	}

	// Verify Required field generates zero-value check
	if !strings.Contains(contentStr, `if input.Name == ""`) {
		t.Error("Generated file should contain required check for Name field")
	}

	// Verify MaxLen check
	if !strings.Contains(contentStr, "len(input.Name) > 100") {
		t.Error("Generated file should contain MaxLen check for Name field")
	}

	// Verify MinLen check
	if !strings.Contains(contentStr, "len(input.Name) < 3") {
		t.Error("Generated file should contain MinLen check for Name field")
	}

	// Verify Enum check
	if !strings.Contains(contentStr, "validValues := []string") {
		t.Error("Generated file should contain enum membership check for Status field")
	}

	// Verify Email format check
	if !strings.Contains(contentStr, `mail.ParseAddress(input.Email)`) {
		t.Error("Generated file should contain email format check")
	}

	// Verify generated code compiles (can be formatted)
	_, err = format.Source(content)
	if err != nil {
		t.Errorf("Generated code does not compile: %v", err)
	}

	// Verify types.go content
	typesContent, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("Failed to read types.go: %v", err)
	}

	typesStr := string(typesContent)
	if !strings.Contains(typesStr, "type FieldError struct") {
		t.Error("types.go should contain FieldError type")
	}

	if !strings.Contains(typesStr, "type ValidationErrors map[string][]FieldError") {
		t.Error("types.go should contain ValidationErrors type")
	}

	if !strings.Contains(typesStr, "func (ve ValidationErrors) Add") {
		t.Error("types.go should contain Add method")
	}

	if !strings.Contains(typesStr, "func (ve ValidationErrors) HasErrors") {
		t.Error("types.go should contain HasErrors method")
	}

	if !strings.Contains(typesStr, "func (ve ValidationErrors) Error") {
		t.Error("types.go should contain Error method")
	}
}
