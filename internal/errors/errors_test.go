package errors

import (
	"strings"
	"testing"
)

func TestFormat_ProducesExpectedStructure(t *testing.T) {
	// Create a diagnostic with all fields populated
	d := NewDiagnostic(ErrDynamicValue, "schema values must use literal values for static analysis").
		File("resources/product/schema.go").
		Line(12).
		Column(34).
		SourceLine(`    schema.String("Title").MaxLen(maxLen),`).
		Underline(34, 6).
		Hint("Forge schemas must use literal values for static analysis.\nFound variable 'maxLen' — use a constant or literal instead.").
		Build()

	formatted := Format(d)

	// Check for key components in output
	expectedSubstrings := []string{
		"error[E001]:",                         // Error code header
		"resources/product/schema.go:12:34",    // File position
		`schema.String("Title").MaxLen(maxLen)`, // Source line
		"^^^^^^",                                // Underline (6 carets)
		"hint:",                                 // Hint label
		"forge help E001",                       // Help command
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(formatted, expected) {
			t.Errorf("Format() output missing expected substring: %q\nGot:\n%s", expected, formatted)
		}
	}
}

func TestFormat_HandlesMinimalDiagnostic(t *testing.T) {
	// Create minimal diagnostic with just code and message
	d := NewDiagnostic(ErrUnsupportedType, "field type not supported").Build()

	formatted := Format(d)

	// Should still produce valid output
	if !strings.Contains(formatted, "error[E002]:") {
		t.Errorf("Format() should include error code even with minimal diagnostic")
	}
	if !strings.Contains(formatted, "field type not supported") {
		t.Errorf("Format() should include message even with minimal diagnostic")
	}
}

func TestDiagnosticSet_CollectsMultipleErrors(t *testing.T) {
	ds := DiagnosticSet{
		Diagnostics: []Diagnostic{
			NewDiagnostic(ErrDuplicateField, "duplicate field 'name'").
				File("schema.go").
				Line(10).
				Column(5).
				Build(),
			NewDiagnostic(ErrMissingPrimaryKey, "resource must have a primary key").
				File("schema.go").
				Line(1).
				Column(1).
				Build(),
		},
	}

	if !ds.HasErrors() {
		t.Error("DiagnosticSet.HasErrors() should return true when diagnostics present")
	}

	if ds.Count() != 2 {
		t.Errorf("DiagnosticSet.Count() = %d, want 2", ds.Count())
	}

	formatted := FormatAll(ds)

	// Should contain both error codes
	if !strings.Contains(formatted, "error[E100]:") {
		t.Error("FormatAll() should include first diagnostic (E100)")
	}
	if !strings.Contains(formatted, "error[E102]:") {
		t.Error("FormatAll() should include second diagnostic (E102)")
	}
}

func TestErrorCodeLookup(t *testing.T) {
	// Test valid lookup
	info := Lookup(ErrDynamicValue)
	if info.Code != ErrDynamicValue {
		t.Errorf("Lookup(ErrDynamicValue).Code = %v, want %v", info.Code, ErrDynamicValue)
	}
	if info.Title == "" {
		t.Error("Lookup(ErrDynamicValue).Title should not be empty")
	}
	if info.HelpURL == "" {
		t.Error("Lookup(ErrDynamicValue).HelpURL should not be empty")
	}

	// Test invalid lookup returns fallback
	info = Lookup("E999")
	if info.Title != "unknown error" {
		t.Errorf("Lookup(invalid code).Title = %q, want 'unknown error'", info.Title)
	}
}

func TestDiagnosticBuilder_FluentAPI(t *testing.T) {
	// Test that builder methods can be chained
	d := NewDiagnostic(ErrInvalidFieldName, "invalid name").
		File("test.go").
		Line(5).
		Column(10).
		SourceLine("some code").
		Underline(0, 4).
		Hint("use a different name").
		Build()

	if d.File != "test.go" {
		t.Errorf("Diagnostic.File = %q, want 'test.go'", d.File)
	}
	if d.Line != 5 {
		t.Errorf("Diagnostic.Line = %d, want 5", d.Line)
	}
	if d.Column != 10 {
		t.Errorf("Diagnostic.Column = %d, want 10", d.Column)
	}
	if d.Hint != "use a different name" {
		t.Errorf("Diagnostic.Hint = %q, want 'use a different name'", d.Hint)
	}
}

func TestDiagnosticSet_ErrorInterface(t *testing.T) {
	ds := DiagnosticSet{
		Diagnostics: []Diagnostic{
			NewDiagnostic(ErrDuplicateField, "duplicate field").
				File("test.go").
				Line(1).
				Column(1).
				Build(),
		},
	}

	// Test that DiagnosticSet implements error interface
	var _ error = ds

	errorMsg := ds.Error()
	if !strings.Contains(errorMsg, "test.go:1:1") {
		t.Errorf("DiagnosticSet.Error() should include file position, got: %s", errorMsg)
	}
}

func TestDiagnostic_ErrorInterface(t *testing.T) {
	d := NewDiagnostic(ErrInvalidFieldName, "invalid name").
		File("test.go").
		Line(5).
		Column(10).
		Build()

	// Test that Diagnostic implements error interface
	var _ error = d

	errorMsg := d.Error()
	expected := "test.go:5:10: invalid name"
	if errorMsg != expected {
		t.Errorf("Diagnostic.Error() = %q, want %q", errorMsg, expected)
	}
}
