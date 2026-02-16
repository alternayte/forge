package migrate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/forge-framework/forge/internal/ui"
)

// destructivePatterns contains regex patterns that match destructive SQL operations.
var destructivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)DROP\s+TABLE`),
	regexp.MustCompile(`(?i)DROP\s+COLUMN`),
	regexp.MustCompile(`(?i)ALTER\s+COLUMN\s+\S+\s+TYPE`),
	regexp.MustCompile(`(?i)DROP\s+INDEX`),
}

// ContainsDestructiveChange checks if SQL contains any destructive operations.
func ContainsDestructiveChange(sql string) bool {
	for _, pattern := range destructivePatterns {
		if pattern.MatchString(sql) {
			return true
		}
	}
	return false
}

// FindDestructiveChanges returns all lines containing destructive operations.
func FindDestructiveChanges(sql string) []string {
	var changes []string
	lines := strings.Split(sql, "\n")

	for i, line := range lines {
		for _, pattern := range destructivePatterns {
			if pattern.MatchString(line) {
				// Include line number (1-indexed) and the line content (trimmed)
				changes = append(changes, fmt.Sprintf("%s (line %d)", strings.TrimSpace(line), i+1))
				break // Only add each line once even if multiple patterns match
			}
		}
	}

	return changes
}

// DestructiveWarning generates a formatted warning message for destructive changes.
func DestructiveWarning(changes []string) string {
	var b strings.Builder

	// Header with error icon
	b.WriteString("\n")
	b.WriteString(ui.WarnStyle.Render("WARNING: Destructive migration detected!"))
	b.WriteString("\n\n")

	// Explanation
	b.WriteString("This migration contains operations that will permanently delete data:\n")

	// List each destructive change
	for _, change := range changes {
		b.WriteString("  ")
		b.WriteString(ui.ErrorIcon)
		b.WriteString(" ")
		b.WriteString(change)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Instructions for proceeding
	b.WriteString("If you are ")
	b.WriteString(ui.BoldStyle.Render("CERTAIN"))
	b.WriteString(" you want to proceed, run:\n")
	b.WriteString("  ")
	b.WriteString(ui.CommandStyle.Render("forge migrate diff --force"))
	b.WriteString("\n\n")

	b.WriteString("Otherwise, review your schema changes.\n")

	return b.String()
}
