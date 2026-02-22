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
	return []ToolDef{
		{
			Name:        "templ",
			Version:     "0.3.977",
			URLTemplate: "https://github.com/a-h/templ/releases/download/v{{.Version}}/templ_{{.OS}}_{{.Arch}}.tar.gz",
			BinaryName:  "templ",
			IsArchive:   true,
			Checksums: map[string]string{
				"darwin_arm64": "84ed7b6a3ce2a70928ea59e9b1fc48676eb6788b0ff1a284f1cc5121dcb72ebe",
				"darwin_amd64": "cfc5b40881a6a428632bc3676438b4a7ec9cbe0630c23962dd76018856fe1a92",
				"linux_arm64":  "eb9f6d24221435969278b0336330ac67e0f0a0fef208fe4420121a1e7b043453",
				"linux_amd64":  "231087561d1df2001a0176fc7dda4f0fd8ad0a8ff7b36d268389e43fa3a64073",
			},
		},
		{
			Name:        "sqlc",
			Version:     "1.30.0",
			URLTemplate: "https://downloads.sqlc.dev/sqlc_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
			BinaryName:  "sqlc",
			IsArchive:   true,
			Checksums: map[string]string{
				"darwin_arm64": "ff18793b97715d08dde364446f43082a06da87b7797b9ec79ef2b31aeb0894e5",
				"darwin_amd64": "eb065ca44f02a9500f8e51cb63594a6bbd2486af04d18c0f81efadf7eadf5e29",
				"linux_arm64":  "dd9ab43b022ba3b3402054f99d7ae6e5efea33c949e869c3c66b214415e0c82d",
				"linux_amd64":  "468aecee071bfe55e97fcbcac52ea0208eeca444f67736f3b8f0f3d6a106132e",
			},
		},
		{
			Name:        "tailwind",
			Version:     "4.1.8",
			URLTemplate: "https://github.com/tailwindlabs/tailwindcss/releases/download/v{{.Version}}/tailwindcss-{{.OS}}-{{.Arch}}",
			BinaryName:  "tailwindcss",
			IsArchive:   false, // Standalone binary - zero npm dependency
			Checksums: map[string]string{
				"darwin_arm64": "19e52791d356dd59db68274ae36a5879bab0ce9dac23cc7b0f19fc7b7c1d37a2",
				"darwin_amd64": "4a6cb260d75c4bdca0724fbcc3b23a5adb52715ad6d78595463c86128ca1c329",
				"linux_arm64":  "28a77d1e59b0e45b41683c1e3947621fdfe73f6895b05db7c34f63f3f4898e8d",
				"linux_amd64":  "8f84ce810bdff225e599781d1e2daa82b4282229021c867a71b419f59f9aa836",
			},
			OSMap:   map[string]string{"darwin": "macos"},
			ArchMap: map[string]string{"amd64": "x64"},
		},
		{
			// Atlas uses "latest" channel — checksums are not practical since
			// the binary content changes with each release. Integrity is
			// verified via HTTPS from Ariga's official release endpoint.
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
