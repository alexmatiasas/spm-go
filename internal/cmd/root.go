// Package cmd ...
package cmd

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/alexmatiasas/spm/internal/config"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the spm command tree. version and commit are
// reported by the version subcommand.
func NewRootCmd(version, commit string) *cobra.Command {
	root := &cobra.Command{
		Use:          "spm",
		Short:        "Smart Project Manager",
		Long:         "spm scaffolds projects with the right guardrails and keeps an inventory of everything you build.",
		SilenceUsage: true,
	}

	root.AddCommand(newVersionCmd(version, commit))
	root.AddCommand(newConfigCmd())
	root.AddCommand(newNewCmd())

	return root
}

func newVersionCmd(version, commit string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the spm version",
		RunE: func(c *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(c.OutOrStdout(), "spm %s (commit %s)\n", version, commit)
			return err
		},
	}
}

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect spm configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the resolved configuration after all precedence layers",
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			return toml.NewEncoder(c.OutOrStdout()).Encode(cfg)
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print where spm looks for configuration and data",
		RunE: func(c *cobra.Command, _ []string) error {
			w := c.OutOrStdout()

			for _, row := range []struct{ label, path string }{
				{"user config", config.UserConfigFile()},
				{"project config", config.ProjectConfigFile()},
				{"database", config.DBPath()},
			} {
				if _, err := fmt.Fprintf(w, "%-15s %s\n", row.label+":", row.path); err != nil {
					return err
				}
			}

			return nil
		},
	})

	return configCmd
}
