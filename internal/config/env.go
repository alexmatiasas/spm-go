package config

// NoTUI reports whether the TUI is disabled via SPM_NO_TUI.
func NoTUI() bool {
	return false // stub
}

// Editor resolves the editor to use: SPM_EDITOR wins, then fallback.
func Editor(fallback string) string {
	return fallback // stub
}

// LogLevel resolves SPM_LOG_LEVEL (empty means unset).
func LogLevel() string {
	return "" // stub
}
