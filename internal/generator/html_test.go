package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHTML(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Call GenerateHTML with nil resource slice and sample project module
	err := GenerateHTML(nil, tempDir, "github.com/test/myapp")
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	// Verify gen/html/primitives/primitives.templ was generated
	primitivesPath := filepath.Join(tempDir, "html", "primitives", "primitives.templ")
	primitivesContent, err := os.ReadFile(primitivesPath)
	if err != nil {
		t.Fatalf("Failed to read primitives.templ: %v", err)
	}

	primitivesStr := string(primitivesContent)

	// Assert primitives.templ contains all required components
	requiredPrimitivesElements := []string{
		"package primitives",
		"FormField",
		"TextInput",
		"DecimalInput",
		"SelectInput",
		"RelationSelect",
		"data-bind",
	}

	for _, element := range requiredPrimitivesElements {
		if !strings.Contains(primitivesStr, element) {
			t.Errorf("primitives.templ missing required element: %s", element)
		}
	}

	// Verify gen/html/sse/sse.go was generated
	ssePath := filepath.Join(tempDir, "html", "sse", "sse.go")
	sseContent, err := os.ReadFile(ssePath)
	if err != nil {
		t.Fatalf("Failed to read sse.go: %v", err)
	}

	sseStr := string(sseContent)

	// Assert sse.go contains all required elements
	requiredSSEElements := []string{
		"package sse",
		"MergeFragment",
		"Redirect",
		"RedirectTo",
		"PatchElementTempl",
		"datastar",
	}

	for _, element := range requiredSSEElements {
		if !strings.Contains(sseStr, element) {
			t.Errorf("sse.go missing required element: %s", element)
		}
	}
}
