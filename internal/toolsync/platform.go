package toolsync

import (
	"fmt"
	"runtime"
)

// Platform represents the target OS and architecture for tool downloads.
type Platform struct {
	OS   string // darwin, linux, windows
	Arch string // amd64, arm64
}

// DetectPlatform returns the current runtime platform.
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// String returns the platform string in the format "os_arch" (e.g., "darwin_arm64").
func (p Platform) String() string {
	return fmt.Sprintf("%s_%s", p.OS, p.Arch)
}

// Validate checks if the OS/Arch combination is supported.
func (p Platform) Validate() error {
	// Supported OS values
	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}

	// Supported Arch values
	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
	}

	if !validOS[p.OS] {
		return fmt.Errorf("unsupported OS: %s (supported: darwin, linux, windows)", p.OS)
	}

	if !validArch[p.Arch] {
		return fmt.Errorf("unsupported architecture: %s (supported: amd64, arm64)", p.Arch)
	}

	return nil
}
