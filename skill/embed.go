// Package skill carries the agent skill that teaches the gofi CLI, embedded in
// the binary so `gofi install-skill` can install it with no files beside the
// executable (C-SKILL-001).
//
// This file sits one level above the skill directory on purpose: a directive
// cannot embed its own package directory, so an embed.go inside gofi-cli/ would
// have to embed `all:*` and ship itself as part of the skill.
package skill

import (
	"embed"
	"io/fs"
)

// Name is the carried skill's name: its directory name, its frontmatter name,
// and the directory it installs as in the store.
const Name = "gofi-cli"

//go:embed all:gofi-cli
var embedded embed.FS

// FS returns the skill directory's contents, rooted at the skill directory
// (SKILL.md at its top level).
func FS() fs.FS {
	sub, err := fs.Sub(embedded, Name)
	if err != nil {
		// Unreachable: Name is a valid path and the directive embeds it.
		panic(err)
	}
	return sub
}

// MD returns SKILL.md's exact embedded bytes -- frontmatter and body, nothing
// added or stripped -- for `gofi print-skill` to write verbatim (C-SKILL-008).
func MD() ([]byte, error) {
	return fs.ReadFile(embedded, Name+"/SKILL.md")
}
