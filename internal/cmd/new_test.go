package cmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/alexmatiasas/spm/internal/manifest"
	"github.com/alexmatiasas/spm/internal/scaffold"
)

// fakeToolRunner records native commands instead of executing them,
// keeping command wiring tests independent of installed toolchains.
type fakeToolRunner struct {
	calls [][]string
}

func (f *fakeToolRunner) record(_ context.Context, _, name string, args ...string) error {
	f.calls = append(f.calls, append([]string{name}, args...))

	return nil
}

func (f *fakeToolRunner) ran(name string, args ...string) bool {
	want := append([]string{name}, args...)

	return slices.ContainsFunc(f.calls, func(c []string) bool {
		return slices.Equal(c, want)
	})
}

// useFakeTools swaps the exec runner for a recording fake for one
// test and returns it for assertions.
func useFakeTools(t *testing.T) *fakeToolRunner {
	t.Helper()

	fake := &fakeToolRunner{}
	prev := execRunner
	execRunner = fake.record
	t.Cleanup(func() { execRunner = prev })

	return fake
}

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

func TestNewPrintsNextStepsAdvice(t *testing.T) {
	cases := map[string]struct {
		args     []string
		hookHint string
	}{
		"go":     {args: []string{"--lang", "go", "--type", "service"}, hookHint: "lefthook install"},
		"python": {args: []string{"--lang", "python", "--type", "cli"}, hookHint: "prek install"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			useCatalog(t)
			useFakeTools(t)

			out := &bytes.Buffer{}
			rootCmd := NewRootCmd("dev", "test")
			rootCmd.SetOut(out)
			rootCmd.SetErr(&bytes.Buffer{})
			rootCmd.SetArgs(append([]string{
				"new", "demo", "--rigor", manifest.RigorMinimal,
				"--no-git", "--root", t.TempDir(),
			}, tc.args...))

			if err := rootCmd.ExecuteContext(t.Context()); err != nil {
				t.Fatalf("spm new: %v", err)
			}

			for _, want := range []string{"unstaged", tc.hookHint, "git add"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("advice missing %q:\n%s", want, out)
				}
			}
		})
	}
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
	fake := useFakeTools(t)
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := runNew(t, "demo", "--lang", "go", "--type", "service", "--rigor", "minimal", "--root", root, "--no-git"); err != nil {
		t.Fatalf("spm new: %v", err)
	}

	if !fake.ran("go", "mod", "init", "demo") {
		t.Errorf("native setup not delegated to go mod init, calls: %v", fake.calls)
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

	fake := useFakeTools(t)

	cwd := t.TempDir()
	t.Chdir(cwd)

	if err := runNew(t, "demo", "--type", "service", "--rigor", "minimal"); err != nil {
		t.Fatalf("spm new without --lang/--root: %v", err)
	}

	if !fake.ran("git", "init") || !fake.ran("git", "commit", "--allow-empty", "-m", "Initial commit") {
		t.Errorf("git sequence wrong, calls: %v", fake.calls)
	}

	got, err := manifest.Read(filepath.Join(cwd, "demo", ".spm", "manifest.yaml"))
	if err != nil {
		t.Fatalf("project not created in cwd: %v", err)
	}

	if got.Language != "go" {
		t.Errorf("language = %q, want config default go", got.Language)
	}
}

// TestNewRealToolchainSmoke exercises the real exec path against
// installed native tools — the coverage the wiring tests above give
// up by faking the runner. Skips per language when a tool is absent;
// CI installs both tools so the smoke runs fully there.
func TestNewRealToolchainSmoke(t *testing.T) {
	cases := []struct{ lang, typ, bin string }{
		{"go", "service", "go"},
		{"python", "cli", "uv"},
	}

	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			if _, err := exec.LookPath(tc.bin); err != nil {
				t.Skipf("%s not installed", tc.bin)
			}

			useCatalog(t)

			prev := execRunner
			execRunner = execCommand
			t.Cleanup(func() { execRunner = prev })

			if err := runNew(t, "demo",
				"--lang", tc.lang, "--type", tc.typ,
				"--rigor", manifest.RigorMinimal,
				"--root", t.TempDir(), "--no-git",
			); err != nil {
				t.Fatalf("spm new with real toolchain: %v", err)
			}
		})
	}
}
