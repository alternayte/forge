package forge

import (
	"fmt"

	"github.com/alternayte/forge/internal/config"
)

// Config is the runtime configuration for a forge application.
// Load from forge.toml with LoadConfig, or construct programmatically for testing.
type Config struct {
	cfg *config.Config
}

// LoadConfig reads forge.toml from the given path and returns a Config.
// Environment variable overrides (FORGE_*) are applied automatically.
func LoadConfig(path string) (Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return Config{}, err
	}
	return Config{cfg: cfg}, nil
}

// DatabaseURL returns the configured database connection string.
func (c Config) DatabaseURL() string {
	return c.cfg.Database.URL
}

// ServerAddr returns the configured host:port for the HTTP server.
func (c Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.cfg.Server.Host, c.cfg.Server.Port)
}
