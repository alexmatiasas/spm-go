// Package scaffold creates new projects from the language x type
// catalog, layered with rigor policy fragments, delegating environment
// setup to each language's native tooling.
package scaffold

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/alexmatiasas/spm/internal/manifest"
)

// Supported languages and package managers.
const (
	LangPython = "python"
	LangGo     = "go"
	LangRust   = "rust"

	PMUV    = "uv"
	PMConda = "conda"
)

// Options describes everything spm needs to scaffold one project.
type Options struct {
	Name           string
	Language       string
	ProjectType    string
	RigorLevel     string
	PackageManager string // python only: uv | conda
	InitGit        bool
	InitialCommit  bool
}

var (
	languages   = map[string]bool{LangPython: true, LangGo: true, LangRust: true}
	rigorLevels = map[string]bool{
		manifest.RigorMinimal:  true,
		manifest.RigorStandard: true,
		manifest.RigorStrict:   true,
	}
	packageManagers = map[string]bool{PMUV: true, PMConda: true}

	// safeName allows a single path-safe token: no separators, no
	// leading dot or dash, no whitespace.
	safeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
)

// TypesFor returns the valid project types for a language.
func TypesFor(language string) []string {
	switch language {
	case LangPython:
		return []string{"cli", "ml-pipeline"}
	case LangGo:
		return []string{"service"}
	case LangRust:
		return []string{"cli"}
	default:
		return nil
	}
}

// Validate reports whether opts is a well-formed scaffold request.
func Validate(opts Options) error {
	if opts.Name == "" {
		return errors.New("scaffold: project name is required")
	}

	if !safeName.MatchString(opts.Name) {
		return fmt.Errorf("scaffold: unsafe project name %q", opts.Name)
	}

	if !languages[opts.Language] {
		return fmt.Errorf("scaffold: unknown language %q", opts.Language)
	}

	if !validTypeFor(opts.Language, opts.ProjectType) {
		return fmt.Errorf("scaffold: type %q invalid for %s", opts.ProjectType, opts.Language)
	}

	if !rigorLevels[opts.RigorLevel] {
		return fmt.Errorf("scaffold: unknown rigor level %q", opts.RigorLevel)
	}

	switch {
	case opts.Language != LangPython && opts.PackageManager != "":
		return fmt.Errorf(
			"scaffold: package manager %q only applies to python", opts.PackageManager,
		)
	case opts.PackageManager != "" && !packageManagers[opts.PackageManager]:
		return fmt.Errorf(
			"scaffold: unknown package manager %q (want uv or conda)", opts.PackageManager,
		)
	}

	return nil
}

func validTypeFor(language, projectType string) bool {
	for _, t := range TypesFor(language) {
		if t == projectType {
			return true
		}
	}

	return false
}
