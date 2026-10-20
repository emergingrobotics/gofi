package installskill

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const name = "demo"

func demoSkill() fstest.MapFS {
	return fstest.MapFS{
		"SKILL.md":      {Data: []byte("---\nname: demo\n---\nbody\n")},
		"refs/extra.md": {Data: []byte("extra\n")},
		"refs/.hidden":  {Data: []byte("dot\n")},
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// assertNoLeftovers fails if any staging or set-aside directory remains
// beside the store (C-SKILL-007).
func assertNoLeftovers(t *testing.T, store string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(store))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".install-skill-") {
			t.Errorf("leftover staging entry %s", entry.Name())
		}
	}
}

func TestStorePath(t *testing.T) {
	cases := []struct {
		name, xdg, home, want string
		wantErr               bool
	}{
		{"xdg set", "/data", "/home/u", "/data/skills", false},
		{"xdg unset", "", "/home/u", "/home/u/.local/share/skills", false},
		{"xdg relative", "rel/data", "/home/u", "/home/u/.local/share/skills", false},
		{"nothing usable", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_DATA_HOME", tc.xdg)
			t.Setenv("HOME", tc.home)
			got, err := StorePath()
			if (err != nil) != tc.wantErr {
				t.Fatalf("StorePath() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("StorePath() = %q, want %q", got, tc.want)
			}
		})
	}
}

