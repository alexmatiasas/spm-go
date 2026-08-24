package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alexmatiasas/spm/internal/manifest"
)

// TemplatesDir locates the template catalog. During development it is
// resolved relative to the repository checkout; embedding replaces it
// for distribution builds.
var TemplatesDir = "templates"

// applyTemplates copies the base skeleton for the requested
// language x type into target, then layers the rigor fragments on
// top. The minimal level applies no fragments by definition: its
// guardrails are exactly what the base skeleton ships. Existing files
// are never overwritten: native tooling owns whatever it created.
func applyTemplates(target string, opts Options) error {
	base := filepath.Join(TemplatesDir, opts.Language, opts.ProjectType)
	if err := copyTree(base, target); err != nil {
		return err
	}

	if opts.RigorLevel == manifest.RigorMinimal {
		return nil
	}

	fragments := filepath.Join(TemplatesDir, "_fragments", opts.RigorLevel, opts.Language)

	return copyTree(fragments, target)
}

// copyTree copies every file under src into dst, creating directories
// as needed and preserving permissions. Existing files at the
// destination are left untouched.
func copyTree(src, dst string) error {
	info, statErr := os.Stat(src)
	switch {
	case errors.Is(statErr, fs.ErrNotExist):
		return fmt.Errorf("scaffold: template %s missing from catalog %s", src, TemplatesDir)
	case statErr != nil:
		return fmt.Errorf("scaffold: inspect %s: %w", src, statErr)
	case !info.IsDir():
		return fmt.Errorf("scaffold: %s is not a directory", src)
	}

	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("scaffold: walk %s: %w", path, walkErr)
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("scaffold: resolve %s: %w", path, err)
		}

		dstPath := filepath.Join(dst, rel)

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("scaffold: mkdir %s: %w", dstPath, err)
			}

			return nil
		}

		if _, err := os.Stat(dstPath); err == nil {
			return nil // native tooling owns existing files; skip
		}

		srcInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("scaffold: stat %s: %w", path, err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("scaffold: read %s: %w", path, err)
		}

		if err := os.WriteFile(dstPath, data, srcInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("scaffold: write %s: %w", dstPath, err)
		}

		return nil
	})
}
