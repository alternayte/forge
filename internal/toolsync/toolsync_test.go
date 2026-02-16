package toolsync

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	if platform.OS == "" {
		t.Error("DetectPlatform returned empty OS")
	}

	if platform.Arch == "" {
		t.Error("DetectPlatform returned empty Arch")
	}

	// Should match runtime values
	if platform.OS != runtime.GOOS {
		t.Errorf("DetectPlatform OS = %s, want %s", platform.OS, runtime.GOOS)
	}

	if platform.Arch != runtime.GOARCH {
		t.Errorf("DetectPlatform Arch = %s, want %s", platform.Arch, runtime.GOARCH)
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{Platform{OS: "darwin", Arch: "arm64"}, "darwin_arm64"},
		{Platform{OS: "linux", Arch: "amd64"}, "linux_amd64"},
		{Platform{OS: "windows", Arch: "amd64"}, "windows_amd64"},
	}

	for _, tt := range tests {
		got := tt.platform.String()
		if got != tt.want {
			t.Errorf("Platform.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestPlatformValidate(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		wantErr  bool
	}{
		{"valid darwin arm64", Platform{OS: "darwin", Arch: "arm64"}, false},
		{"valid linux amd64", Platform{OS: "linux", Arch: "amd64"}, false},
		{"valid windows amd64", Platform{OS: "windows", Arch: "amd64"}, false},
		{"invalid OS", Platform{OS: "freebsd", Arch: "amd64"}, true},
		{"invalid arch", Platform{OS: "darwin", Arch: "386"}, true},
		{"invalid both", Platform{OS: "solaris", Arch: "sparc"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.platform.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindTool(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		wantFound bool
		wantName  string
	}{
		{"find templ", "templ", true, "templ"},
		{"find sqlc", "sqlc", true, "sqlc"},
		{"find tailwind", "tailwind", true, "tailwind"},
		{"find atlas", "atlas", true, "atlas"},
		{"not found", "nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, found := FindTool(tt.toolName)

			if found != tt.wantFound {
				t.Errorf("FindTool() found = %v, want %v", found, tt.wantFound)
			}

			if found && tool.Name != tt.wantName {
				t.Errorf("FindTool() tool.Name = %s, want %s", tool.Name, tt.wantName)
			}
		})
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	if len(registry) == 0 {
		t.Fatal("DefaultRegistry returned empty slice")
	}

	// Check that all expected tools are present
	expectedTools := []string{"templ", "sqlc", "tailwind", "atlas"}
	foundTools := make(map[string]bool)

	for _, tool := range registry {
		foundTools[tool.Name] = true

		// Verify required fields are set
		if tool.Name == "" {
			t.Error("Tool has empty Name")
		}
		if tool.Version == "" {
			t.Error("Tool has empty Version")
		}
		if tool.URLTemplate == "" {
			t.Error("Tool has empty URLTemplate")
		}
		if tool.BinaryName == "" {
			t.Error("Tool has empty BinaryName")
		}
	}

	for _, name := range expectedTools {
		if !foundTools[name] {
			t.Errorf("Expected tool %s not found in registry", name)
		}
	}
}

func TestConstructURL(t *testing.T) {
	tests := []struct {
		name        string
		urlTemplate string
		platform    Platform
		version     string
		wantContain []string // Strings that should be in the result
		wantErr     bool
	}{
		{
			name:        "templ URL",
			urlTemplate: "https://github.com/a-h/templ/releases/download/v{{.Version}}/templ_{{.OS}}_{{.Arch}}.tar.gz",
			platform:    Platform{OS: "darwin", Arch: "arm64"},
			version:     "0.2.793",
			wantContain: []string{"v0.2.793", "darwin", "arm64", "templ_darwin_arm64.tar.gz"},
			wantErr:     false,
		},
		{
			name:        "tailwind URL",
			urlTemplate: "https://github.com/tailwindlabs/tailwindcss/releases/download/v{{.Version}}/tailwindcss-{{.OS}}-{{.Arch}}",
			platform:    Platform{OS: "linux", Arch: "amd64"},
			version:     "3.4.17",
			wantContain: []string{"v3.4.17", "tailwindcss-linux-amd64"},
			wantErr:     false,
		},
		{
			name:        "atlas URL",
			urlTemplate: "https://release.ariga.io/atlas/atlas-{{.OS}}-{{.Arch}}-latest",
			platform:    Platform{OS: "darwin", Arch: "amd64"},
			version:     "latest",
			wantContain: []string{"atlas-darwin-amd64-latest"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := constructURL(tt.urlTemplate, tt.platform, tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("constructURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for _, substr := range tt.wantContain {
					if !strings.Contains(got, substr) {
						t.Errorf("constructURL() = %s, should contain %s", got, substr)
					}
				}
			}
		})
	}
}

func TestToolBinPath(t *testing.T) {
	tests := []struct {
		destDir  string
		toolName string
		want     string
	}{
		{"/usr/local/bin", "templ", "/usr/local/bin/templ"},
		{".forge/tools", "sqlc", ".forge/tools/sqlc"},
		{"/opt/tools", "tailwindcss", "/opt/tools/tailwindcss"},
	}

	for _, tt := range tests {
		got := ToolBinPath(tt.destDir, tt.toolName)
		if got != tt.want {
			t.Errorf("ToolBinPath(%s, %s) = %s, want %s", tt.destDir, tt.toolName, got, tt.want)
		}
	}
}

func TestIsToolInstalled(t *testing.T) {
	// This test requires actual filesystem operations
	// We test with a non-existent path to verify false case
	if IsToolInstalled("/nonexistent/path", "tool") {
		t.Error("IsToolInstalled() should return false for non-existent path")
	}
}

func TestToolDefArchiveFlags(t *testing.T) {
	// Verify templ and sqlc are archives
	templ, _ := FindTool("templ")
	if !templ.IsArchive {
		t.Error("templ should be marked as IsArchive=true")
	}

	sqlc, _ := FindTool("sqlc")
	if !sqlc.IsArchive {
		t.Error("sqlc should be marked as IsArchive=true")
	}

	// Verify tailwind and atlas are NOT archives (standalone binaries)
	tailwind, _ := FindTool("tailwind")
	if tailwind.IsArchive {
		t.Error("tailwind should be marked as IsArchive=false (standalone binary)")
	}

	atlas, _ := FindTool("atlas")
	if atlas.IsArchive {
		t.Error("atlas should be marked as IsArchive=false (standalone binary)")
	}
}
