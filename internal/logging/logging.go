// Package logging ...
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// DefaultLevel is applied when no level name is configured.
const DefaultLevel = slog.LevelInfo

// ParseLevel maps a level name (debug, info, warn, error,
// case-insensitive) to its slog level. An empty name maps to
// DefaultLevel with ok=true. Unknown names report ok=false and map to
// DefaultLevel so callers can degrade gracefully.
func ParseLevel(name string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	case "":
		return DefaultLevel, true
	default:
		return DefaultLevel, false
	}
}

// Setup configures the process-wide default slog logger to write human
// readable text to out, filtered at the level named by name (unknown
// names fall back to DefaultLevel), and returns the configured logger.
func Setup(out io.Writer, name string) *slog.Logger {
	level, _ := ParseLevel(name)

	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	return logger
}
