package toolsync

// ToolDef defines a downloadable tool with version, URL pattern, and checksums.
type ToolDef struct {
	Name        string            // Tool name (e.g., "templ")
	Version     string            // Pinned version
	URLTemplate string            // URL template with {{.OS}}, {{.Arch}}, {{.Version}} placeholders
	BinaryName  string            // Expected binary filename after extraction
	IsArchive   bool              // true if download is tar.gz/zip that needs extraction
	Checksums   map[string]string // platform string -> SHA256 hex (e.g., "darwin_arm64" -> "abc123...")
	OSMap       map[string]string // optional: runtime OS -> tool OS name (e.g., "darwin" -> "macos")
	ArchMap     map[string]string // optional: runtime Arch -> tool arch name (e.g., "amd64" -> "x64")
}

// DefaultRegistry returns the tool definitions for all supported tools.
func DefaultRegistry() []ToolDef {
	// TODO: Populate actual checksums when versions are pinned
	// For now, empty checksums mean skip verification during development
	return []ToolDef{
		{
			Name:        "templ",
			Version:     "0.3.977",
			URLTemplate: "https://github.com/a-h/templ/releases/download/v{{.Version}}/templ_{{.OS}}_{{.Arch}}.tar.gz",
			BinaryName:  "templ",
			IsArchive:   true,
			Checksums:   map[string]string{},
		},
		{
			Name:        "sqlc",
			Version:     "1.30.0",
			URLTemplate: "https://downloads.sqlc.dev/sqlc_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
			BinaryName:  "sqlc",
			IsArchive:   true,
			Checksums:   map[string]string{},
		},
		{
			Name:        "tailwind",
			Version:     "4.1.8",
			URLTemplate: "https://github.com/tailwindlabs/tailwindcss/releases/download/v{{.Version}}/tailwindcss-{{.OS}}-{{.Arch}}",
			BinaryName:  "tailwindcss",
			IsArchive:   false, // Standalone binary - zero npm dependency
			Checksums:   map[string]string{},
			OSMap:       map[string]string{"darwin": "macos"},
			ArchMap:     map[string]string{"amd64": "x64"},
		},
		{
			Name:        "atlas",
			Version:     "latest",
			URLTemplate: "https://release.ariga.io/atlas/atlas-{{.OS}}-{{.Arch}}-latest",
			BinaryName:  "atlas",
			IsArchive:   false,
			Checksums:   map[string]string{},
		},
	}
}

// FindTool looks up a tool by name in the registry.
func FindTool(name string) (*ToolDef, bool) {
	registry := DefaultRegistry()
	for i := range registry {
		if registry[i].Name == name {
			return &registry[i], true
		}
	}
	return nil, false
}
