package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-framework/forge/internal/parser"
)

func TestGenerateSQLCConfig(t *testing.T) {
	// Create temporary directory for output
	tempDir := t.TempDir()

	// Mock resource (SQLC config is static, doesn't use resource data)
	resources := []parser.ResourceIR{
		{
			Name: "Product",
			Fields: []parser.FieldIR{
				{Name: "ID", Type: "UUID"},
			},
		},
	}

	// Generate SQLC config (projectRoot = tempDir, outputDir would be tempDir/gen)
	err := GenerateSQLCConfig(resources, filepath.Join(tempDir, "gen"), "testmodule", tempDir)
	if err != nil {
		t.Fatalf("GenerateSQLCConfig failed: %v", err)
	}

	// Verify sqlc.yaml was created at project root
	sqlcPath := filepath.Join(tempDir, "sqlc.yaml")
	if _, err := os.Stat(sqlcPath); os.IsNotExist(err) {
		t.Fatalf("sqlc.yaml was not created at %s", sqlcPath)
	}

	// Read generated content
	content, err := os.ReadFile(sqlcPath)
	if err != nil {
		t.Fatalf("Failed to read sqlc.yaml: %v", err)
	}

	contentStr := string(content)

	// Verify it contains expected configuration
	expectedElements := []string{
		`version: "2"`,
		`engine: "postgresql"`,
		`queries: "queries/custom/"`,
		`schema: "gen/atlas/"`,
		`package: "custom"`,
		`out: "queries/custom"`,
		`sql_package: "pgx/v5"`,
		"emit_json_tags: true",
		"emit_result_struct_pointers: true",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated sqlc.yaml missing expected element: %s", expected)
		}
	}

	// Verify YAML structure (basic check - starts with version)
	if !strings.HasPrefix(strings.TrimSpace(contentStr), `version: "2"`) {
		t.Error("Generated sqlc.yaml does not start with version field")
	}
}
