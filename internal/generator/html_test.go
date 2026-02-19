package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGenerateHTML(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Use a sample resource so register_all.go includes registry.Get
	resources := []parser.ResourceIR{
		{Name: "Product"},
	}

	// Call GenerateHTML with a resource and sample project module
	err := GenerateHTML(resources, tempDir, "github.com/test/myapp")
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

	// Verify gen/html/register_all.go was generated
	registerPath := filepath.Join(tempDir, "html", "register_all.go")
	registerContent, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatalf("Failed to read register_all.go: %v", err)
	}

	registerStr := string(registerContent)

	// Assert register_all.go contains all required elements
	requiredRegisterElements := []string{
		"package html",
		"RegisterAllHTMLRoutes",
		"registry.Get",
	}

	for _, element := range requiredRegisterElements {
		if !strings.Contains(registerStr, element) {
			t.Errorf("register_all.go missing required element: %s", element)
		}
	}
}
