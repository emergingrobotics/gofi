// Package installskill writes gofi's embedded agent skill into the user's
// skills store, per the Skill Self-Installation Specification v0.3 (the SKILL
// area of REQUIREMENTS.md). It touches only the local filesystem.
package installskill

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrConflict means a different directory already holds the skill's name and
// --force was not given (C-SKILL-005). A guard, not a failure: the CLI reports
// it as exit 3.
var ErrConflict = errors.New("conflict")

// staleAfter is how old a leftover staging directory must be before a new run
// sweeps it, so a sweep never removes the staging area of an install running
// concurrently (C-SKILL-007).
const staleAfter = 10 * time.Minute

// Outcome reports what Install did.
type Outcome int

const (
	// Installed means no directory existed and the skill was written.
	Installed Outcome = iota
	// Replaced means a differing directory was replaced under --force.
	Replaced
	// Unchanged means an equivalent directory was already present.
	Unchanged
)

// StorePath resolves the skills store: $XDG_DATA_HOME/skills, falling back to
// $HOME/.local/share/skills when XDG_DATA_HOME is unset, empty, or relative
// (C-SKILL-003).
func StorePath() (string, error) {
	if base := os.Getenv("XDG_DATA_HOME"); base != "" && filepath.IsAbs(base) {
		return filepath.Join(base, "skills"), nil
	}
	home := os.Getenv("HOME")
	if home == "" || !filepath.IsAbs(home) {
		return "", errors.New("cannot resolve the skills store: XDG_DATA_HOME and HOME are both unset or relative")
	}
	return filepath.Join(home, ".local", "share", "skills"), nil
}

// Install writes skill (rooted at the skill directory) into store as the
// directory name. It refuses with ErrConflict when a differing directory
// already exists, unless force; an equivalent existing directory is a no-op.
// It returns the installed directory's path.
func Install(skill fs.FS, name, store string, force bool) (Outcome, string, error) {
	if err := os.MkdirAll(store, 0o755); err != nil {
		return 0, "", fmt.Errorf("creating skills store %s: %w", store, err)
	}
	// Stage beside the store's real location, so the final move is a rename
	// even when the store is a symlink into another filesystem.
	realStore, err := filepath.EvalSymlinks(store)
	if err != nil {
		return 0, "", fmt.Errorf("resolving skills store %s: %w", store, err)
	}
	stagingArea := filepath.Dir(realStore)
	target := filepath.Join(realStore, name)
	sweepStale(stagingArea, name)

	_, statErr := os.Lstat(target)
	exists := statErr == nil
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return 0, "", fmt.Errorf("checking %s: %w", target, statErr)
	}
	if exists {
		same, err := Equivalent(skill, target)
		if err != nil {
			return 0, "", err
		}
		if same {
			return Unchanged, target, nil
		}
		if !force {
			return 0, "", fmt.Errorf("%w: skill %q already exists in %s and differs; pass --force to replace it",
				ErrConflict, name, realStore)
		}
	}

	stage, err := os.MkdirTemp(stagingArea, stagingPrefix(name))
	if err != nil {
		return 0, "", fmt.Errorf("creating staging directory: %w", err)
	}
	// A cleanup that fails leaves a leftover the next run sweeps.
	defer func() { _ = os.RemoveAll(stage) }()
	if err := writeTree(skill, stage); err != nil {
		return 0, "", err
	}

	if !exists {
		if err := os.Rename(stage, target); err != nil {
			return 0, "", fmt.Errorf("moving skill into %s: %w", target, err)
		}
		return Installed, target, nil
	}

	// No portable atomic directory replacement exists: set the old directory
	// aside only now that the new one is fully staged, move the new one in,
	// then delete the old. If the move in fails, put the old one back.
	aside := stage + ".old"
	if err := os.Rename(target, aside); err != nil {
		return 0, "", fmt.Errorf("setting aside %s: %w", target, err)
	}
	defer func() { _ = os.RemoveAll(aside) }()
	if err := os.Rename(stage, target); err != nil {
		if restoreErr := os.Rename(aside, target); restoreErr != nil {
			return 0, "", fmt.Errorf("moving skill into %s: %w (restoring the previous copy also failed: %v; it is at %s)",
				target, err, restoreErr, aside)
		}
		return 0, "", fmt.Errorf("moving skill into %s: %w", target, err)
	}
	return Replaced, target, nil
}

// Equivalent reports whether dir holds exactly the skill: the same set of
// relative paths and byte-identical files. Permissions and timestamps are not
// compared (Skills Directory Specification §5.5).
func Equivalent(skill fs.FS, dir string) (bool, error) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		// A file, or a symlink to nothing, at the skill's name is simply
		// not the skill.
		return false, nil
	}
	want, err := listTree(skill)
	if err != nil {
		return false, fmt.Errorf("reading embedded skill: %w", err)
	}
	installed := os.DirFS(dir)
	have, err := listTree(installed)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", dir, err)
	}
	if len(want) != len(have) {
		return false, nil
	}
	for path, isDir := range want {
		haveDir, ok := have[path]
		if !ok || haveDir != isDir {
			return false, nil
		}
		if isDir {
			continue
		}
		a, err := fs.ReadFile(skill, path)
		if err != nil {
			return false, fmt.Errorf("reading embedded %s: %w", path, err)
		}
		b, err := fs.ReadFile(installed, path)
		if err != nil {
			return false, fmt.Errorf("reading %s: %w", filepath.Join(dir, path), err)
		}
		if !bytes.Equal(a, b) {
			return false, nil
		}
	}
	return true, nil
}

// listTree maps every relative path under the root (excluding the root
// itself) to whether it is a directory.
func listTree(fsys fs.FS) (map[string]bool, error) {
	paths := map[string]bool{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != "." {
			paths[path] = d.IsDir()
		}
		return nil
	})
	return paths, err
}

// writeTree copies the skill into dir, which already exists. Directories are
// written 0755 and files 0644: readable by the user, never world-writable.
func writeTree(skill fs.FS, dir string) error {
	if err := os.Chmod(dir, 0o755); err != nil {
		return fmt.Errorf("preparing staging directory: %w", err)
	}
	return fs.WalkDir(skill, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		dest := filepath.Join(dir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.Mkdir(dest, 0o755)
		}
		data, err := fs.ReadFile(skill, path)
		if err != nil {
			return fmt.Errorf("reading embedded %s: %w", path, err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("staging %s: %w", path, err)
		}
		return nil
	})
}

func stagingPrefix(name string) string { return ".install-skill-" + name + "." }

// sweepStale removes staging and set-aside directories an earlier, crashed
// run left beside the store. Best effort: a leftover that cannot be removed
// does not block the install.
func sweepStale(stagingArea, name string) {
	entries, err := os.ReadDir(stagingArea)
	if err != nil {
		return
	}
	prefix := stagingPrefix(name)
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		info, err := entry.Info()
		if err != nil || time.Since(info.ModTime()) < staleAfter {
			continue
		}
		_ = os.RemoveAll(filepath.Join(stagingArea, entry.Name()))
	}
}
