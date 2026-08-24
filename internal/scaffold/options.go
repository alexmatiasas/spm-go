// Package scaffold creates new projects from the language x type
// catalog, layered with rigor policy fragments, delegating environment
// setup to each language's native tooling.
package scaffold

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	License        string // optional: mit | apache-2.0 | bsd-3-clause
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

	// Licenses lists the SPDX-style keys spm can stamp into a new
	// project. Order is the recommendation order shown in the wizard:
	// MIT for portfolio/public work, Apache-2.0 when patent protection
	// matters, BSD-3-Clause as the permissive alternative.
	Licenses = []string{"mit", "apache-2.0", "bsd-3-clause"}

	licenses = func() map[string]bool {
		m := make(map[string]bool, len(Licenses))
		for _, l := range Licenses {
			m[l] = true
		}

		return m
	}()

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

	if opts.License != "" && !licenses[opts.License] {
		return fmt.Errorf("scaffold: unknown license %q (want one of %s)",
			opts.License, strings.Join(Licenses, ", "))
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
