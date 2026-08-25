package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexmatiasas/spm/internal/config"
	"github.com/alexmatiasas/spm/internal/manifest"
	"github.com/alexmatiasas/spm/internal/scaffold"
	"github.com/spf13/cobra"
)

// execCommand shells out to native tooling inside dir, streaming tool
// output to stderr so spm's stdout stays clean for data.
func execCommand(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// execRunner is the runner scaffold.Run delegates native commands
// through. Production always uses execCommand; wiring tests swap it
// for a recording fake so they stay hermetic without uv/git/cargo.
var execRunner scaffold.CommandRunner = execCommand

// expandTilde resolves a leading ~ to the user home directory.
func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, string(filepath.Separator)+"~") &&
		!strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	switch {
	case path == "~":
		return home
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(home, path[2:])
	default:
		return path
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

// printNextSteps advises what to do with the freshly scaffolded tree.
// spm leaves every file unstaged behind an empty initial commit, so
// the user curates their first real commit after adjusting hooks,
// linters and configs to taste.
func printNextSteps(out io.Writer, opts scaffold.Options) {
	hookTool := "lefthook"
	if opts.Language == scaffold.LangPython {
		hookTool = "prek" // or pre-commit: the hook format is shared
	}

	// Advice is best-effort: a closed stdout must not fail an
	// otherwise successful scaffold.
	_, _ = fmt.Fprintf(out, `
Scaffolded %s. Next steps:
  1. Review the generated files — everything is unstaged. Adjust hooks,
     linters and configs before your first real commit.
  2. Install the git hooks once:  %s install
  3. When ready:                  git add -A && git commit -m "Set up project"
  4. Optional remote:             gh repo create <owner>/%s --source . --push
`, opts.Name, hookTool, opts.Name)
}

// newNewCmd builds `spm new`, the scaffolding entry point.
func newNewCmd() *cobra.Command {
	var (
		lang     string
		typ      string
		rigor    string
		pm       string
		license  string
		root     string
		noGit    bool
		noCommit bool
	)

	newCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Scaffold a new project with guardrails by rigor level",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			opts := scaffold.Options{
				Name:           args[0],
				Language:       firstNonEmpty(lang, cfg.Defaults.ProjectType),
				ProjectType:    typ,
				RigorLevel:     firstNonEmpty(rigor, manifest.RigorStandard),
				PackageManager: pm,
				License:        license,
				InitGit:        !noGit,
				InitialCommit:  !noCommit,
			}

			base := expandTilde(firstNonEmpty(root, cfg.Defaults.ProjectRoot, "."))

			if err := scaffold.Run(cmd.Context(), base, opts, execRunner); err != nil {
				return err
			}

			printNextSteps(cmd.OutOrStdout(), opts)

			return nil
		},
	}

	flags := newCmd.Flags()
	flags.StringVar(&lang, "lang", "", "project language: python, go or rust")
	flags.StringVar(&typ, "type", "",
		"project type valid for the language: cli, service or ml-pipeline")
	flags.StringVar(&rigor, "rigor", "", "guardrails level: minimal, standard or strict")
	flags.StringVar(&pm, "pm", "", "python package manager: uv or conda")
	flags.StringVar(&license, "license", "",
		"stamp a LICENSE file: "+strings.Join(scaffold.Licenses, ", "))
	flags.StringVar(&root, "root", "", "directory where the project is created")
	flags.BoolVar(&noGit, "no-git", false, "skip git initialization")
	flags.BoolVar(&noCommit, "no-commit", false, "stop after git init without committing")

	return newCmd
}
