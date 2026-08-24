package scaffold

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alexmatiasas/spm/internal/manifest"
)

// useFixtureCatalog points TemplatesDir at a throwaway catalog for the
// duration of one test.
func useFixtureCatalog(t *testing.T, files map[string]string) {
	t.Helper()

	dir := t.TempDir()

	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("fixture mkdir %s: %v", rel, err)
		}

		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("fixture write %s: %v", rel, err)
		}
	}

	prev := TemplatesDir
	TemplatesDir = dir
	t.Cleanup(func() { TemplatesDir = prev })
}

func TestRunCopiesBaseTemplateIntoTarget(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/cli/README.md":       "# demo",
		"python/cli/src/__init__.py": "",
	})

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "README.md"))
	if err != nil || string(data) != "# demo" {
		t.Errorf("target README.md = %q (%v), want template content", data, err)
	}

	if _, err := os.Stat(filepath.Join(root, "demo", "src", "__init__.py")); err != nil {
		t.Errorf("target src/__init__.py missing: %v", err)
	}
}

func TestRunAppliesRigorFragmentOnTopOfBase(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/ml-pipeline/pyproject.toml":      "[project]",
		"_fragments/standard/python/ruff.toml":   "[lint]",
		"_fragments/strict/python/mutation.toml": "[mut]",
	})

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "ml-pipeline", RigorLevel: manifest.RigorStandard}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "demo", "ruff.toml")); err != nil {
		t.Errorf("standard fragment missing in target: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "demo", "mutation.toml")); !os.IsNotExist(err) {
		t.Error("strict fragment leaked into standard project")
	}
}

func TestStandardFragmentOverridesBaseFile(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/cli/lefthook.yml":                 "base: true",
		"_fragments/standard/python/lefthook.yml": "fragment: upgraded",
	})

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorStandard}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "lefthook.yml"))
	if err != nil {
		t.Fatalf("target lefthook.yml: %v", err)
	}

	if string(data) != "fragment: upgraded" {
		t.Errorf(
			"lefthook.yml = %q, want fragment content to override the base skeleton", data,
		)
	}
}

func TestMinimalKeepsBaseFileWhenFragmentWouldOverride(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/cli/lefthook.yml":                 "base: true",
		"_fragments/standard/python/lefthook.yml": "fragment: upgraded",
	})

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "lefthook.yml"))
	if err != nil {
		t.Fatalf("target lefthook.yml: %v", err)
	}

	if string(data) != "base: true" {
		t.Errorf("lefthook.yml = %q, minimal projects must keep the base content", data)
	}
}

func TestTemplatePlaceholdersAreRendered(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/cli/README.md": "# {{PROJECT_NAME}}\n\nCreated {{YEAR}}.\n",
	})

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal}

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "README.md"))
	if err != nil {
		t.Fatalf("target README.md: %v", err)
	}

	want := fmt.Sprintf("# demo\n\nCreated %d.\n", time.Now().UTC().Year())
	if string(data) != want {
		t.Errorf("README.md = %q, want rendered placeholders %q", data, want)
	}
}

func TestRunWritesRequestedLicense(t *testing.T) {
	cases := map[string]string{
		"mit":          "MIT License",
		"apache-2.0":   "Apache License",
		"bsd-3-clause": "BSD 3-Clause",
	}

	for key, want := range cases {
		t.Run(key, func(t *testing.T) {
			prev := TemplatesDir
			TemplatesDir = realCatalogDir(t)
			t.Cleanup(func() { TemplatesDir = prev })

			root := t.TempDir()
			opts := Options{
				Name:        "licensed",
				Language:    LangPython,
				ProjectType: "cli",
				RigorLevel:  manifest.RigorMinimal,
				License:     key,
			}

			if err := Run(t.Context(), root, opts); err != nil {
				t.Fatalf("Run: %v", err)
			}

			data, err := os.ReadFile(filepath.Join(root, "licensed", "LICENSE"))
			if err != nil {
				t.Fatalf("LICENSE: %v", err)
			}

			if !strings.Contains(string(data), want) {
				t.Errorf("LICENSE missing %q:\n%s", want, data)
			}

			if !strings.Contains(string(data), strconv.Itoa(time.Now().UTC().Year())) {
				t.Errorf("LICENSE missing rendered {{YEAR}}:\n%s", data)
			}
		})
	}
}

func TestValidateRejectsUnknownLicense(t *testing.T) {
	opts := Options{
		Name:        "ok",
		Language:    LangGo,
		ProjectType: "service",
		RigorLevel:  manifest.RigorMinimal,
		License:     "gpl-9000",
	}

	if err := Validate(opts); err == nil {
		t.Fatal("expected unknown license to fail validation")
	}
}

func TestRunNeverOverwritesNativeToolOutput(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/cli/README.md": "template version",
	})

	nativeWrites := func(_ context.Context, dir, _ string, _ ...string) error {
		return os.WriteFile(filepath.Join(dir, "README.md"), []byte("native version"), 0o644)
	}

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal}

	if err := Run(t.Context(), root, opts, nativeWrites); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "README.md"))
	if err != nil || string(data) != "native version" {
		t.Errorf("README.md = %q (%v), native output must win", data, err)
	}
}

// realCatalogDir locates the actual templates/ directory in the
// repository checkout, independent of the test working directory.
func realCatalogDir(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate templates_test.go on disk")
	}

	// thisFile = <repo>/internal/scaffold/templates_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	return filepath.Join(repoRoot, "templates")
}

// TestCatalogFragmentsNeverShadowNativeOutputs guards the overwrite
// semantics of the fragment pass: a fragment carrying a native-tool
// path (go.mod, pyproject.toml, ...) would clobber what uv, go mod or
// cargo generated.
func TestCatalogFragmentsNeverShadowNativeOutputs(t *testing.T) {
	reserved := map[string]bool{
		"pyproject.toml": true,
		"uv.lock":        true,
		"go.mod":         true,
		"go.sum":         true,
		"Cargo.toml":     true,
		"Cargo.lock":     true,
	}

	fragments := filepath.Join(realCatalogDir(t), "_fragments")

	err := filepath.WalkDir(fragments, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			t.Fatalf("walk %s: %v", path, walkErr)
		}

		if entry.IsDir() || !reserved[entry.Name()] {
			return nil
		}

		t.Errorf("%s is owned by native tooling; fragments must not carry it", path)

		return nil
	})
	if err != nil {
		t.Fatalf("walk fragments: %v", err)
	}
}

func TestRunFailsWhenBaseTemplateMissing(t *testing.T) {
	useFixtureCatalog(t, map[string]string{}) // catalog exists but is empty

	root := t.TempDir()
	opts := Options{Name: "demo", Language: LangRust, ProjectType: "cli", RigorLevel: manifest.RigorStandard}

	err := Run(t.Context(), root, opts)
	if err == nil {
		t.Fatal("Run(missing template) = nil, want error")
	}

	if !strings.Contains(err.Error(), "template") {
		t.Errorf("error %q should mention the template problem", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "demo")); !os.IsNotExist(statErr) {
		t.Error("failed template phase left target behind")
	}
}