// B-SKILL-001, C-SKILL-003: a missing store is created and the whole skill
// directory, dotfiles included, is written.
func TestInstall_freshStore(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")

	outcome, path, err := Install(demoSkill(), name, store, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if outcome != Installed {
		t.Errorf("outcome = %v, want Installed", outcome)
	}
	if path != filepath.Join(store, name) {
		t.Errorf("path = %q, want %q", path, filepath.Join(store, name))
	}
	if got := readFile(t, filepath.Join(path, "refs", ".hidden")); got != "dot\n" {
		t.Errorf("bundled dotfile = %q", got)
	}
	info, err := os.Stat(filepath.Join(path, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("SKILL.md mode = %v, want 0644", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o755 {
		t.Errorf("skill dir mode = %v, want 0755", dirInfo.Mode().Perm())
	}
	assertNoLeftovers(t, store)
}

// B-SKILL-003: an identical install is a no-op, with or without force.
func TestInstall_equivalentIsNoOp(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")
	if _, _, err := Install(demoSkill(), name, store, false); err != nil {
		t.Fatal(err)
	}
	// Permissions are not part of equivalence.
	if err := os.Chmod(filepath.Join(store, name, "SKILL.md"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, force := range []bool{false, true} {
		outcome, _, err := Install(demoSkill(), name, store, force)
		if err != nil {
			t.Fatalf("force=%v: error = %v", force, err)
		}
		if outcome != Unchanged {
			t.Errorf("force=%v: outcome = %v, want Unchanged", force, outcome)
		}
	}
	info, err := os.Stat(filepath.Join(store, name, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Error("a no-op install rewrote the file")
	}
}

// B-SKILL-002: a differing copy is refused without force and left alone.
func TestInstall_differingRefusedWithoutForce(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")
	if _, _, err := Install(demoSkill(), name, store, false); err != nil {
		t.Fatal(err)
	}
	edited := filepath.Join(store, name, "SKILL.md")
	if err := os.WriteFile(edited, []byte("hand edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := Install(demoSkill(), name, store, false)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
	if !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), store) {
		t.Errorf("error %q must name the skill and the store", err)
	}
	if got := readFile(t, edited); got != "hand edit\n" {
		t.Errorf("refused install changed the file to %q", got)
	}
	assertNoLeftovers(t, store)
}

// B-SKILL-002: --force replaces a differing copy, including extra files.
func TestInstall_forceReplaces(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")
	if _, _, err := Install(demoSkill(), name, store, false); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(store, name, "stray.txt")
	if err := os.WriteFile(extra, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	outcome, _, err := Install(demoSkill(), name, store, true)
	if err != nil {
		t.Fatalf("Install(force) error = %v", err)
	}
	if outcome != Replaced {
		t.Errorf("outcome = %v, want Replaced", outcome)
	}
	if _, err := os.Stat(extra); !errors.Is(err, os.ErrNotExist) {
		t.Error("replacement kept a file that is not in the skill")
	}
	same, err := Equivalent(demoSkill(), filepath.Join(store, name))
	if err != nil || !same {
		t.Errorf("after --force the directory is not equivalent (same=%v, err=%v)", same, err)
	}
	assertNoLeftovers(t, store)
}

// A plain file at the skill's name is a differing entry, not an error.
func TestInstall_fileAtTargetIsConflict(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, name), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Install(demoSkill(), name, store, false); !errors.Is(err, ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
	if _, _, err := Install(demoSkill(), name, store, true); err != nil {
		t.Fatalf("Install(force) error = %v", err)
	}
}

// I-SKILL-002: other store entries are untouched.
func TestInstall_leavesOtherSkillsAlone(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skills")
	other := filepath.Join(store, "other", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(other), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Install(demoSkill(), name, store, true); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(store)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("store has %d entries, want 2 (other + %s)", len(entries), name)
	}
	if got := readFile(t, other); got != "other\n" {
		t.Errorf("other skill changed to %q", got)
	}
}

// C-SKILL-007: a store that is a symlink is written through, and staging
// happens beside its real location.
func TestInstall_symlinkedStore(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "dotfiles", "skills")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	data := filepath.Join(root, "data")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(data, "skills")
	if err := os.Symlink(real, store); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Install(demoSkill(), name, store, false); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(real, name, "SKILL.md")); err != nil {
		t.Errorf("skill not written into the real store: %v", err)
	}
	assertNoLeftovers(t, real)
}

// C-SKILL-007: stale leftovers from a crashed run are swept; a fresh one
// (possibly a concurrent install) is not.
func TestInstall_sweepsOnlyStaleLeftovers(t *testing.T) {
	root := t.TempDir()
	store := filepath.Join(root, "skills")
	stale := filepath.Join(root, ".install-skill-demo.111")
	fresh := filepath.Join(root, ".install-skill-demo.222")
	unrelated := filepath.Join(root, ".install-skill-other.333")
	for _, dir := range []string{stale, fresh, unrelated} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-time.Hour)
	for _, dir := range []string{stale, unrelated} {
		if err := os.Chtimes(dir, old, old); err != nil {
			t.Fatal(err)
		}
	}

	if _, _, err := Install(demoSkill(), name, store, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
		t.Error("stale leftover was not swept")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh staging directory was swept; it may belong to a running install")
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Error("another skill's staging directory was swept")
	}
}

// I-SKILL-001: a write failure leaves no partial directory in the store.
func TestInstall_unwritableStoreLeavesNothing(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	root := t.TempDir()
	store := filepath.Join(root, "skills")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(store, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(store, 0o755) })

	if _, _, err := Install(demoSkill(), name, store, false); err == nil {
		t.Fatal("Install() into a read-only store: error = nil")
	}
	if _, err := os.Lstat(filepath.Join(store, name)); !errors.Is(err, os.ErrNotExist) {
		t.Error("failed install left an entry in the store")
	}
	assertNoLeftovers(t, store)
}

func TestEquivalent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	if _, _, err := Install(demoSkill(), name, dir, false); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(dir, name)

	same, err := Equivalent(demoSkill(), installed)
	if err != nil || !same {
		t.Fatalf("fresh install: same=%v err=%v, want true", same, err)
	}

	missing := demoSkill()
	missing["new.md"] = &fstest.MapFile{Data: []byte("n")}
	if same, _ := Equivalent(missing, installed); same {
		t.Error("a file missing from the install must make it non-equivalent")
	}

	if same, _ := Equivalent(demoSkill(), filepath.Join(dir, "absent")); same {
		t.Error("an absent directory must be non-equivalent")
	}
}
