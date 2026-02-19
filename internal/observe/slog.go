package observe

import (
	"log/slog"
	"os"

	"github.com/forge-framework/forge/internal/config"
)

// NewHandler returns an environment-aware slog handler. When cfg.LogFormat is
// "json" (or FORGE_ENV=production forces it), JSON output is written to stderr
// for structured log aggregation. Otherwise a human-readable text handler is
// returned for local development.
func NewHandler(cfg config.ObserveConfig) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.LogLevel),
	}

	if cfg.LogFormat == "json" {
		return slog.NewJSONHandler(os.Stderr, opts)
	}

	return slog.NewTextHandler(os.Stderr, opts)
}

// parseLevel maps a log level string to the corresponding slog.Level constant.
// Unknown strings default to LevelInfo.
func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
