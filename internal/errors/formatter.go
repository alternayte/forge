package errors

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for error formatting (Rust-style diagnostics)
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Red
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")) // Blue

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")) // Cyan

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Dim gray

	underlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")) // Red
)

// Format renders a single diagnostic in Rust-style format.
//
// Output structure:
//
//	error[E001]: schema values must use literal values for static analysis
//	  --> resources/product/schema.go:12:34
//	   |
//	12 |     schema.String("Title").MaxLen(maxLen),
//	   |                                   ^^^^^^
//	   |
//	   = hint: Forge schemas must use literal values for static analysis.
//	           Found variable 'maxLen' — use a constant or literal instead.
//	   = help: Run `forge help E001` for more information
func Format(d Diagnostic) string {
	var b strings.Builder

	// Header: error[E001]: message
	header := errorStyle.Render(fmt.Sprintf("error[%s]:", d.Code))
	b.WriteString(header)
	b.WriteString(" ")
	b.WriteString(d.Message)
	b.WriteString("\n")

	// Position: --> file:line:col
	if d.File != "" {
		arrow := fileStyle.Render("  -->")
		position := fmt.Sprintf(" %s:%d:%d", d.File, d.Line, d.Column)
		b.WriteString(arrow)
		b.WriteString(position)
		b.WriteString("\n")
	}

	// Source context with line numbers
	if d.SourceLine != "" {
		// Empty gutter line
		b.WriteString("   ")
		b.WriteString(fileStyle.Render("|"))
		b.WriteString("\n")

		// Line number gutter
		lineNum := fmt.Sprintf("%d", d.Line)
		gutter := fileStyle.Render(lineNum)
		b.WriteString(" ")
		b.WriteString(gutter)
		b.WriteString(" ")
		b.WriteString(fileStyle.Render("|"))
		b.WriteString(" ")
		b.WriteString(d.SourceLine)
		b.WriteString("\n")

		// Underline with carets
		if d.UnderlineLen > 0 {
			// Calculate spaces needed (accounting for gutter width)
			gutterWidth := len(lineNum) + 1 // +1 for space after number
			leadingSpaces := gutterWidth + 2 + d.UnderlineStart // +2 for " | "

			b.WriteString(strings.Repeat(" ", leadingSpaces))
			b.WriteString(fileStyle.Render("|"))
			b.WriteString(strings.Repeat(" ", d.UnderlineStart))
			b.WriteString(underlineStyle.Render(strings.Repeat("^", d.UnderlineLen)))
			b.WriteString("\n")
		}

		// Empty gutter line
		b.WriteString("   ")
		b.WriteString(fileStyle.Render("|"))
		b.WriteString("\n")
	}

	// Hint
	if d.Hint != "" {
		hintLabel := hintStyle.Render("   = hint:")
		b.WriteString(hintLabel)
		b.WriteString(" ")

		// Handle multi-line hints with proper indentation
		hintLines := strings.Split(d.Hint, "\n")
		for i, line := range hintLines {
			if i > 0 {
				b.WriteString("\n           ") // Indent continuation lines
			}
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	// Help link
	info := Lookup(d.Code)
	if info.HelpURL != "" {
		helpLabel := dimStyle.Render("   = help:")
		helpText := fmt.Sprintf(" Run `%s` for more information", info.HelpURL)
		b.WriteString(helpLabel)
		b.WriteString(helpText)
		b.WriteString("\n")
	}

	return b.String()
}

// FormatAll renders all diagnostics in a DiagnosticSet with separators.
func FormatAll(ds DiagnosticSet) string {
	if !ds.HasErrors() {
		return ""
	}

	var b strings.Builder
	for i, d := range ds.Diagnostics {
		if i > 0 {
			// Add separator between diagnostics
			b.WriteString("\n")
		}
		b.WriteString(Format(d))
	}

	return b.String()
}
