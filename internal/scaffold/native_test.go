package scaffold

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alexmatiasas/spm/internal/manifest"
)

type recordedCall struct {
	dir  string
	name string
	args []string
}

type recorder struct {
	calls []recordedCall
}

func (r *recorder) run(_ context.Context, dir, name string, args ...string) error {
	r.calls = append(r.calls, recordedCall{dir: dir, name: name, args: args})

	return nil
}

func TestRunInvokesNativeToolInTargetDir(t *testing.T) {
	cases := []struct {
		label    string
		opts     Options
		binary   string
		args     []string
		packages string
	}{
		{
			label:  "python uv",
			opts:   Options{Name: "demo", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal, PackageManager: PMUV},
			binary: "uv",
			args:   []string{"init", "--name", "demo"},
		},
		{
			label:  "python conda",
			opts:   Options{Name: "demo", Language: LangPython, ProjectType: "ml-pipeline", RigorLevel: manifest.RigorStandard, PackageManager: PMConda},
			binary: "conda",
			args:   []string{"create", "--name", "demo", "-y"},
		},
		{
			label:  "go service",
			opts:   Options{Name: "demo", Language: LangGo, ProjectType: "service", RigorLevel: manifest.RigorStrict},
			binary: "go",
			args:   []string{"mod", "init", "demo"},
		},
		{
			label:  "rust cli",
			opts:   Options{Name: "demo", Language: LangRust, ProjectType: "cli", RigorLevel: manifest.RigorStandard},
			binary: "cargo",
			args:   []string{"init", "--name", "demo"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			useFixtureCatalog(t, map[string]string{
				"python/cli/README.md":                 "# cli",
				"python/ml-pipeline/README.md":         "# ml",
				"go/service/README.md":                 "# service",
				"rust/cli/README.md":                   "# rust cli",
				"_fragments/standard/python/lint.toml": "[lint]",
				"_fragments/standard/go/lint.toml":     "[lint]",
				"_fragments/standard/rust/lint.toml":   "[lint]",
				"_fragments/strict/go/mutation.toml":   "[mut]",
			})

			root := t.TempDir()
			rec := &recorder{}

			if err := Run(t.Context(), root, tc.opts, rec.run); err != nil {
				t.Fatalf("Run(%s): %v", tc.label, err)
			}

			if len(rec.calls) != 1 {
				t.Fatalf("got %d native calls, want exactly 1", len(rec.calls))
			}

			call := rec.calls[0]

			if want := filepath.Join(root, tc.opts.Name); call.dir != want {
				t.Errorf("call.dir = %q, want %q", call.dir, want)
			}

			if call.name != tc.binary {
				t.Errorf("call.name = %q, want %q", call.name, tc.binary)
			}

			if !slices.Equal(call.args, tc.args) {
				t.Errorf("call.args = %v, want %v", call.args, tc.args)
			}
		})
	}
}

func TestRunNativeFailureAbortsBeforeCompletion(t *testing.T) {
	root := t.TempDir()
	opts := validPython()

	fail := func(context.Context, string, string, ...string) error {
		return errors.New("tool crashed")
	}

	if err := Run(t.Context(), root, opts, fail); err == nil {
		t.Fatal("Run(failing native tool) = nil, want error")
	}
}
