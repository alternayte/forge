package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the forge.toml configuration structure
type Config struct {
	Project  ProjectConfig  `toml:"project"`
	Database DatabaseConfig `toml:"database"`
	Tools    ToolsConfig    `toml:"tools"`
	Server   ServerConfig   `toml:"server"`
	Session  SessionConfig  `toml:"session"`
	Jobs     JobsConfig     `toml:"jobs"`
	SSE      SSEConfig      `toml:"sse"`
	Observe  ObserveConfig  `toml:"telemetry"`
	Admin    AdminConfig    `toml:"admin"`
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

// SessionConfig holds session cookie and store settings
type SessionConfig struct {
	// Secret is the HMAC signing key for session data integrity. Should be a
	// 32- or 64-byte random string. Not required when using pgxstore (server-
	// side storage), but reserved for future HMAC signing use.
	Secret string `toml:"secret"`

	// Secure controls whether the session cookie is sent only over HTTPS.
	// Set to true in production, false during local development.
	Secure bool `toml:"secure"`

	// Lifetime is the session duration expressed as a Go duration string
	// (e.g. "24h", "168h"). Defaults to 24h when empty.
	Lifetime string `toml:"lifetime"`
}

// JobsConfig holds background job processing settings
type JobsConfig struct {
	Enabled bool           `toml:"enabled"`
	Queues  map[string]int `toml:"queues"` // queue_name -> max_workers
}

// SSEConfig holds Server-Sent Events connection settings
type SSEConfig struct {
	MaxTotalConnections int `toml:"max_total_connections"` // default: 5000
	MaxPerUser          int `toml:"max_per_user"`          // default: 10
	BufferSize          int `toml:"buffer_size"`           // default: 32
}

// ObserveConfig holds observability and telemetry settings
type ObserveConfig struct {
	OTLPEndpoint string `toml:"otlp_endpoint"` // empty = stdout in dev
	LogLevel     string `toml:"log_level"`     // debug, info, warn, error
	LogFormat    string `toml:"log_format"`    // json or text
}

// AdminConfig holds the admin/ops HTTP server settings
type AdminConfig struct {
	Port int `toml:"port"` // default: 9090
}

// ApplyEnvOverrides overlays FORGE_* environment variables on top of
// any values loaded from forge.toml. Environment variables always win
// (12-factor app config, DEPLOY-01).
func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("FORGE_DATABASE_URL"); v != "" {
		c.Database.URL = v
	}

	if v := os.Getenv("FORGE_SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Server.Port = n
		} else {
			slog.Warn("FORGE_SERVER_PORT is not a valid integer, ignoring", "value", v)
		}
	}

	if v := os.Getenv("FORGE_SERVER_HOST"); v != "" {
		c.Server.Host = v
	}

	if v := os.Getenv("FORGE_JOBS_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Jobs.Enabled = b
		} else {
			slog.Warn("FORGE_JOBS_ENABLED is not a valid boolean, ignoring", "value", v)
		}
	}

	if v := os.Getenv("FORGE_SSE_MAX_TOTAL_CONNECTIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.SSE.MaxTotalConnections = n
		} else {
			slog.Warn("FORGE_SSE_MAX_TOTAL_CONNECTIONS is not a valid integer, ignoring", "value", v)
		}
	}

	if v := os.Getenv("FORGE_SSE_MAX_PER_USER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.SSE.MaxPerUser = n
		} else {
			slog.Warn("FORGE_SSE_MAX_PER_USER is not a valid integer, ignoring", "value", v)
		}
	}

	if v := os.Getenv("FORGE_OTEL_ENDPOINT"); v != "" {
		c.Observe.OTLPEndpoint = v
	}

	if v := os.Getenv("FORGE_LOG_LEVEL"); v != "" {
		c.Observe.LogLevel = v
	}

	if v := os.Getenv("FORGE_LOG_FORMAT"); v != "" {
		c.Observe.LogFormat = v
	}

	if v := os.Getenv("FORGE_ADMIN_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Admin.Port = n
		} else {
			slog.Warn("FORGE_ADMIN_PORT is not a valid integer, ignoring", "value", v)
		}
	}

	// FORGE_ENV=production forces JSON log format for structured log aggregation
	if v := os.Getenv("FORGE_ENV"); v == "production" {
		c.Observe.LogFormat = "json"
	}

	if v := os.Getenv("FORGE_SESSION_SECRET"); v != "" {
		c.Session.Secret = v
	}
}

// Load reads and parses a forge.toml file, then applies any FORGE_* env var
// overrides so environment variables always win over file-based config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	config.ApplyEnvOverrides()

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
		Jobs: JobsConfig{
			Enabled: false,
		},
		SSE: SSEConfig{
			MaxTotalConnections: 5000,
			MaxPerUser:          10,
			BufferSize:          32,
		},
		Observe: ObserveConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Admin: AdminConfig{
			Port: 9090,
		},
	}
}
