package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the forge.toml configuration structure
type Config struct {
	Project  ProjectConfig  `toml:"project"`
	Database DatabaseConfig `toml:"database"`
	Tools    ToolsConfig    `toml:"tools"`
	Server   ServerConfig   `toml:"server"`
}

// ProjectConfig holds project-level settings
type ProjectConfig struct {
	Name    string `toml:"name"`
	Module  string `toml:"module"`
	Version string `toml:"version"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	URL string `toml:"url"`
}

// ToolsConfig holds tool version specifications
type ToolsConfig struct {
	TemplVersion    string `toml:"templ_version"`
	SQLCVersion     string `toml:"sqlc_version"`
	TailwindVersion string `toml:"tailwind_version"`
	AtlasVersion    string `toml:"atlas_version"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

// Load reads and parses a forge.toml file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Default returns a Config with sensible defaults
func Default() Config {
	return Config{
		Project: ProjectConfig{
			Name:    "my-app",
			Module:  "github.com/user/my-app",
			Version: "0.1.0",
		},
		Database: DatabaseConfig{
			URL: "postgres://localhost:5432/my-app?sslmode=disable",
		},
		Tools: ToolsConfig{
			TemplVersion:    "0.2.793",
			SQLCVersion:     "1.27.0",
			TailwindVersion: "3.4.17",
			AtlasVersion:    "0.29.0",
		},
		Server: ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
	}
}
