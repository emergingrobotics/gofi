# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.2.0] - 2026-10-21

### Changed
- **BREAKING:** The module path changed from `github.com/unifi-go/gofi` to
  `github.com/gherlein/gofi`, matching the repository it actually lives at (the old
  path resolved nowhere). Every import must be updated; there is no compatibility
  shim or redirect.
- **BREAKING:** `gofips`, `gofimac`, `gofinet`, `gofidns`, and `gofiuser` are removed.
  One binary, `gofi`, replaces all five. See README's "Migrating from the old tools"
  for the command mapping.

### Added
- Released via GoReleaser to GitHub Releases and the `gherlein/homebrew-tap` Homebrew
  tap. See the README for `brew install gofi` and `go install` instructions.
- `gofi config` — a TOML config file with named targets, supporting both local
  (username/password) and connector (API key) authentication modes.
- `gofi ips clear`, `gofi network show`, `gofi clients vendor` — new commands with no
  predecessor in the old tools.
- `gofi profile export`/`import` — captures and applies a site's networks, WLANs, and
  fixed IPs as one JSON file.
- `gofi install-skill` — installs the agent skill embedded in the binary
  (`skill/gofi-cli/`) into `$XDG_DATA_HOME/skills/`, per the Skill Self-Installation
  Specification v0.4. `-f`/`--force` replaces a differing copy; exit 3 when it refuses.
- `gofi print-skill` — writes the same embedded skill's `SKILL.md` to stdout, byte for
  byte, with no filesystem change.

### Fixed
- `gofi ips rm --force` and `gofi ips clear` no longer delete user records. Both
  issued `DELETE /rest/user/<id>`, which a UDM answers with 404, after already
  removing the DNS record; they now release the fixed IP and keep the record
  (C-IPS-009). `ips rm --force` now does what its help says: it releases every
  reservation an ambiguous `--name`/`--ip` matches.
- `gofi ips add` conflicts and ambiguous `gofi ips rm` identifiers exit 3
  (refused), not 1 (C-IPS-004, C-IPS-006, C-GLOBAL-012). `ips rm --name` no
  longer silently picks the first of several matches.
- The mock server answers `DELETE /rest/user/<id>` with 404 like a UDM, so code
  depending on it fails in tests instead of only on hardware.
