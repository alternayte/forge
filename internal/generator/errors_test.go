package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-framework/forge/internal/parser"
)

// TestGenerateErrors tests end-to-end error generation.
func TestGenerateErrors(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	// Generate errors (empty resource slice since errors are not per-resource)
	err := GenerateErrors([]parser.ResourceIR{}, outputDir, "example.com/project")
	if err != nil {
		t.Fatalf("GenerateErrors failed: %v", err)
	}

	// Verify errors.go was created
	errorsPath := filepath.Join(outputDir, "errors", "errors.go")
	errorsContent, err := os.ReadFile(errorsPath)
	if err != nil {
		t.Fatalf("Failed to read errors.go: %v", err)
	}

	errorsStr := string(errorsContent)

	// Verify DO NOT EDIT header
	if !strings.Contains(errorsStr, "DO NOT EDIT") {
		t.Error("errors.go missing 'DO NOT EDIT' header")
	}

	// Verify Error struct
	if !strings.Contains(errorsStr, "type Error struct") {
		t.Error("errors.go missing 'type Error struct'")
	}

	// Verify Error struct fields
	requiredFields := []string{"Status", "Code", "Message", "Detail", "Err"}
	for _, field := range requiredFields {
		if !strings.Contains(errorsStr, field) {
			t.Errorf("Error struct missing field: %s", field)
		}
	}

	// Verify Error interface methods
	if !strings.Contains(errorsStr, "func (e *Error) Error()") {
		t.Error("errors.go missing Error() method")
	}
	if !strings.Contains(errorsStr, "func (e *Error) Unwrap()") {
		t.Error("errors.go missing Unwrap() method")
	}

	// Verify constructor functions
	constructors := []string{
		"func NotFound",
		"func UniqueViolation",
		"func ForeignKeyViolation",
		"func Unauthorized",
		"func Forbidden",
		"func BadRequest",
		"func InternalError",
		"func NewValidationError",
	}

	for _, constructor := range constructors {
		if !strings.Contains(errorsStr, constructor) {
			t.Errorf("errors.go missing: %s", constructor)
		}
	}

	// Verify db_mapping.go was created
	dbMappingPath := filepath.Join(outputDir, "errors", "db_mapping.go")
	dbMappingContent, err := os.ReadFile(dbMappingPath)
	if err != nil {
		t.Fatalf("Failed to read db_mapping.go: %v", err)
	}

	dbMappingStr := string(dbMappingContent)

	// Verify DO NOT EDIT header
	if !strings.Contains(dbMappingStr, "DO NOT EDIT") {
		t.Error("db_mapping.go missing 'DO NOT EDIT' header")
	}

	// Verify MapDBError function
	if !strings.Contains(dbMappingStr, "func MapDBError") {
		t.Error("db_mapping.go missing 'func MapDBError'")
	}

	// Verify PostgreSQL error code constants
	errorCodes := []string{
		"ErrCodeUniqueViolation",
		"ErrCodeFKViolation",
		"ErrCodeNotNullViolation",
		"ErrCodeCheckViolation",
	}

	for _, code := range errorCodes {
		if !strings.Contains(dbMappingStr, code) {
			t.Errorf("db_mapping.go missing constant: %s", code)
		}
	}

	// Verify pgconn.PgError type assertion
	if !strings.Contains(dbMappingStr, "pgconn.PgError") {
		t.Error("db_mapping.go missing pgconn.PgError import/usage")
	}

	// Verify errors.As usage
	if !strings.Contains(dbMappingStr, "errors.As") {
		t.Error("db_mapping.go should use errors.As for type checking")
	}

	// Verify IsNotFound helper
	if !strings.Contains(dbMappingStr, "func IsNotFound") {
		t.Error("db_mapping.go missing 'func IsNotFound'")
	}
}

// TestGenerateErrors_FileCreation verifies both files are created in correct location.
func TestGenerateErrors_FileCreation(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "gen")

	// Generate errors
	err := GenerateErrors([]parser.ResourceIR{}, outputDir, "example.com/project")
	if err != nil {
		t.Fatalf("GenerateErrors failed: %v", err)
	}

	// Verify directory structure
	errorsDir := filepath.Join(outputDir, "errors")
	if _, err := os.Stat(errorsDir); os.IsNotExist(err) {
		t.Error("errors directory was not created")
	}

	// Verify errors.go exists
	errorsPath := filepath.Join(errorsDir, "errors.go")
	if _, err := os.Stat(errorsPath); os.IsNotExist(err) {
		t.Error("errors.go was not created")
	}

	// Verify db_mapping.go exists
	dbMappingPath := filepath.Join(errorsDir, "db_mapping.go")
	if _, err := os.Stat(dbMappingPath); os.IsNotExist(err) {
		t.Error("db_mapping.go was not created")
	}
}
