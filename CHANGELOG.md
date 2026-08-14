# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
- **BREAKING:** `gofips`, `gofimac`, `gofinet`, `gofidns`, and `gofiuser` are removed.
  One binary, `gofi`, replaces all five. See README's "Migrating from the old tools"
  for the command mapping.

### Added
- `gofi config` — a TOML config file with named targets, supporting both local
  (username/password) and connector (API key) authentication modes.
- `gofi ips clear`, `gofi network show`, `gofi clients vendor` — new commands with no
  predecessor in the old tools.
- `gofi profile export`/`import` — captures and applies a site's networks, WLANs, and
  fixed IPs as one JSON file.
