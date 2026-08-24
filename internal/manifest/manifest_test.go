package manifest

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"pgregory.net/rapid"
)

func validManifest() Manifest {
	return Manifest{
		SchemaVersion:   SchemaVersion,
		Name:            "demo",
		Language:        "python",
		ProjectType:     "ml-pipeline",
		RigorLevel:      RigorStandard,
		TemplateVersion: "0.1.0",
		CreatedAt:       time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC),
	}
}

func TestValidateAcceptsWellFormedManifest(t *testing.T) {
	if err := Validate(validManifest()); err != nil {
		t.Errorf("Validate(well-formed) = %v, want nil", err)
	}
}

func TestValidateRejectsUnknownLanguage(t *testing.T) {
	m := validManifest()
	m.Language = "haskell"

	if err := Validate(m); err == nil {
		t.Error("Validate(unknown language) = nil, want error")
	}
}

func TestValidateRejectsUnknownRigorLevel(t *testing.T) {
	m := validManifest()
	m.RigorLevel = "extreme"

	if err := Validate(m); err == nil {
		t.Error("Validate(unknown rigor level) = nil, want error")
	}
}

func TestValidateRequiresName(t *testing.T) {
	m := validManifest()
	m.Name = ""

	if err := Validate(m); err == nil {
		t.Error("Validate(empty name) = nil, want error")
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	want := validManifest()

	if err := Write(path, want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func TestReadMissingFileFails(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Error("Read(missing file) = nil error, want failure")
	}
}

// Property: any well-formed manifest survives a Write/Read cycle unchanged.
func TestManifestRoundTripProperty(t *testing.T) {
	manifestGen := rapid.Custom(func(t *rapid.T) Manifest {
		return Manifest{
			SchemaVersion:   SchemaVersion,
			Name:            rapid.StringMatching(`[a-z][a-z0-9-]{0,30}`).Draw(t, "name"),
			Language:        rapid.SampledFrom([]string{"python", "go", "rust"}).Draw(t, "language"),
			ProjectType:     rapid.StringMatching(`[a-z][a-z0-9-]{0,30}`).Draw(t, "project_type"),
			RigorLevel:      rapid.SampledFrom([]string{RigorMinimal, RigorStandard, RigorStrict}).Draw(t, "rigor"),
			TemplateVersion: rapid.StringMatching(`[0-9]+\.[0-9]+\.[0-9]+`).Draw(t, "template_version"),
			CreatedAt:       time.Unix(rapid.Int64Range(0, 4_102_444_800).Draw(t, "created_at"), 0).UTC(),
		}
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")

	rapid.Check(t, func(t *rapid.T) {
		want := manifestGen.Draw(t, "manifest")

		if err := Write(path, want); err != nil {
			t.Fatalf("Write: %v", err)
		}

		got, err := Read(path)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("round trip changed the manifest:\n got %+v\nwant %+v", got, want)
		}
	})
}
