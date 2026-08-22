package config

import (
	"os"
	"strconv"
)

// NoTUI reports whether the TUI is disabled via SPM_NO_TUI.
// Any non-empty value that does not parse as an explicit false
// counts as disabled.
func NoTUI() bool {
	v := os.Getenv("SPM_NO_TUI")
	if v == "" {
		return false
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return true
	}

	return b
}

// Editor resolves the editor to use: SPM_EDITOR wins, then fallback.
func Editor(fallback string) string {
	if e := os.Getenv("SPM_EDITOR"); e != "" {
		return e
	}

	return fallback
}

// LogLevel resolves SPM_LOG_LEVEL (empty means unset).
func LogLevel() string {
	return os.Getenv("SPM_LOG_LEVEL")
}
