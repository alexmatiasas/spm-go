// Package manifest defines the .spm/manifest.yaml schema shared by
// scaffolding (writer) and inventory (reader). The manifest is the
// single source of truth connecting both subsystems.
package manifest

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the manifest schema version written by this build of spm.
const SchemaVersion = 1

// Supported rigor levels.
const (
	RigorMinimal  = "minimal"
	RigorStandard = "standard"
	RigorStrict   = "strict"
)

var (
	languages   = map[string]bool{"python": true, "go": true, "rust": true}
	rigorLevels = map[string]bool{RigorMinimal: true, RigorStandard: true, RigorStrict: true}
)

// Manifest describes a project scaffolded by spm.
type Manifest struct {
	SchemaVersion   int       `yaml:"schema_version"`
	Name            string    `yaml:"name"`
	Language        string    `yaml:"language"`
	ProjectType     string    `yaml:"project_type"`
	RigorLevel      string    `yaml:"rigor_level"`
	TemplateVersion string    `yaml:"template_version"`
	CreatedAt       time.Time `yaml:"created_at"`
}

// Validate reports whether m satisfies the schema invariants.
func Validate(m Manifest) error {
	if m.Name == "" {
		return errors.New("manifest: name is required")
	}

	if !languages[m.Language] {
		return fmt.Errorf("manifest: unknown language %q", m.Language)
	}

	if m.ProjectType == "" {
		return errors.New("manifest: project type is required")
	}

	if !rigorLevels[m.RigorLevel] {
		return fmt.Errorf("manifest: unknown rigor level %q", m.RigorLevel)
	}

	return nil
}

// Write validates m and serializes it to path as YAML.
func Write(path string, m Manifest) error {
	if err := Validate(m); err != nil {
		return err
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("manifest: encode %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("manifest: write %s: %w", path, err)
	}

	return nil
}

// Read parses and validates the manifest at path.
func Read(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: read %s: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest: parse %s: %w", path, err)
	}

	if err := Validate(m); err != nil {
		return Manifest{}, err
	}

	return m, nil
}
