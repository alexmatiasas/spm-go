package scaffold

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// failingRunner simulates a native tool crashing mid-scaffold.
func failingRunner(context.Context, string, string, ...string) error {
	return errors.New("boom")
}

func TestRunCreatesTargetDirectory(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/ml-pipeline/README.md":         "# demo",
		"_fragments/standard/python/lint.toml": "[lint]",
	})

	root := t.TempDir()
	opts := validPython()

	err := Run(t.Context(), root, opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	target := filepath.Join(root, opts.Name)
	info, statErr := os.Stat(target)
	if statErr != nil {
		t.Fatalf("target dir missing: %v", statErr)
	}

	if !info.IsDir() {
		t.Errorf("target %s is not a directory", target)
	}
}

func TestRunRejectsOccupiedDestination(t *testing.T) {
	root := t.TempDir()
	opts := validPython()

	occupied := filepath.Join(root, opts.Name)
	if err := os.MkdirAll(filepath.Join(occupied, "user-data"), 0o755); err != nil {
		t.Fatalf("seed occupied dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(occupied, "keep.txt"), []byte("mine"), 0o644); err != nil {
		t.Fatalf("seed occupied file: %v", err)
	}

	if err := Run(t.Context(), root, opts); err == nil {
		t.Fatal("Run(occupied destination) = nil, want error")
	}

	data, readErr := os.ReadFile(filepath.Join(occupied, "keep.txt"))
	if readErr != nil || string(data) != "mine" {
		t.Errorf("user data damaged after rejected run: %q %v", data, readErr)
	}
}

func TestRunAllowsEmptyExistingDestination(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/ml-pipeline/README.md":         "# demo",
		"_fragments/standard/python/lint.toml": "[lint]",
	})

	root := t.TempDir()
	opts := validPython()

	if err := os.Mkdir(filepath.Join(root, opts.Name), 0o755); err != nil {
		t.Fatalf("seed empty dir: %v", err)
	}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run(empty existing dir): %v", err)
	}
}

func TestRunRollsBackWhenNativeToolFails(t *testing.T) {
	root := t.TempDir()
	opts := validPython()

	err := Run(t.Context(), root, opts, failingRunner)
	if err == nil {
		t.Fatal("Run(failing runner) = nil, want error")
	}

	if _, statErr := os.Stat(filepath.Join(root, opts.Name)); !os.IsNotExist(statErr) {
		t.Errorf("rollback left target behind: %v", statErr)
	}
}

func TestRunRespectsCanceledContext(t *testing.T) {
	root := t.TempDir()
	opts := validPython()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := Run(ctx, root, opts)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run(canceled ctx) = %v, want context.Canceled", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, opts.Name)); !os.IsNotExist(statErr) {
		t.Error("canceled run left target behind")
	}
}
