package scaffold

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// CommandRunner executes an external command. It exists so tests can
// replace native tool invocations with fakes.
type CommandRunner func(ctx context.Context, name string, args ...string) error

// ErrDestinationOccupied reports that the target directory already
// contains user work and scaffold refuses to touch it.
var ErrDestinationOccupied = errors.New("scaffold: destination exists and is not empty")

// primaryBinary names the native tool that owns environment setup for
// the requested stack. Full argument wiring arrives with native tool
// integration; lifecycle tests only pin which binary would run.
func primaryBinary(opts Options) string {
	switch opts.Language {
	case LangPython:
		if opts.PackageManager == PMConda {
			return PMConda
		}

		return PMUV
	default:
		return opts.Language
	}
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

	// Native step only runs when a runner is injected; the production
	// exec runner arrives with native tool integration.
	for _, r := range run {
		if r != nil {
			if err := r(ctx, primaryBinary(opts)); err != nil {
				return err
			}

			break
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	completed = true

	return nil
}
