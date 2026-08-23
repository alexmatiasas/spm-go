// Package scaffold creates new projects from the language x type
// catalog, layered with rigor policy fragments, delegating environment
// setup to each language's native tooling.
package scaffold

import "errors"

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
func Validate(_ Options) error {
	return errors.New("not implemented")
}
