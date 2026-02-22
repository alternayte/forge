package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGeneratePagination(t *testing.T) {
	// Create temporary directory for output
	tempDir := t.TempDir()

	// Mock resource (pagination is generic, doesn't use resource data)
	resources := []parser.ResourceIR{
		{
			Name: "Product",
			Fields: []parser.FieldIR{
				{Name: "ID", Type: "UUID"},
				{Name: "Name", Type: "String"},
			},
		},
	}

	// Generate pagination
	err := GeneratePagination(resources, tempDir, "testmodule")
	if err != nil {
		t.Fatalf("GeneratePagination failed: %v", err)
	}

	// Verify pagination.go was created
	paginationPath := filepath.Join(tempDir, "queries", "pagination.go")
	if _, err := os.Stat(paginationPath); os.IsNotExist(err) {
		t.Fatalf("pagination.go was not created at %s", paginationPath)
	}

	// Read generated content
	content, err := os.ReadFile(paginationPath)
	if err != nil {
		t.Fatalf("Failed to read pagination.go: %v", err)
	}

	contentStr := string(content)

	// Verify it contains expected types and functions
	expectedElements := []string{
		"package queries",
		"type PageInfo struct",
		"HasNextPage",
		"HasPrevPage",
		"StartCursor",
		"EndCursor",
		"TotalCount",
		"type cursor struct",
		"func EncodeCursor",
		"func DecodeCursor",
		"func OffsetPaginationMods",
		"func CursorPaginationMods",
		"sm.Limit",
		"sm.Offset",
		"psql.Raw",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated pagination.go missing expected element: %s", expected)
		}
	}

	// Verify dialect import for correct SelectQuery type
	if !strings.Contains(contentStr, `"github.com/stephenafamo/bob/dialect/psql/dialect"`) {
		t.Error("Generated pagination.go should import Bob dialect package for dialect.SelectQuery type")
	}

	// Verify dialect.SelectQuery is used instead of psql.SelectQuery
	if !strings.Contains(contentStr, "*dialect.SelectQuery") {
		t.Error("Generated pagination.go should use *dialect.SelectQuery, not *psql.SelectQuery")
	}

	// Verify generated code is valid Go (formatGoSource would have failed if not)
	if !strings.HasPrefix(contentStr, "// Code generated") {
		t.Error("Generated file missing DO NOT EDIT header")
	}
}
