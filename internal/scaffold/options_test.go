package scaffold

import (
	"testing"

	"github.com/alexmatiasas/spm/internal/manifest"
)

func validPython() Options {
	return Options{
		Name:           "fraud-detector",
		Language:       LangPython,
		ProjectType:    "ml-pipeline",
		RigorLevel:     manifest.RigorStandard,
		PackageManager: PMUV,
		InitGit:        true,
	}
}

func TestTypesForKnownLanguages(t *testing.T) {
	cases := map[string][]string{
		LangPython: {"cli", "ml-pipeline"},
		LangGo:     {"service"},
		LangRust:   {"cli"},
	}

	for lang, want := range cases {
		got := TypesFor(lang)
		if len(got) != len(want) {
			t.Fatalf("TypesFor(%q) = %v, want %v", lang, got, want)
		}

		for i, typ := range want {
			if got[i] != typ {
				t.Errorf("TypesFor(%q)[%d] = %q, want %q", lang, i, got[i], typ)
			}
		}
	}
}

func TestValidateAcceptsEachLanguageTypePair(t *testing.T) {
	cases := []Options{
		{Name: "toolkit", Language: LangPython, ProjectType: "cli", RigorLevel: manifest.RigorMinimal},
		validPython(),
		{Name: "payments", Language: LangGo, ProjectType: "service", RigorLevel: manifest.RigorStrict},
		{Name: "notes-cli", Language: LangRust, ProjectType: "cli", RigorLevel: manifest.RigorStandard},
	}

	for _, opts := range cases {
		if err := Validate(opts); err != nil {
			t.Errorf("Validate(%+v) = %v, want nil", opts, err)
		}
	}
}

func TestValidateRejectsEmptyName(t *testing.T) {
	opts := validPython()
	opts.Name = ""

	if err := Validate(opts); err == nil {
		t.Error("Validate(empty name) = nil, want error")
	}
}

func TestValidateRejectsUnsafeNames(t *testing.T) {
	unsafe := []string{"..", ".", "a/b", `a\b`, ".hidden", "-lead", "sp ace"}

	for _, name := range unsafe {
		opts := validPython()
		opts.Name = name

		if err := Validate(opts); err == nil {
			t.Errorf("Validate(name %q) = nil, want error", name)
		}
	}
}

func TestValidateRejectsUnknownLanguage(t *testing.T) {
	opts := validPython()
	opts.Language = "haskell"

	if err := Validate(opts); err == nil {
		t.Error("Validate(unknown language) = nil, want error")
	}
}

func TestValidateRejectsTypeMismatchForLanguage(t *testing.T) {
	opts := validPython()
	opts.Language = LangGo
	opts.ProjectType = "ml-pipeline"

	if err := Validate(opts); err == nil {
		t.Error("Validate(go + ml-pipeline) = nil, want error")
	}
}

func TestValidateRejectsUnknownRigorLevel(t *testing.T) {
	opts := validPython()
	opts.RigorLevel = "extreme"

	if err := Validate(opts); err == nil {
		t.Error("Validate(unknown rigor) = nil, want error")
	}
}

func TestValidateRejectsPackageManagerOutsidePython(t *testing.T) {
	opts := validPython()
	opts.Language = LangGo
	opts.ProjectType = "service"

	if err := Validate(opts); err == nil {
		t.Error("Validate(go + package manager set) = nil, want error")
	}
}
