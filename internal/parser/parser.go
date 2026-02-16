package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseString parses a schema definition from a string.
// This is primarily for testing.
func ParseString(source, filename string) (*ParseResult, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return extractFromFile(fset, file, []byte(source), filename)
}

// ParseFile parses a single schema file.
func ParseFile(path string) (*ParseResult, error) {
	fset := token.NewFileSet()

	// Read source file for error diagnostics
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return extractFromFile(fset, file, source, path)
}

// ParseDir finds all schema.go files in subdirectories and parses them.
// It looks for the pattern resources/*/schema.go
func ParseDir(dir string) (*ParseResult, error) {
	result := &ParseResult{
		Resources: []ResourceIR{},
		Errors:    []error{},
	}

	// Find all schema.go files
	pattern := filepath.Join(dir, "**", "schema.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// If no matches with ** pattern, try direct subdirectories
	if len(matches) == 0 {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				schemaFile := filepath.Join(dir, entry.Name(), "schema.go")
				if _, err := os.Stat(schemaFile); err == nil {
					matches = append(matches, schemaFile)
				}
			}
		}
	}

	// Parse each file and merge results
	for _, path := range matches {
		fileResult, err := ParseFile(path)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		result.Resources = append(result.Resources, fileResult.Resources...)
		result.Errors = append(result.Errors, fileResult.Errors...)
	}

	return result, nil
}

// extractFromFile extracts resources from a parsed AST file.
func extractFromFile(fset *token.FileSet, file *ast.File, source []byte, filename string) (*ParseResult, error) {
	resources, diagnostics := extractResources(fset, file, source, filename)

	result := &ParseResult{
		Resources: resources,
		Errors:    make([]error, len(diagnostics)),
	}

	// Convert diagnostics to errors
	for i, diag := range diagnostics {
		result.Errors[i] = diag
	}

	return result, nil
}

// extractLineFromSource extracts the line at the given position.
func extractLineFromSource(source []byte, fset *token.FileSet, pos token.Pos) string {
	position := fset.Position(pos)
	lines := strings.Split(string(source), "\n")
	if position.Line > 0 && position.Line <= len(lines) {
		return lines[position.Line-1]
	}
	return ""
}
