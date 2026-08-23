// Package manifest defines the .spm/manifest.yaml schema shared by
// scaffolding (writer) and inventory (reader). The manifest is the
// single source of truth connecting both subsystems.
package manifest

import (
	"errors"
	"time"
)

// SchemaVersion is the manifest schema version written by this build of spm.
const SchemaVersion = 1

// Supported rigor levels.
const (
	RigorMinimal  = "minimal"
	RigorStandard = "standard"
	RigorStrict   = "strict"
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
func Validate(_ Manifest) error {
	return errors.New("not implemented")
}

// Write serializes m to path as YAML.
func Write(_ string, _ Manifest) error {
	return errors.New("not implemented")
}

// Read parses the manifest at path.
func Read(_ string) (Manifest, error) {
	return Manifest{}, errors.New("not implemented")
}
