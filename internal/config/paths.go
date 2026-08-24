package config

import (
	"os"
	"path/filepath"
)

// UserConfigDir returns ~/.config/spm (XDG).
func UserConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "spm")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "spm")
}

// UserConfigFile returns the SPM_CONFIG override or
// ~/.config/spm/config.toml (XDG).
func UserConfigFile() string {
	if p := os.Getenv("SPM_CONFIG"); p != "" {
		return p
	}

	return filepath.Join(UserConfigDir(), "config.toml")
}

// ProjectConfigDir returns .spm/ in current directory.
func ProjectConfigDir() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".spm")
}

// ProjectConfigFile returns .spm/config.toml.
func ProjectConfigFile() string {
	return filepath.Join(ProjectConfigDir(), "config.toml")
}

// DBPath returns the SPM_DB override or ~/.config/spm/spm.db.
func DBPath() string {
	if p := os.Getenv("SPM_DB"); p != "" {
		return p
	}

	return filepath.Join(UserConfigDir(), "spm.db")
}
