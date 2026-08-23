// Package cmd ...
package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the spm command tree. version and commit are
// reported by the version subcommand.
func NewRootCmd(_, _ string) *cobra.Command {
	root := &cobra.Command{
		Use:   "spm",
		Short: "Smart Project Manager",
		Long:  "spm scaffolds projects with the right guardrails and keeps an inventory of everything you build.",
	}

	return root
}
