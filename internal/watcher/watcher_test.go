package watcher

import (
	"testing"
	"time"
)

func TestIsRelevantFile_GoFiles(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"resources/product/schema.go", true},
		{"internal/handlers/product.go", true},
		{"cmd/server/main.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsRelevantFile_TemplFiles(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"internal/views/product.templ", true},
		{"resources/templates/layout.templ", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsRelevantFile_SQLFiles(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"migrations/001_initial.sql", true},
		{"resources/queries/products.sql", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsRelevantFile_CSSFiles(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"assets/styles.css", true},
		{"internal/static/app.css", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsRelevantFile_IgnoredExtensions(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"README.md", false},
		{"config.json", false},
		{"notes.txt", false},
		{"schema.go.swp", false},
		{"temp.tmp", false},
		{"backup~", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsRelevantFile_IgnoreGenDir(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"gen/models/product.go", false},
		{"gen/factories/product.go", false},
		{"gen/atlas/schema.hcl", false},
		{"internal/gen/something.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestNewWatcher_CreatesSuccessfully(t *testing.T) {
	onChange := func() {
		// Callback for watcher
	}

	w, err := New(onChange, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer w.Close()

	if w.fsw == nil {
		t.Error("Watcher.fsw is nil")
	}

	if w.debounce != 100*time.Millisecond {
		t.Errorf("Watcher.debounce = %v, want 100ms", w.debounce)
	}
}

func TestWatcher_Close(t *testing.T) {
	onChange := func() {}

	w, err := New(onChange, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
