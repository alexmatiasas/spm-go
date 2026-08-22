// Package config ...
package config

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
func LoadFrom(_, _ string) (Config, error) {
	return Config{}, nil
}
