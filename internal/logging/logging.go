// Package logging ...
package logging

import (
	"io"
	"log/slog"
)

// ParseLevel maps a level name (debug, info, warn, error,
// case-insensitive) to its slog level. An empty name maps to the
// default level (info, ok=true). Unknown names report ok=false.
func ParseLevel(_ string) (slog.Level, bool) {
	return 0, false // stub: implementation lands in the green commit
}

// Setup configures the process-wide default slog logger to write human
// readable text to out, filtered at the level named by name (unknown
// names fall back to info), and returns the configured logger.
func Setup(_ io.Writer, _ string) *slog.Logger {
	return slog.Default() // stub
}
