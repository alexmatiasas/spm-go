package scaffold

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/alexmatiasas/spm/internal/manifest"
)

func TestRunWritesManifestWithCreationMetadata(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"python/ml-pipeline/README.md":         "# ml",
		"_fragments/standard/python/lint.toml": "[lint]",
	})

	root := t.TempDir()
	opts := validPython()

	if err := Run(t.Context(), root, opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := manifest.Read(filepath.Join(root, opts.Name, ".spm", "manifest.yaml"))
	if err != nil {
		t.Fatalf("manifest.Read from scaffolded project: %v", err)
	}

	if got.Name != opts.Name || got.Language != opts.Language ||
		got.ProjectType != opts.ProjectType || got.RigorLevel != opts.RigorLevel {
		t.Errorf("manifest axes = %+v, want them mirrored from options %+v", got, opts)
	}

	if got.SchemaVersion != manifest.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, manifest.SchemaVersion)
	}

	if got.TemplateVersion != TemplateVersion {
		t.Errorf("template_version = %q, want %q", got.TemplateVersion, TemplateVersion)
	}

	if got.CreatedAt.IsZero() {
		t.Error("created_at must be stamped at creation")
	}
}

func gitCalls(calls []recordedCall) []recordedCall {
	out := make([]recordedCall, 0, len(calls))

	for _, c := range calls {
		if c.name == "git" {
			out = append(out, c)
		}
	}

	return out
}

func equalCalls(got, want []recordedCall) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i].dir != want[i].dir || got[i].name != want[i].name ||
			!slices.Equal(got[i].args, want[i].args) {
			return false
		}
	}

	return true
}

func TestRunRunsFullGitSequenceWhenEnabled(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"go/service/README.md":             "# service",
		"_fragments/standard/go/lint.toml": "[lint]",
		"_fragments/strict/go/mut.toml":    "[mut]",
	})

	root := t.TempDir()
	opts := Options{Name: "payments", Language: LangGo, ProjectType: "service", RigorLevel: manifest.RigorStrict, InitGit: true, InitialCommit: true}
	rec := &recorder{}

	if err := Run(t.Context(), root, opts, rec.run); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []recordedCall{
		{dir: filepath.Join(root, opts.Name), name: "git", args: []string{"init"}},
		{dir: filepath.Join(root, opts.Name), name: "git", args: []string{"add", "-A"}},
		{dir: filepath.Join(root, opts.Name), name: "git", args: []string{"commit", "-m", "Scaffold payments with spm"}},
	}

	got := gitCalls(rec.calls)

	if !equalCalls(got, want) {
		t.Errorf("git sequence:\n got %+v\nwant %+v", got, want)
	}
}

func TestRunSkipsGitWhenDisabled(t *testing.T) {
	useFixtureCatalog(t, map[string]string{
		"go/service/README.md":             "# service",
		"_fragments/standard/go/lint.toml": "[lint]",
	})

	root := t.TempDir()
	opts := Options{Name: "payments", Language: LangGo, ProjectType: "service", RigorLevel: manifest.RigorStandard, InitGit: false, InitialCommit: true}
	rec := &recorder{}

	if err := Run(t.Context(), root, opts, rec.run); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := gitCalls(rec.calls); len(got) != 0 {
		t.Errorf("git disabled but got calls %+v", got)
	}
}

func TestRunGitInitWithoutCommitStopsAtStagingFree(t *testing.T) {
	useFixtureCatalog(t, map[string]string{"go/service/README.md": "# service"})

	root := t.TempDir()
	opts := Options{Name: "payments", Language: LangGo, ProjectType: "service", RigorLevel: manifest.RigorMinimal, InitGit: true, InitialCommit: false}
	rec := &recorder{}

	if err := Run(t.Context(), root, opts, rec.run); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := gitCalls(rec.calls)

	if len(got) != 1 || !slices.Equal(got[0].args, []string{"init"}) {
		t.Errorf("--no-commit must stop after git init, got %+v", got)
	}
}
