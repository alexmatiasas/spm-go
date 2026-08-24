package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if len(cfg.Inventory.Roots) != 1 || cfg.Inventory.Roots[0] != "~/projects" {
		t.Errorf("Inventory.Roots = %v, want [~/projects]", cfg.Inventory.Roots)
	}

	wantExcluded := []string{"Icon", "Credenciales", "notes"}
	if !slices.Equal(cfg.Inventory.Exclude.Dirs, wantExcluded) {
		t.Errorf("Inventory.Exclude.Dirs = %v, want %v", cfg.Inventory.Exclude.Dirs, wantExcluded)
	}

	if cfg.Defaults.ProjectType != "python" {
		t.Errorf("Defaults.ProjectType = %q, want %q", cfg.Defaults.ProjectType, "python")
	}

	if cfg.Shell.AutoCD {
		t.Error("Shell.AutoCD = true, want false by default")
	}
}

func TestLoadFromMergesLayersOverDefaults(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user.toml")
	projectPath := filepath.Join(dir, ".spm", "config.toml")

	writeFile(t, userPath, `
[defaults]
project_type = "go"
`)
	writeFile(t, projectPath, `
[inventory]
roots = ["~/code"]

[shell]
auto_cd = true
`)

	cfg, err := LoadFrom(userPath, projectPath)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	wantRoots := []string{"~/code"}
	if !reflect.DeepEqual(cfg.Inventory.Roots, wantRoots) {
		t.Errorf("Roots = %v, want %v (project overrides default)", cfg.Inventory.Roots, wantRoots)
	}

	if !cfg.Shell.AutoCD {
		t.Error("AutoCD = false, want true (set by project layer)")
	}

	if cfg.Defaults.ProjectType != "go" {
		t.Errorf("ProjectType = %q, want %q (set by user layer)", cfg.Defaults.ProjectType, "go")
	}

	wantExcluded := []string{"Icon", "Credenciales", "notes"}
	if !slices.Equal(cfg.Inventory.Exclude.Dirs, wantExcluded) {
		t.Errorf("Exclude.Dirs = %v, want %v (default preserved when absent in files)", cfg.Inventory.Exclude.Dirs, wantExcluded)
	}
}

func TestLoadFromMissingFilesReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.toml")

	cfg, err := LoadFrom(missing, missing)
	if err != nil {
		t.Fatalf("LoadFrom with missing files: %v", err)
	}

	if !reflect.DeepEqual(cfg, Defaults()) {
		t.Errorf("cfg = %+v, want Defaults() %+v", cfg, Defaults())
	}
}

func TestLoadFromMalformedTOMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "broken.toml")

	writeFile(t, broken, "[inventory\nroots=")

	if _, err := LoadFrom(broken, filepath.Join(dir, "missing.toml")); err == nil {
		t.Fatal("LoadFrom with malformed TOML: got nil error, want error")
	}
}

func TestLoadWiresUserProjectAndEnvPaths(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user.toml")
	writeFile(t, userPath, `
[inventory]
roots = ["~/from-env"]
`)
	writeFile(t, filepath.Join(dir, ".spm", "config.toml"), `
[shell]
auto_cd = true
`)

	t.Setenv("SPM_CONFIG", userPath)
	t.Chdir(dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantRoots := []string{"~/from-env"}
	if !reflect.DeepEqual(cfg.Inventory.Roots, wantRoots) {
		t.Errorf("Roots = %v, want %v", cfg.Inventory.Roots, wantRoots)
	}
	if !cfg.Shell.AutoCD {
		t.Error("AutoCD = false, want true from project layer")
	}
}

func TestDefaultsIncludeProjectRoot(t *testing.T) {
	cfg := Defaults()

	if cfg.Defaults.ProjectRoot != "~/projects" {
		t.Errorf("Defaults.ProjectRoot = %q, want %q", cfg.Defaults.ProjectRoot, "~/projects")
	}
}
