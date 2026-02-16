package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Icon constants for consistent CLI output
var (
	// SuccessIcon is a green checkmark
	SuccessIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Render("✓")

	// ErrorIcon is a red cross
	ErrorIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Render("✗")

	// WarnIcon is a yellow warning symbol
	WarnIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Render("!")

	// InfoIcon is a blue info symbol
	InfoIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Render("i")
)

// Reusable style objects for consistent formatting
var (
	// HeaderStyle is used for section headers
	HeaderStyle = lipgloss.NewStyle().
			Bold(true)

	// SuccessStyle renders text in green
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	// ErrorStyle renders text in red and bold
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	// WarnStyle renders text in yellow
	WarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	// DimStyle renders text dimmed/faint for secondary info
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// BoldStyle renders text in bold
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// FilePathStyle renders file paths in cyan
	FilePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))

	// CommandStyle renders CLI commands in bold cyan
	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true)
)

// Success renders a success message with checkmark icon.
// Example output: "  ✓ Project initialized"
func Success(msg string) string {
	return "  " + SuccessIcon + " " + msg
}

// Error renders an error message with cross icon.
// Example output: "  ✗ Failed to parse schema"
func Error(msg string) string {
	return "  " + ErrorIcon + " " + msg
}

// Warn renders a warning message with warning icon.
// Example output: "  ! Deprecated field type"
func Warn(msg string) string {
	return "  " + WarnIcon + " " + msg
}

// Info renders an info message with info icon.
// Example output: "  i Using cached dependencies"
func Info(msg string) string {
	return "  " + InfoIcon + " " + msg
}

// Header renders a bold header.
// Example output: "Compiling resources..."
func Header(msg string) string {
	return HeaderStyle.Render(msg)
}

// Grouped renders a header followed by indented items (Cargo-style grouped output).
// Example:
//
//	Compiling resources
//	  ✓ Product
//	  ✓ Category
//	  ✓ User
func Grouped(header string, items []string) string {
	var b strings.Builder

	// Render header
	b.WriteString(Header(header))
	b.WriteString("\n")

	// Render items with indentation
	for _, item := range items {
		b.WriteString(item)
		b.WriteString("\n")
	}

	return b.String()
}
