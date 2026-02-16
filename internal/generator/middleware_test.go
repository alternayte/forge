package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMiddleware(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Call GenerateMiddleware with empty resource slice
	err := GenerateMiddleware(nil, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateMiddleware failed: %v", err)
	}

	// Verify recovery.go was generated
	recoveryPath := filepath.Join(tempDir, "middleware", "recovery.go")
	recoveryContent, err := os.ReadFile(recoveryPath)
	if err != nil {
		t.Fatalf("Failed to read recovery.go: %v", err)
	}

	recoveryStr := string(recoveryContent)

	// Assert recovery.go contains required elements
	requiredRecoveryElements := []string{
		"func Recovery",
		"recover()",
		"debug.Stack()",
		"slog.Error",
		"ErrorResponder",
		"github.com/example/testapp/gen/errors",
	}

	for _, element := range requiredRecoveryElements {
		if !strings.Contains(recoveryStr, element) {
			t.Errorf("recovery.go missing required element: %s", element)
		}
	}

	// Verify errors.go was generated
	errorsPath := filepath.Join(tempDir, "middleware", "errors.go")
	errorsContent, err := os.ReadFile(errorsPath)
	if err != nil {
		t.Fatalf("Failed to read errors.go: %v", err)
	}

	errorsStr := string(errorsContent)

	// Assert errors.go contains required elements
	requiredErrorsElements := []string{
		"func ErrorResponder",
		"writeSSEError",
		"writeJSONError",
		"writeHTMLError",
		"text/event-stream",
		"application/json",
		"datastar-merge-fragments",
	}

	for _, element := range requiredErrorsElements {
		if !strings.Contains(errorsStr, element) {
			t.Errorf("errors.go missing required element: %s", element)
		}
	}
}
