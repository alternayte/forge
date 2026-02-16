package errors

import (
	"fmt"
	"strings"
)

// Diagnostic represents a rich error with source context and suggestions.
type Diagnostic struct {
	Code           ErrorCode // Error code (e.g., "E001")
	Message        string    // Human-readable error description
	Hint           string    // Suggestion for fixing the error
	File           string    // Source file path
	Line           int       // 1-based line number
	Column         int       // 1-based column number
	SourceLine     string    // The actual line of source code
	UnderlineStart int       // Column where underline begins (0-based within SourceLine)
	UnderlineLen   int       // Length of underline (number of chars to underline)
}

// Error implements the error interface.
func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", d.File, d.Line, d.Column, d.Message)
}

// DiagnosticSet is a collection of diagnostics.
// It implements the error interface so it can be returned as an error.
type DiagnosticSet struct {
	Diagnostics []Diagnostic
}

// HasErrors returns true if the set contains any diagnostics.
func (ds DiagnosticSet) HasErrors() bool {
	return len(ds.Diagnostics) > 0
}

// Error implements the error interface.
// Returns a formatted string of all diagnostics.
func (ds DiagnosticSet) Error() string {
	if !ds.HasErrors() {
		return "no errors"
	}

	var b strings.Builder
	for i, d := range ds.Diagnostics {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(d.Error())
	}
	return b.String()
}

// Count returns the number of diagnostics in the set.
func (ds DiagnosticSet) Count() int {
	return len(ds.Diagnostics)
}

// DiagnosticBuilder provides a fluent API for constructing diagnostics.
type DiagnosticBuilder struct {
	diagnostic Diagnostic
}

// NewDiagnostic creates a new diagnostic builder.
func NewDiagnostic(code ErrorCode, msg string) *DiagnosticBuilder {
	return &DiagnosticBuilder{
		diagnostic: Diagnostic{
			Code:    code,
			Message: msg,
		},
	}
}

// File sets the source file path.
func (b *DiagnosticBuilder) File(file string) *DiagnosticBuilder {
	b.diagnostic.File = file
	return b
}

// Line sets the line number (1-based).
func (b *DiagnosticBuilder) Line(line int) *DiagnosticBuilder {
	b.diagnostic.Line = line
	return b
}

// Column sets the column number (1-based).
func (b *DiagnosticBuilder) Column(col int) *DiagnosticBuilder {
	b.diagnostic.Column = col
	return b
}

// SourceLine sets the actual source code line.
func (b *DiagnosticBuilder) SourceLine(line string) *DiagnosticBuilder {
	b.diagnostic.SourceLine = line
	return b
}

// Underline sets the underline position and length (0-based within source line).
func (b *DiagnosticBuilder) Underline(start, length int) *DiagnosticBuilder {
	b.diagnostic.UnderlineStart = start
	b.diagnostic.UnderlineLen = length
	return b
}

// Hint sets a suggestion for fixing the error.
func (b *DiagnosticBuilder) Hint(hint string) *DiagnosticBuilder {
	b.diagnostic.Hint = hint
	return b
}

// Build returns the constructed Diagnostic.
func (b *DiagnosticBuilder) Build() Diagnostic {
	return b.diagnostic
}
