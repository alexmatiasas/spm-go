package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newNewCmd builds `spm new`, the scaffolding entry point.
func newNewCmd() *cobra.Command {
	newCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Scaffold a new project with guardrails by rigor level",
		Args:  cobra.ExactArgs(1),
		RunE: func(*cobra.Command, []string) error {
			return errors.New("not implemented") // stub: wiring lands in the green commit
		},
	}

	return newCmd
}
