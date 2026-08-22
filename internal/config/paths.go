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

// UserConfigFile returns ~/.config/spm/config.toml.
func UserConfigFile() string {
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

// DBPath returns ~/.config/spm/spm.db.
func DBPath() string {
	return filepath.Join(UserConfigDir(), "spm.db")
}
