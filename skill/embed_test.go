package skill

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// frontmatter returns SKILL.md's frontmatter lines and body, failing the test
// if the file does not open with a `---` ... `---` block.
func frontmatter(t *testing.T) (map[string]string, string) {
	t.Helper()
	data, err := fs.ReadFile(FS(), "SKILL.md")
	if err != nil {
		t.Fatalf("embedded SKILL.md: %v", err)
	}
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		t.Fatal("SKILL.md must open with a --- frontmatter line")
	}
	rest := strings.TrimPrefix(text, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		t.Fatal("SKILL.md frontmatter has no closing --- line")
	}
	fields := map[string]string{}
	for _, line := range strings.Split(rest[:end], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") {
			t.Fatalf("frontmatter line %q is not a single-line key: value (C-SKILL-002)", line)
		}
		fields[key] = strings.TrimSpace(value)
	}
	return fields, rest[end+len("\n---\n"):]
}

// C-SKILL-001/002: the embedded skill is a conforming skill directory.
func TestEmbeddedSkill_conforms(t *testing.T) {
	fields, _ := frontmatter(t)

	if fields["name"] != Name {
		t.Errorf("frontmatter name = %q, want %q (the directory name)", fields["name"], Name)
	}
	if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(Name) {
		t.Errorf("skill name %q is not a valid store directory name", Name)
	}
	desc := fields["description"]
	if desc == "" || strings.HasPrefix(desc, ">") || strings.HasPrefix(desc, "|") {
		t.Errorf("description must be one inline line, got %q", desc)
	}
	if len(desc) > 200 {
		t.Errorf("description is %d characters; keep it to one short sentence", len(desc))
	}
}

// C-SKILL-002: the skill is installed from a lone binary, so it must not point
// at files that only exist in this repository.
func TestEmbeddedSkill_isSelfContained(t *testing.T) {
	_, body := frontmatter(t)
	for _, ref := range []string{
		"docs/", "REQUIREMENTS.md", "VISION.md", "CLAUDE.md", "README", "src/", "utilities/",
		"make ", "the repo",
	} {
		if strings.Contains(body, ref) {
			t.Errorf("SKILL.md references %q, which is not inside the installed skill", ref)
		}
	}
}

// C-SKILL-001: the skill directory carries nothing but the skill -- in
// particular not a Go source file.
func TestEmbeddedSkill_carriesNoGoSource(t *testing.T) {
	err := fs.WalkDir(FS(), ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") {
			t.Errorf("skill directory contains Go source %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
