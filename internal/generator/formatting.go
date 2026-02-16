package generator

import (
	"golang.org/x/tools/imports"
)

// FormatGoSource formats Go source code and manages imports automatically.
// It takes a filename (for context) and raw source bytes, returning formatted
// source with imports added/removed as needed, or an error if formatting fails.
func FormatGoSource(filename string, source []byte) ([]byte, error) {
	return imports.Process(filename, source, nil)
}
