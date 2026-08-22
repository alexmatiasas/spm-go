package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config ...
type Config struct {
	Inventory InventoryConfig `toml:"inventory"`
	Defaults  DefaultsConfig  `toml:"defaults"`
	Shell     ShellConfig     `toml:"shell"`
}

// InventoryConfig ...
type InventoryConfig struct {
	Roots   []string      `toml:"roots"`
	Exclude ExcludeConfig `toml:"exclude"`
}

// ExcludeConfig ...
type ExcludeConfig struct {
	Dirs []string `toml:"dirs"`
}

// DefaultsConfig ...
type DefaultsConfig struct {
	ProjectType string `toml:"project_type"`
}

// ShellConfig ...
type ShellConfig struct {
	AutoCD bool `toml:"auto_cd"`
}

// Load loads configuration by precedence order:
// defaults → user config → project config → env vars.
func Load() (Config, error) {
	return LoadFrom(UserConfigFile(), ProjectConfigFile())
}

// LoadFrom loads configuration from explicit user and project config
// file paths, applying defaults first and merging each layer over them.
// Parameter order matters: the user layer merges over defaults, the
// project layer over the user layer.
func LoadFrom(userPath, projectPath string) (Config, error) {
	cfg := Defaults()

	layers := []struct{ name, path string }{
		{"user", userPath},
		{"project", projectPath},
	}

	for _, layer := range layers {
		if err := mergeFile(layer.path, &cfg); err != nil {
			return cfg, fmt.Errorf("config: %s layer %s: %w", layer.name, layer.path, err)
		}
	}

	return cfg, nil
}

// mergeFile decodes path over cfg. Keys present in the file overwrite;
// keys absent keep their current values. A missing file is not an error.
func mergeFile(path string, cfg *Config) error {
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}
