package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alexmatiasas/spm/internal/manifest"
)

// TemplatesDir locates the template catalog. During development it is
// resolved relative to the repository checkout; embedding replaces it
// for distribution builds.
var TemplatesDir = "templates"

// TemplateVersion identifies which catalog revision produced a
// project; inventory compares it against the current value to detect
// drift.
const TemplateVersion = "0.1.0"

// applyTemplates copies the base skeleton for the requested
// language x type into target, then layers rigor fragments on top.
// Rigor is cumulative: a level applies every fragment set from
// standard up to its own, in ascending order, so strict gets
// everything standard ships plus its own upgrades. The minimal level
// applies no fragments by definition: its guardrails are exactly what
// the base skeleton ships. The base pass never overwrites existing
// files (native tooling owns whatever it created); the fragment pass
// does, so a rigor level can upgrade a file an earlier layer shipped.
func applyTemplates(target string, opts Options) error {
	base := filepath.Join(TemplatesDir, opts.Language, opts.ProjectType)
	if err := copyTree(base, target, false, opts); err != nil {
		return err
	}

	if opts.RigorLevel == manifest.RigorMinimal {
		return nil
	}

	for _, level := range []string{manifest.RigorStandard, manifest.RigorStrict} {
		fragments := filepath.Join(TemplatesDir, "_fragments", level, opts.Language)
		if level == manifest.RigorStrict && opts.RigorLevel == manifest.RigorStandard {
			break
		}

		if err := copyTree(fragments, target, true, opts); err != nil {
			return err
		}
	}

	return nil
}

// applyLicense stamps the requested license text into target. License
// texts live in the catalog under _licenses/<key>/LICENSE and carry
// the same {{YEAR}} placeholder as every other template file, so an
// empty license leaves the project untouched.
func applyLicense(target string, opts Options) error {
	if opts.License == "" {
		return nil
	}

	src := filepath.Join(TemplatesDir, "_licenses", opts.License, "LICENSE")
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf(
			"scaffold: license %s missing from catalog %s: %w",
			opts.License,
			TemplatesDir,
			err,
		)
	}

	dst := filepath.Join(target, "LICENSE")
	if err := os.WriteFile(dst, renderTemplate(data, opts), 0o644); err != nil {
		return fmt.Errorf("scaffold: write %s: %w", dst, err)
	}

	return nil
}

// renderTemplate substitutes the placeholder tokens spm owns. Tokens
// are plain text so files without them pass through byte-identical.
func renderTemplate(data []byte, opts Options) []byte {
	out := bytes.ReplaceAll(data, []byte("{{PROJECT_NAME}}"), []byte(opts.Name))
	out = bytes.ReplaceAll(out, []byte("{{YEAR}}"), []byte(strconv.Itoa(time.Now().UTC().Year())))

	return out
}

// copyTree copies every file under src into dst, creating directories
// as needed and preserving permissions. With overwrite=false existing
// files at the destination are left untouched; with overwrite=true
// they are replaced, which is how rigor fragments upgrade base files.
func copyTree(src, dst string, overwrite bool, opts Options) error {
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

		if _, err := os.Stat(dstPath); err == nil && !overwrite {
			return nil // base pass: native tooling owns existing files; skip
		}

		srcInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("scaffold: stat %s: %w", path, err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("scaffold: read %s: %w", path, err)
		}

		if err := os.WriteFile(
			dstPath,
			renderTemplate(data, opts),
			srcInfo.Mode().Perm(),
		); err != nil {
			return fmt.Errorf("scaffold: write %s: %w", dstPath, err)
		}

		return nil
	})
}
