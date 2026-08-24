package scaffold

import (
	"errors"
	"os"
)

// TemplatesDir locates the template catalog. During development it is
// resolved relative to the repository checkout; embedding replaces it
// for distribution builds.
var TemplatesDir = "templates"

// applyTemplates copies the base skeleton for the requested
// language x type into target, then layers the rigor fragments on top.
// Existing files are never overwritten: native tooling owns whatever
// it already created.
func applyTemplates(_ string, _ Options) error {
	if _, statErr := os.Stat(TemplatesDir); statErr == nil {
		return errors.New("scaffold: templates not implemented") // stub
	}

	return nil // provisional: catalog absent, nothing to apply yet
}
