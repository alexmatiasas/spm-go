package scaffold

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/alexmatiasas/spm/internal/manifest"
)

// CommandRunner executes an external command inside dir. It exists so
// tests can replace native tool invocations with fakes.
type CommandRunner func(ctx context.Context, dir, name string, args ...string) error

// ErrDestinationOccupied reports that the target directory already
// contains user work and scaffold refuses to touch it.
var ErrDestinationOccupied = errors.New("scaffold: destination exists and is not empty")

// nativeCommand maps the requested stack to its native tooling
// invocation, run inside the project directory. conda creates the
// environment; templates add any project files it does not cover.
func nativeCommand(opts Options) (string, []string) {
	const (
		nameFlag = "--name"
		initWord = "init"
	)

	switch opts.Language {
	case LangPython:
		if opts.PackageManager == PMConda {
			return PMConda, []string{"create", nameFlag, opts.Name, "-y"}
		}

		return PMUV, []string{initWord, nameFlag, opts.Name}
	case LangGo:
		return "go", []string{"mod", initWord, opts.Name}
	case LangRust:
		return "cargo", []string{initWord, nameFlag, opts.Name}
	default:
		return opts.Language, nil
	}
}

// writeProjectManifest records the creation decision inside the new
// project, under .spm/manifest.yaml.
func writeProjectManifest(target string, opts Options) error {
	m := manifest.Manifest{
		SchemaVersion:   manifest.SchemaVersion,
		Name:            opts.Name,
		Language:        opts.Language,
		ProjectType:     opts.ProjectType,
		RigorLevel:      opts.RigorLevel,
		TemplateVersion: TemplateVersion,
		CreatedAt:       time.Now().UTC(),
	}

	path := filepath.Join(target, ".spm", "manifest.yaml")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("scaffold: create %s: %w", filepath.Dir(path), err)
	}

	if err := manifest.Write(path, m); err != nil {
		return fmt.Errorf("scaffold: %w", err)
	}

	return nil
}

// gitSteps returns the git invocation sequence for the requested
// options. --no-commit stops after init: staging and the first real
// commit stay in the user's hands. The initial commit is empty on
// purpose — like git flow's root commit — so scaffold files stay
// unstaged and the user curates what lands in their first real
// commit; an existing commit also unblocks pushing a remote.
func gitSteps(opts Options) [][]string {
	if !opts.InitGit {
		return nil
	}

	steps := [][]string{{"init"}}

	if opts.InitialCommit {
		steps = append(steps,
			[]string{"commit", "--allow-empty", "-m", "Initial commit"},
		)
	}

	return steps
}

// Run scaffolds a new project at <root>/<opts.Name>. The destination
// must not exist or be an empty directory; anything else is refused.
// Every step after the directory is created is transactional: on error
// or cancellation the created tree is removed entirely.
func Run(ctx context.Context, root string, opts Options, run ...CommandRunner) error {
	if err := Validate(opts); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	target := filepath.Join(root, opts.Name)

	info, statErr := os.Stat(target)
	switch {
	case statErr == nil && !info.IsDir():
		return fmt.Errorf("scaffold: %s is not a directory", target)
	case statErr == nil:
		entries, readErr := os.ReadDir(target)
		if readErr != nil {
			return fmt.Errorf("scaffold: inspect %s: %w", target, readErr)
		}

		if len(entries) > 0 {
			return fmt.Errorf("%w: %s", ErrDestinationOccupied, target)
		}
	case !errors.Is(statErr, fs.ErrNotExist):
		return fmt.Errorf("scaffold: stat %s: %w", target, statErr)
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("scaffold: create %s: %w", target, err)
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(target)
		}
	}()

	// Native and git steps only run when a runner is injected; the
	// production exec runner arrives with CLI wiring.
	var runner CommandRunner
	for _, r := range run {
		if r != nil {
			runner = r
			break
		}
	}

	if runner != nil {
		binary, args := nativeCommand(opts)
		if err := runner(ctx, target, binary, args...); err != nil {
			return err
		}
	}

	if err := applyTemplates(target, opts); err != nil {
		return err
	}

	if err := applyLicense(target, opts); err != nil {
		return err
	}

	if err := writeProjectManifest(target, opts); err != nil {
		return err
	}

	for _, args := range gitSteps(opts) {
		if runner == nil {
			break // provisional: production runner lands with CLI wiring
		}

		if err := runner(ctx, target, "git", args...); err != nil {
			return err
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	completed = true

	return nil
}
