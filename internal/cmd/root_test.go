package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := NewRootCmd("dev", "abc1234")
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)

	err := root.Execute()

	return out.String(), err
}

func TestVersionSubcommand(t *testing.T) {
	out, err := runRoot(t, "version")
	if err != nil {
		t.Fatalf("Execute version: %v", err)
	}

	if !strings.Contains(out, "spm dev") || !strings.Contains(out, "abc1234") {
		t.Errorf("version output = %q, want it to include binary and commit", out)
	}
}

func TestConfigShowPrintsResolvedConfig(t *testing.T) {
	dir := t.TempDir()
	user := filepath.Join(dir, "user.toml")
	if err := os.WriteFile(user, []byte(`
[defaults]
project_type = "go"
`), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	t.Setenv("SPM_CONFIG", user)
	t.Chdir(dir)

	out, err := runRoot(t, "config", "show")
	if err != nil {
		t.Fatalf("Execute config show: %v", err)
	}

	if !strings.Contains(out, `project_type = "go"`) {
		t.Errorf("config show output missing user override, got:\n%s", out)
	}

	if !strings.Contains(out, "~/projects") {
		t.Errorf("config show output missing default roots, got:\n%s", out)
	}
}

func TestConfigPathPrintsLocations(t *testing.T) {
	out, err := runRoot(t, "config", "path")
	if err != nil {
		t.Fatalf("Execute config path: %v", err)
	}

	for _, want := range []string{"config.toml", "spm.db"} {
		if !strings.Contains(out, want) {
			t.Errorf("config path output missing %q, got:\n%s", want, out)
		}
	}
}
