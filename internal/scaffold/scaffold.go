package scaffold

import (
	"context"
	"errors"
)

// CommandRunner executes an external command. It exists so tests can
// replace native tool invocations with fakes.
type CommandRunner func(ctx context.Context, name string, args ...string) error

// Run scaffolds a new project at <root>/<opts.Name>.
func Run(_ context.Context, _ string, _ Options, _ ...CommandRunner) error {
	return errors.New("not implemented")
}
