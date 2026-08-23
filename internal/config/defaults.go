// Package config ...
package config

// Defaults returns the built-in default configuration.
func Defaults() Config {
	return Config{
		Inventory: InventoryConfig{
			Roots: []string{"~/projects"},
			Exclude: ExcludeConfig{
				Dirs: []string{"Icon", "Credenciales", "notes"},
			},
		},
		Defaults: DefaultsConfig{
			ProjectType: "python",
		},
	}
}
