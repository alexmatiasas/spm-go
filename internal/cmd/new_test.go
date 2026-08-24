package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmatiasas/spm/internal/manifest"
	"github.com/alexmatiasas/spm/internal/scaffold"
)

// useCatalog points the template catalog at a throwaway fixture for
// one test, keeping command tests hermetic.
func useCatalog(t *testing.T) {
	t.Helper()

	dir := t.TempDir()

	files := map[string]string{
		"go/service/README.md":         "# service",
		"python/cli/README.md":         "# cli",
		"python/ml-pipeline/README.md": "# ml",
	}

	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("catalog mkdir: %v", err)
		}

		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("catalog write: %v", err)
		}
	}

	prev := scaffold.TemplatesDir
	scaffold.TemplatesDir = dir
	t.Cleanup(func() { scaffold.TemplatesDir = prev })
}

func runNew(t *testing.T, args ...string) error {
	t.Helper()

	root := NewRootCmd("dev", "test")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"new"}, args...))

	return root.ExecuteContext(t.Context())
}

func TestNewRequiresProjectName(t *testing.T) {
	useCatalog(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := runNew(t); err == nil {
		t.Fatal("spm new without name must fail")
	}
}

func TestNewRejectsUnknownLanguage(t *testing.T) {
	useCatalog(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := runNew(t, "demo", "--lang", "haskell", "--type", "cli", "--rigor", "minimal", "--root", dir)
	if err == nil {
		t.Fatal("spm new with unknown language must fail")
	}

	if !strings.Contains(err.Error(), `unknown language "haskell"`) {
		t.Errorf("error %q should name the bad language", err)
	}
}

func TestNewScaffoldsIntoExplicitRoot(t *testing.T) {
	useCatalog(t)
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := runNew(t, "demo", "--lang", "go", "--type", "service", "--rigor", "minimal", "--root", root, "--no-git"); err != nil {
		t.Fatalf("spm new: %v", err)
	}

	got, err := manifest.Read(filepath.Join(root, "demo", ".spm", "manifest.yaml"))
	if err != nil {
		t.Fatalf("manifest missing from scaffolded project: %v", err)
	}

	if got.Language != "go" || got.RigorLevel != manifest.RigorMinimal {
		t.Errorf("manifest axes wrong: %+v", got)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "README.md"))
	if err != nil || string(data) != "# service" {
		t.Errorf("template file not copied: %q (%v)", data, err)
	}
}

func TestNewDefaultsLanguageFromConfigAndRootFromCwd(t *testing.T) {
	useCatalog(t)

	xdg := t.TempDir()
	if err := os.MkdirAll(filepath.Join(xdg, "spm"), 0o755); err != nil {
		t.Fatalf("xdg mkdir: %v", err)
	}

	userCfg := filepath.Join(xdg, "spm", "config.toml")
	// project_root = "" opts out of the shipped default: spm new then
	// falls back to the current directory.
	if err := os.WriteFile(userCfg, []byte("[defaults]\nproject_type = \"go\"\nproject_root = \"\"\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdg)

	cwd := t.TempDir()
	t.Chdir(cwd)

	if err := runNew(t, "demo", "--type", "service", "--rigor", "minimal"); err != nil {
		t.Fatalf("spm new without --lang/--root: %v", err)
	}

	got, err := manifest.Read(filepath.Join(cwd, "demo", ".spm", "manifest.yaml"))
	if err != nil {
		t.Fatalf("project not created in cwd: %v", err)
	}

	if got.Language != "go" {
		t.Errorf("language = %q, want config default go", got.Language)
	}
}
