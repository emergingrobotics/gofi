# gofi CLI Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the five flag-mode binaries (`gofips`, `gofimac`, `gofinet`, `gofidns`, `gofiuser`) with one cobra-based binary, `gofi <area> <action>`, following `network-cli-convention.md`, with a new TOML config file, dual-auth-mode target resolution, and a new `profile` area — cutting over in one release.

**Architecture:** One binary at `utilities/gofi/` (cobra command tree, flag wiring only) over seven business-logic packages at `utilities/internal/{ips,dns,network,clients,users,profile}` plus a new `utilities/internal/config` package for the TOML file. Each area's existing `operations.go`/`format.go`/`parse.go` files move from their old binary's `package main` into their own importable package, unchanged in logic except where `REQUIREMENTS.md` calls for new behavior (`ips clear`, `clients vendor`, exit-code changes, `network show`, `profile`).

**Tech Stack:** Go, `github.com/spf13/cobra` (new dependency), `github.com/BurntSushi/toml` (new dependency), `golang.org/x/term` (new dependency, for secret prompts), the existing `gofi` library (`src/`).

**Spec:** [`REQUIREMENTS.md`](../../REQUIREMENTS.md) (detailed, numbered requirements — the contract this plan implements), [`VISION.md`](../../VISION.md) (why and the target tree), [`network-cli-convention.md`](../../../network-cli-convention.md) (the shared grammar both `gofi` and `gogl` follow).

## Global Constraints

Copied verbatim from `REQUIREMENTS.md`'s `GLOBAL` section; every task below implicitly inherits these.

- C-GLOBAL-001: One binary, `gofi <area> <action>`. Five existing binaries removed in the same release, with a published mapping table. No wrapper binaries.
- C-GLOBAL-002: Every write action that can meaningfully preview its effect supports `--dry-run`.
- C-GLOBAL-003: `--force` waives only a state guard, never input validation.
- C-GLOBAL-004: `--yes` waives only a confirmation prompt, never a `--force`-gated guard.
- C-GLOBAL-005: No secret-value flag exists on any command.
- C-GLOBAL-006: Secret resolution order: environment variable, then `*_command` in config.toml, then interactive prompt with echo off from `/dev/tty`.
- C-GLOBAL-007: `--secure`/`--insecure` against a connector-mode target is a usage error (exit 2).
- C-GLOBAL-008: `UNIFI_API_KEY` selects connector mode; `UNIFI_CONSOLE_ID` required with it; local username/password ignored with a stderr note if also set.
- C-GLOBAL-009: `--target <name>` resolves a controller and a site together, as one unit.
- C-GLOBAL-010: Connector-mode requests always verify TLS regardless of local flags.
- C-GLOBAL-011: Every read command accepts `--output text|json`; JSON empty collections render `[]`, never `null`.
- C-GLOBAL-012: Exit codes fixed: `0` success, `1` error, `2` usage error, `3` refused by a guard.

---

## Phase 0: Foundation

Everything below depends on this phase. Builds the config file, the dual-mode connection resolver, and the empty `gofi` binary shell with global flags and error handling — no area commands yet.

### Task 0.1: Add dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add the three new dependencies**

```bash
cd /home/gherlein/src/er/gofi
go get github.com/spf13/cobra@latest
go get github.com/BurntSushi/toml@latest
go get golang.org/x/term@latest
go mod tidy
```

- [ ] **Step 2: Verify the module still builds**

Run: `go build ./...`
Expected: succeeds with no errors (nothing imports the new deps yet, so this only proves `go.sum` is consistent).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add cobra, toml, and term dependencies for the gofi CLI migration"
```

### Task 0.2: `utilities/internal/config` package

**Files:**
- Create: `utilities/internal/config/config.go`
- Test: `utilities/internal/config/config_test.go`

**Interfaces:**
- Produces: `config.Path() string`, `config.Load() (*config.File, error)`, `config.LoadFrom(path string) (*config.File, error)`, `(*config.File).Path() string`, `(*config.File).Names() []string`, `(*config.File).Resolve(name string) (*config.Target, error)`, `(*config.Target).Name() string`, `(*config.Target).Mode() string` (`"local"` or `"connector"`), `(*config.Target).Secret(prompt func() (string, error)) (string, error)`.

- [ ] **Step 1: Write the failing tests**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_missingFileIsNotAnError(t *testing.T) {
	f, err := LoadFrom(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(f.Targets) != 0 {
		t.Errorf("Targets = %v, want empty", f.Targets)
	}
}

func TestLoad_localTarget(t *testing.T) {
	path := writeConfig(t, `
default = "home"

[targets.home]
host             = "192.168.1.1"
site             = "default"
password_command = "echo secret"
`)
	f, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	target, err := f.Resolve("")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if target.Mode() != "local" {
		t.Errorf("Mode() = %q, want local", target.Mode())
	}
	if target.Host != "192.168.1.1" || target.Site != "default" {
		t.Errorf("target = %+v, want host 192.168.1.1 site default", target)
	}
}

func TestLoad_connectorTarget(t *testing.T) {
	path := writeConfig(t, `
[targets.cloud]
site           = "default"
console_id     = "abc123"
api_key_command = "echo key"
`)
	f, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	target, err := f.Resolve("cloud")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if target.Mode() != "connector" {
		t.Errorf("Mode() = %q, want connector", target.Mode())
	}
}

func TestLoad_rejectsMixedModeTarget(t *testing.T) {
	path := writeConfig(t, `
[targets.bad]
host       = "192.168.1.1"
console_id = "abc123"
`)
	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want an error for a mixed-mode target")
	}
}

func TestLoad_rejectsUnknownDefault(t *testing.T) {
	path := writeConfig(t, `
default = "missing"

[targets.home]
host = "192.168.1.1"
`)
	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want an error for an unresolvable default")
	}
}

func TestResolve_singleTargetNeedsNoName(t *testing.T) {
	path := writeConfig(t, `
[targets.only]
host = "192.168.1.1"
`)
	f, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	target, err := f.Resolve("")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if target.Name() != "only" {
		t.Errorf("Name() = %q, want only", target.Name())
	}
}

func TestResolve_unknownNameIsAnError(t *testing.T) {
	path := writeConfig(t, `
[targets.home]
host = "192.168.1.1"
`)
	f, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, err := f.Resolve("nope"); err == nil {
		t.Fatal("Resolve(\"nope\") error = nil, want an error naming the unknown target")
	}
}

func TestTargetSecret_environmentWins(t *testing.T) {
	t.Setenv(EnvPassword, "from-env")
	target := &Target{PasswordCommand: "echo from-command"}
	secret, err := target.Secret(func() (string, error) { return "from-prompt", nil })
	if err != nil {
		t.Fatalf("Secret() error = %v", err)
	}
	if secret != "from-env" {
		t.Errorf("Secret() = %q, want from-env", secret)
	}
}

func TestTargetSecret_commandFallsBackFromEnv(t *testing.T) {
	t.Setenv(EnvPassword, "")
	target := &Target{PasswordCommand: "echo from-command"}
	secret, err := target.Secret(nil)
	if err != nil {
		t.Fatalf("Secret() error = %v", err)
	}
	if secret != "from-command" {
		t.Errorf("Secret() = %q, want from-command", secret)
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/config/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Write the implementation**

```go
// Package config reads gofi's TOML configuration and resolves how to reach a
// controller and site.
//
// A target is either local-mode (host + optional password_command) or
// connector-mode (console_id + optional api_key_command) — never a mix. The
// file holds everything except secrets: a secret comes from the environment,
// from a command the file names, or from a prompt, in that order.
package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// Environment variables, matching utilities/internal/conn today.
const (
	EnvUsername  = "UNIFI_USERNAME"
	EnvPassword  = "UNIFI_PASSWORD"
	EnvHost      = "UNIFI_CONTROLLER_IP"
	EnvAPIKey    = "UNIFI_API_KEY"
	EnvConsoleID = "UNIFI_CONSOLE_ID"
)

// File is a parsed configuration file.
type File struct {
	Default string             `toml:"default"`
	Output  string             `toml:"output"`
	Targets map[string]*Target `toml:"targets"`

	path string
}

// Target is one controller+site's connection settings.
type Target struct {
	// Local-mode fields.
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Secure   bool   `toml:"secure"`
	Username string `toml:"username"`

	// Connector-mode fields.
	ConsoleID string `toml:"console_id"`

	// Common.
	Site            string `toml:"site"`
	PasswordCommand string `toml:"password_command"`
	APIKeyCommand   string `toml:"api_key_command"`

	name string
}

// Name returns the key this target was defined under.
func (t *Target) Name() string { return t.name }

// Mode reports "local" or "connector".
func (t *Target) Mode() string {
	if t.ConsoleID != "" {
		return "connector"
	}
	return "local"
}

// Path returns the configuration file's location, honoring XDG_CONFIG_HOME.
func Path() string {
	if explicit := os.Getenv("GOFI_CONFIG"); explicit != "" {
		return explicit
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".config", "gofi", "config.toml")
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "gofi", "config.toml")
}

// Load reads the configuration file. A missing file is not an error: gofi
// works entirely from flags and the environment.
func Load() (*File, error) {
	return LoadFrom(Path())
}

// LoadFrom reads a specific path.
func LoadFrom(path string) (*File, error) {
	f := &File{Targets: map[string]*Target{}, path: path}

	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if err := toml.Unmarshal(raw, f); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if f.Targets == nil {
		f.Targets = map[string]*Target{}
	}
	for name, target := range f.Targets {
		target.name = name
	}
	f.path = path

	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return f, nil
}

func (f *File) validate() error {
	if f.Output != "" && f.Output != "text" && f.Output != "json" {
		return fmt.Errorf("output = %q, want \"text\" or \"json\"", f.Output)
	}
	if f.Default != "" {
		if _, ok := f.Targets[f.Default]; !ok {
			return fmt.Errorf("default = %q but no [targets.%s] section", f.Default, f.Default)
		}
	}
	for name, target := range f.Targets {
		if target.Host == "" && target.ConsoleID == "" {
			return fmt.Errorf("targets.%s has neither host nor console_id", name)
		}
		if target.Host != "" && target.ConsoleID != "" {
			return fmt.Errorf("targets.%s has both host and console_id; a target is local-mode or connector-mode, never both", name)
		}
	}
	return nil
}

// Path returns where this file was read from.
func (f *File) Path() string { return f.path }

// Names returns the configured target names, sorted.
func (f *File) Names() []string {
	names := make([]string, 0, len(f.Targets))
	for name := range f.Targets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Resolve picks a target by name, falling back to the default, or the only
// target if exactly one is defined.
func (f *File) Resolve(name string) (*Target, error) {
	if name != "" {
		target, ok := f.Targets[name]
		if !ok {
			return nil, fmt.Errorf("no target named %q in %s (have: %s)",
				name, f.path, strings.Join(f.Names(), ", "))
		}
		return target, nil
	}
	if f.Default != "" {
		return f.Targets[f.Default], nil
	}
	if len(f.Targets) == 1 {
		for _, target := range f.Targets {
			return target, nil
		}
	}
	if len(f.Targets) > 1 {
		return nil, fmt.Errorf("several targets in %s and no default; pass --target (have: %s)",
			f.path, strings.Join(f.Names(), ", "))
	}
	return nil, nil
}

// Secret resolves this target's secret (password or API key).
//
// Order: the environment variable for this target's mode, then the
// configured command, then prompt. prompt may be nil in non-interactive
// contexts.
func (t *Target) Secret(prompt func() (string, error)) (string, error) {
	envVar, command := EnvPassword, t.PasswordCommand
	if t.Mode() == "connector" {
		envVar, command = EnvAPIKey, t.APIKeyCommand
	}

	if v := os.Getenv(envVar); v != "" {
		return v, nil
	}
	if command != "" {
		out, err := runSecretCommand(command)
		if err != nil {
			return "", err
		}
		if out != "" {
			return out, nil
		}
		return "", fmt.Errorf("secret command for %q produced no output: %s", t.name, command)
	}
	if prompt != nil {
		return prompt()
	}
	return "", fmt.Errorf("no secret: set %s, configure a *_command, or run interactively", envVar)
}

func runSecretCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("secret command failed (%s): %w", command, err)
	}
	line := string(out)
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	return strings.TrimRight(line, "\r"), nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/config/... -v`
Expected: PASS, all nine tests.

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/config
git commit -m "feat: add gofi config.toml package with dual-mode targets"
```

### Task 0.3: Secret prompt helper in `utilities/internal/conn`

**Files:**
- Create: `utilities/internal/conn/secret.go`
- Test: `utilities/internal/conn/secret_test.go`

**Interfaces:**
- Produces: `conn.ReadSecret(prompt, command string) (string, error)`.

- [ ] **Step 1: Write the failing test**

```go
package conn

import "testing"

func TestReadSecret_commandTakesPriorityOverPrompt(t *testing.T) {
	secret, err := ReadSecret("unused prompt: ", "echo from-command")
	if err != nil {
		t.Fatalf("ReadSecret() error = %v", err)
	}
	if secret != "from-command" {
		t.Errorf("ReadSecret() = %q, want from-command", secret)
	}
}

func TestReadSecret_commandFailureIsReported(t *testing.T) {
	if _, err := ReadSecret("prompt: ", "false"); err == nil {
		t.Fatal("ReadSecret() error = nil, want an error for a failing command")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/internal/conn/... -run TestReadSecret -v`
Expected: FAIL — `ReadSecret` undefined.

- [ ] **Step 3: Write the implementation**

```go
// secret.go obtains a secret without putting it on the command line: no CLI
// flag ever accepts a password or API key value directly (C-GLOBAL-005).
package conn

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// ReadSecret resolves a secret: the named command if given, otherwise a
// prompt on the controlling terminal with echo disabled.
func ReadSecret(prompt, command string) (string, error) {
	if command != "" {
		return runSecretCommand(command)
	}
	return promptSecret(prompt)
}

// promptSecret reads from /dev/tty rather than stdin, so a secret can still
// be typed while stdin carries piped data (e.g. `gofi ips import -`).
func promptSecret(prompt string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("no terminal to prompt on: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("not a terminal; use the --*-command form instead")
	}

	fmt.Fprint(tty, prompt)
	raw, err := term.ReadPassword(fd)
	fmt.Fprintln(tty)
	if err != nil {
		return "", fmt.Errorf("reading the secret: %w", err)
	}

	secret := strings.TrimRight(string(raw), "\r\n")
	if secret == "" {
		return "", errors.New("empty secret")
	}
	return secret, nil
}

func runSecretCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("secret command failed (%s): %w", command, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if !scanner.Scan() {
		return "", fmt.Errorf("secret command produced no output: %s", command)
	}
	secret := strings.TrimRight(scanner.Text(), "\r")
	if secret == "" {
		return "", fmt.Errorf("secret command produced an empty first line: %s", command)
	}
	return secret, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/conn/... -run TestReadSecret -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/conn/secret.go utilities/internal/conn/secret_test.go
git commit -m "feat: add secret prompt helper for gofi CLI"
```

### Task 0.4: Extend `conn.ResolveConfig` to accept a resolved `*config.Target`

**Files:**
- Modify: `utilities/internal/conn/conn.go`
- Test: `utilities/internal/conn/conn_test.go` (existing file — add cases)

**Interfaces:**
- Consumes: `config.Target` (Task 0.2), `config.EnvPassword`/`config.EnvAPIKey`/`config.EnvHost`/`config.EnvConsoleID` (Task 0.2).
- Produces: `conn.ResolveTargetConfig(w io.Writer, target *config.Target, hostFlag string, portFlag int, siteFlag string, secureFlag *bool, secret func() (string, error)) (*gofi.Config, error)`. `secureFlag` is a `*bool` (nil = flag not given) so the function can implement C-GLOBAL-007 (reject `--secure` against a connector target) by distinguishing "not passed" from "passed false".

- [ ] **Step 1: Write the failing tests**

```go
func TestResolveTargetConfig_secureFlagRejectedForConnector(t *testing.T) {
	target := &config.Target{ConsoleID: "abc", APIKeyCommand: "echo key"}
	secure := true
	_, err := ResolveTargetConfig(io.Discard, target, "", 0, "", &secure, nil)
	if err == nil {
		t.Fatal("ResolveTargetConfig() error = nil, want a usage error for --secure against a connector target")
	}
}

func TestResolveTargetConfig_connectorAlwaysVerifiesTLS(t *testing.T) {
	target := &config.Target{ConsoleID: "abc", APIKeyCommand: "echo key"}
	cfg, err := ResolveTargetConfig(io.Discard, target, "", 0, "", nil, nil)
	if err != nil {
		t.Fatalf("ResolveTargetConfig() error = %v", err)
	}
	if cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false for connector mode")
	}
}

func TestResolveTargetConfig_localModeFlagOverridesTarget(t *testing.T) {
	target := &config.Target{Host: "192.168.1.1", Site: "default", PasswordCommand: "echo pw"}
	cfg, err := ResolveTargetConfig(io.Discard, target, "10.0.0.5", 0, "", nil, nil)
	if err != nil {
		t.Fatalf("ResolveTargetConfig() error = %v", err)
	}
	if cfg.Host != "10.0.0.5" {
		t.Errorf("Host = %q, want the flag override 10.0.0.5", cfg.Host)
	}
}

func TestResolveTargetConfig_nilTargetFallsBackToEnvAndFlags(t *testing.T) {
	t.Setenv(config.EnvUsername, "admin")
	t.Setenv(config.EnvPassword, "secret")
	cfg, err := ResolveTargetConfig(io.Discard, nil, "192.168.1.1", 443, "default", nil, nil)
	if err != nil {
		t.Fatalf("ResolveTargetConfig() error = %v", err)
	}
	if cfg.Host != "192.168.1.1" || cfg.Username != "admin" || cfg.Password != "secret" {
		t.Errorf("cfg = %+v, want host/username/password from flags+env", cfg)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/conn/... -run TestResolveTargetConfig -v`
Expected: FAIL — `ResolveTargetConfig` undefined.

- [ ] **Step 3: Write the implementation**

Append to `utilities/internal/conn/conn.go` (the existing `ResolveConfig` and its constants stay, unchanged, for now — `ResolveTargetConfig` is additive and becomes what `root.go` actually calls):

```go
import "github.com/unifi-go/gofi/utilities/internal/config"

// ResolveTargetConfig builds a *gofi.Config from a resolved config target
// (may be nil), CLI flag overrides, and the environment, implementing
// C-GLOBAL-006 through C-GLOBAL-010.
//
// secureFlag is a pointer so ResolveTargetConfig can tell "the operator
// passed --secure/--insecure" from "the flag was not given" — required to
// reject the flag against a connector target (C-GLOBAL-007) rather than
// silently ignoring it.
func ResolveTargetConfig(w io.Writer, target *config.Target, hostFlag string, portFlag int, siteFlag string, secureFlag *bool, prompt func() (string, error)) (*gofi.Config, error) {
	mode := "local"
	if target != nil {
		mode = target.Mode()
	}
	if os.Getenv(EnvAPIKey) != "" {
		mode = "connector"
	}

	if mode == "connector" {
		if secureFlag != nil {
			return nil, fmt.Errorf("--secure/--insecure has no effect against a connector target: connector traffic is always TLS-verified")
		}
		return resolveConnectorConfig(w, target, siteFlag, prompt)
	}
	return resolveLocalConfig(w, target, hostFlag, portFlag, siteFlag, secureFlag, prompt)
}

func resolveConnectorConfig(w io.Writer, target *config.Target, siteFlag string, prompt func() (string, error)) (*gofi.Config, error) {
	consoleID := os.Getenv(EnvConsoleID)
	apiKey := os.Getenv(EnvAPIKey)
	site := siteFlag

	if target != nil {
		if consoleID == "" {
			consoleID = target.ConsoleID
		}
		if site == "" {
			site = target.Site
		}
		if apiKey == "" {
			secret, err := target.Secret(prompt)
			if err != nil {
				return nil, err
			}
			apiKey = secret
		}
	}
	if consoleID == "" {
		return nil, fmt.Errorf("%s is required for connector mode", EnvConsoleID)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no API key: set %s, configure api_key_command, or run interactively", EnvAPIKey)
	}
	if site == "" {
		site = "default"
	}

	if os.Getenv(EnvUsername) != "" || os.Getenv(EnvPassword) != "" {
		fmt.Fprintf(w, "Note: %s is set; ignoring %s/%s.\n", EnvAPIKey, EnvUsername, EnvPassword)
	}

	return &gofi.Config{
		APIKey:        apiKey,
		ConsoleID:     consoleID,
		Site:          site,
		SkipTLSVerify: false,
		RetryConfig: &gofi.RetryConfig{
			MaxRetries:     8,
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
		},
	}, nil
}

func resolveLocalConfig(w io.Writer, target *config.Target, hostFlag string, portFlag int, siteFlag string, secureFlag *bool, prompt func() (string, error)) (*gofi.Config, error) {
	host, port, site := hostFlag, portFlag, siteFlag
	secure := false
	if secureFlag != nil {
		secure = *secureFlag
	}
	username, password := os.Getenv(EnvUsername), os.Getenv(EnvPassword)

	if target != nil {
		if host == "" {
			host = target.Host
		}
		if port == 0 {
			port = target.Port
		}
		if site == "" {
			site = target.Site
		}
		if secureFlag == nil {
			secure = target.Secure
		}
		if username == "" {
			username = target.Username
		}
		if password == "" {
			secret, err := target.Secret(prompt)
			if err != nil {
				return nil, err
			}
			password = secret
		}
	}
	if host == "" {
		host = os.Getenv(EnvControllerIP)
	}
	if port == 0 {
		port = 443
	}
	if site == "" {
		site = "default"
	}
	if username == "" {
		return nil, fmt.Errorf("%s environment variable is required", EnvUsername)
	}
	if password == "" {
		return nil, fmt.Errorf("no password: set %s, configure password_command, or run interactively", EnvPassword)
	}
	if host == "" {
		return nil, fmt.Errorf("--host is required (or set %s, or configure a target)", EnvControllerIP)
	}

	return &gofi.Config{
		Host:          host,
		Port:          port,
		Username:      username,
		Password:      password,
		Site:          site,
		SkipTLSVerify: !secure,
		RetryConfig: &gofi.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 200 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
		},
	}, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/conn/... -v`
Expected: PASS, all cases including the pre-existing `ResolveConfig` tests (untouched).

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/conn/conn.go utilities/internal/conn/conn_test.go
git commit -m "feat: resolve gofi connections from a config target, local or connector mode"
```

### Task 0.5: `utilities/gofi` binary shell — root command, global flags, exit codes

**Files:**
- Create: `utilities/gofi/main.go`
- Create: `utilities/gofi/root.go`
- Create: `utilities/gofi/output.go`

**Interfaces:**
- Consumes: `config.Load`, `config.File.Resolve`, `conn.ResolveTargetConfig`, `conn.ReadSecret` (Tasks 0.2–0.4).
- Produces: package-level `opts globals` (mirrors gogl's `utilities/gogl/root.go`), `connect() (gofi.Client, error)`, `explain(err error) error`, `asJSON() bool`, `writeJSON(w io.Writer, v any) error`, `errRefused error`, `errUsage error`. Every area's command file (Phases 1–6) calls `connect()`/`explain()`/`asJSON()`/`writeJSON()` and returns errors wrapping `errUsage`/`errRefused` to get the right exit code.

- [ ] **Step 1: Write `output.go`**

```go
package main

import (
	"encoding/json"
	"io"
)

// writeJSON emits v as indented JSON. Indented because this output is read
// by people at least as often as by jq (C-GLOBAL-011).
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
```

- [ ] **Step 2: Write `root.go`**

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/config"
	"github.com/unifi-go/gofi/utilities/internal/conn"
)

// globals holds every persistent flag, shared by every command (C-GLOBAL-009).
type globals struct {
	target string
	host   string
	port   int
	site   string
	secure *bool // nil = flag not given, distinguishing "not passed" for C-GLOBAL-007
	output string

	file          *config.File
	resolvedTarget *config.Target
}

var opts globals

const (
	outputText = "text"
	outputJSON = "json"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "gofi",
		Short: "Manage a UniFi controller's networks, DNS, reservations, and clients",
		Long: `gofi manages a UniFi controller (UDM Pro and compatible consoles) over its
local or cloud-connector API.

It exists to make a site's addressing, DNS names, networks, and known clients
reproducible from a file rather than from click history in the UniFi app.`,

		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadConfig(cmd)
		},
	}

	flags := root.PersistentFlags()
	flags.StringVar(&opts.target, "target", "", "named controller+site from the config file")
	flags.StringVarP(&opts.host, "host", "H", "", "controller address")
	flags.IntVarP(&opts.port, "port", "p", 0, "controller port (default 443)")
	flags.StringVarP(&opts.site, "site", "S", "", "site name (default \"default\")")
	flags.StringVar(&opts.output, "output", outputText, "output format: text or json")
	flags.Bool("secure", false, "enforce TLS certificate verification (local mode only)")

	root.AddCommand(
		newConfigCommand(),
		// newIPsCommand(), newDNSCommand(), newNetworkCommand(), newClientsCommand(),
		// newUsersCommand(), newProfileCommand() are added by Phases 1-6.
	)
	return root
}

func loadConfig(cmd *cobra.Command) error {
	file, err := config.Load()
	if err != nil {
		return err
	}
	opts.file = file

	if opts.output != outputText && opts.output != outputJSON {
		return fmt.Errorf("%w: --output %q: want %q or %q", errUsage, opts.output, outputText, outputJSON)
	}
	if !cmd.Flags().Changed("output") && file.Output != "" {
		opts.output = file.Output
	}

	if cmd.Flags().Changed("secure") {
		v, _ := cmd.Flags().GetBool("secure")
		opts.secure = &v
	}

	target, err := file.Resolve(opts.target)
	if err != nil {
		return fmt.Errorf("%w: %s", errUsage, err)
	}
	opts.resolvedTarget = target
	return nil
}

// connect resolves the connection, authenticates, and returns a live client.
func connect() (gofi.Client, error) {
	cfg, err := conn.ResolveTargetConfig(os.Stderr, opts.resolvedTarget, opts.host, opts.port, opts.site, opts.secure,
		func() (string, error) {
			prompt := "password: "
			if opts.resolvedTarget != nil && opts.resolvedTarget.Mode() == "connector" {
				prompt = "API key: "
			}
			return conn.ReadSecret(prompt, "")
		})
	if err != nil {
		return nil, err
	}

	client, err := gofi.New(cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}

// explain is a pass-through today; kept as the single seam every command
// routes errors through, matching gogl's convention, so a future
// backend-specific error-annotation pass has one place to land.
func explain(err error) error { return err }

func asJSON() bool { return opts.output == outputJSON }

// errRefused marks an error as a guard refusal (C-GLOBAL-012, exit 3) rather
// than a failure (exit 1).
var errRefused = errors.New("refused")

// errUsage marks an error caused by how the command was invoked (exit 2)
// rather than by controller/site state.
var errUsage = errors.New("usage")
```

- [ ] **Step 3: Write `main.go`**

```go
// Command gofi manages a UniFi controller's networks, DNS, reservations,
// and clients. One binary, `gofi <area> <action>`, replacing gofips,
// gofimac, gofinet, gofidns, and gofiuser (C-GLOBAL-001). See VISION.md and
// REQUIREMENTS.md for the command tree and its rationale.
package main

import (
	"errors"
	"fmt"
	"os"
)

const (
	exitOK      = 0
	exitError   = 1
	exitUsage   = 2
	exitRefused = 3
)

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gofi:", err)
		os.Exit(codeFor(err))
	}
}

// codeFor maps an error to an exit code, per C-GLOBAL-012.
func codeFor(err error) int {
	switch {
	case errors.Is(err, errRefused):
		return exitRefused
	case errors.Is(err, errUsage):
		return exitUsage
	default:
		return exitError
	}
}
```

- [ ] **Step 4: Build and smoke-test**

Run: `go build -o bin/gofi ./utilities/gofi && ./bin/gofi --help`
Expected: prints usage with `config` as the only subcommand and the six global flags (`--target`, `-H/--host`, `-p/--port`, `-S/--site`, `--secure`, `--output`).

- [ ] **Step 5: Commit**

```bash
git add utilities/gofi
git commit -m "feat: add the gofi binary shell -- root command, global flags, exit codes"
```

### Task 0.6: `gofi config` area (`show`, `targets`, `init`)

**Files:**
- Create: `utilities/gofi/config.go`
- Test: `utilities/gofi/cli_test.go` (new file; end-to-end command tests, following `gogl`'s pattern of testing cobra commands directly rather than shelling out)

**Interfaces:**
- Consumes: `opts globals`, `config.File`, `config.Path` (Task 0.2), `writeJSON` (Task 0.5).
- Produces: `newConfigCommand() *cobra.Command`, referenced by `root.go`'s `newRootCommand`.

- [ ] **Step 1: Write the failing test**

```go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInit_writesStarterFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOFI_CONFIG", filepath.Join(dir, "config.toml"))

	cmd := newConfigCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("expected config.toml to exist: %v", err)
	}
}

func TestConfigInit_refusesToOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	t.Setenv("GOFI_CONFIG", path)
	if err := os.WriteFile(path, []byte("default = \"x\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newConfigCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("config init: error = nil, want a refusal for an existing file")
	}
}

func TestConfigTargets_listsConfiguredNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	t.Setenv("GOFI_CONFIG", path)
	if err := os.WriteFile(path, []byte("[targets.home]\nhost = \"192.168.1.1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := newConfigCommand()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"targets"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config targets: %v", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/gofi/... -run TestConfig -v`
Expected: FAIL — `newConfigCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/config"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "gofi's own configuration file",
		Long: `gofi's own configuration file.

Acts on your machine, never on a controller. Holds named targets and an
output preference. Secrets never appear here (C-CONFIG's invariant): a
password or API key comes from the environment, a command the file names,
or an interactive prompt.`,
	}
	cmd.AddCommand(newConfigShowCommand(), newConfigTargetsCommand(), newConfigInitCommand())
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Report the config file's location and what it resolves to",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			exists := "yes"
			if _, err := os.Stat(opts.file.Path()); err != nil {
				exists = "no (flags and the environment still work)"
			}

			if asJSON() {
				out := map[string]any{
					"path":   opts.file.Path(),
					"exists": exists == "yes",
					"output": opts.output,
				}
				if opts.resolvedTarget != nil {
					out["target"] = opts.resolvedTarget.Name()
					out["mode"] = opts.resolvedTarget.Mode()
				}
				return writeJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "PATH       %s\nEXISTS     %s\nOUTPUT     %s\n",
				opts.file.Path(), exists, opts.output)
			if opts.resolvedTarget != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "TARGET     %s (%s mode)\n",
					opts.resolvedTarget.Name(), opts.resolvedTarget.Mode())
			}
			return nil
		},
	}
}

func newConfigTargetsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List the configured targets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			names := opts.file.Names()
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), names)
			}
			if len(names) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no targets configured in %s\n", opts.file.Path())
				fmt.Fprintln(cmd.OutOrStdout(), "run `gofi config init` to write a starting point")
				return nil
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tMODE")
			for _, name := range names {
				t := opts.file.Targets[name]
				marker := name
				if name == opts.file.Default {
					marker += " (default)"
				}
				fmt.Fprintf(tw, "%s\t%s\n", marker, t.Mode())
			}
			return tw.Flush()
		},
	}
}

func newConfigInitCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starting configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.Path()
			if _, err := os.Stat(path); err == nil && !force {
				return fmt.Errorf("%w: %s already exists; pass --force to overwrite", errUsage, path)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(starterConfig), 0o600); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
			fmt.Fprintln(cmd.OutOrStdout(), "edit it, then run `gofi config show`")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")
	return cmd
}

const starterConfig = `# gofi configuration. See ` + "`gofi config show`" + ` for where this was read from.
#
# Secrets never belong here. A password or API key comes from the environment,
# from a command named below, or from a prompt -- in that order.

# Which target to use when --target is not given. Optional with only one defined.
default = "home"

# Output format for every command: "text" or "json". Overridden by --output.
output = "text"

# Local mode: connect directly to a controller.
[targets.home]
host             = "192.168.1.1"
site             = "default"
# password_command = "pass show unifi/home"

# Connector mode: connect through api.ui.com. Uncomment and fill in console_id.
# [targets.cloud]
# site             = "default"
# console_id       = "your-console-id"
# api_key_command  = "pass show unifi/cloud-key"
`
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestConfig -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/gofi/config.go utilities/gofi/cli_test.go
git commit -m "feat: add gofi config show/targets/init"
```

**Phase 0 checkpoint:** `go build ./... && go test ./...` passes; `./bin/gofi config init && ./bin/gofi config show` works end to end with no other area yet wired.

---

## Phase 1: `ips` area (supersedes `gofips`)

Closest 1:1 port — `gofips`'s `operations.go`/`parse.go`/`format.go` already isolate the business logic from `main.go`'s flag parsing; this phase moves that logic into an importable package and adds `ips clear`.

### Task 1.1: Extract business logic into `utilities/internal/ips`

**Files:**
- Create: `utilities/internal/ips/operations.go` (moved from `utilities/gofips/operations.go`)
- Create: `utilities/internal/ips/parse.go` (moved from `utilities/gofips/parse.go`)
- Create: `utilities/internal/ips/format.go` (moved from `utilities/gofips/format.go`)
- Create: `utilities/internal/ips/operations_test.go`, `parse_test.go`, `format_test.go` (moved from the corresponding `_test.go` files in `utilities/gofips/`)
- Delete: `utilities/gofips/` (entire directory, once Task 1.2 replaces its command surface — deleted at the end of Phase 7, not here; see Task 7.1)

**Interfaces:**
- Produces (unchanged signatures, now exported from package `ips` instead of living in `gofips`'s `package main`): `ips.DoGet`, `ips.DoSet`, `ips.DoAdd`, `ips.DoDel`, `ips.Parse`, `ips.ParseSingle`, `ips.Format`, `ips.HostEntry`, `ips.ParseResult`, `ips.SetResult`, `ips.DeleteIdentifier`, `ips.FormatOptions`.

- [ ] **Step 1: Move the files and rename the package**

```bash
cd /home/gherlein/src/er/gofi
mkdir -p utilities/internal/ips
git mv utilities/gofips/operations.go utilities/internal/ips/operations.go
git mv utilities/gofips/operations_test.go utilities/internal/ips/operations_test.go
git mv utilities/gofips/parse.go utilities/internal/ips/parse.go
git mv utilities/gofips/parse_test.go utilities/internal/ips/parse_test.go
git mv utilities/gofips/format.go utilities/internal/ips/format.go
git mv utilities/gofips/format_test.go utilities/internal/ips/format_test.go
sed -i 's/^package main$/package ips/' utilities/internal/ips/*.go
```

- [ ] **Step 2: Fix the import of `utilities/internal/conn`'s `EnvDNSDomain`**

`operations.go` doesn't reference `conn` directly (confirmed: it takes `dnsDomainOverride` as a parameter, resolved by the caller). No import changes needed inside `operations.go`/`parse.go`/`format.go` themselves — they only import `context`, `io`, `net`, `sort`, `strings`, `time`, the `gofi` root package, and `gofi/src/types`, none of which changed location. Confirm this by building:

Run: `go build ./utilities/internal/ips/...`
Expected: fails only on the test files still referencing package-`main`-only helpers, if any — inspect the error and move any such helper (e.g. test fixtures) into the same package alongside its test.

- [ ] **Step 3: Run the moved tests**

Run: `go test ./utilities/internal/ips/... -v`
Expected: PASS — every test that passed under `gofips` passes identically under `ips`, since only the package name changed.

- [ ] **Step 4: Commit**

```bash
git add utilities/internal/ips
git commit -m "refactor: extract gofips business logic into utilities/internal/ips"
```

### Task 1.2: Wire `gofi ips list/export/import/add/rm`

**Files:**
- Create: `utilities/gofi/ips.go`
- Test: extend `utilities/gofi/cli_test.go`

**Interfaces:**
- Consumes: `ips.DoGet`, `ips.DoSet`, `ips.DoAdd`, `ips.DoDel`, `ips.Parse`, `ips.ParseSingle`, `ips.HostEntry`, `ips.DeleteIdentifier`, `ips.FormatOptions` (Task 1.1); `connect`, `explain`, `asJSON`, `writeJSON`, `errUsage`, `errRefused` (Task 0.5).
- Produces: `newIPsCommand() *cobra.Command`, added to `root.go`'s `newRootCommand`.

- [ ] **Step 1: Write the failing test**

```go
func TestIPsRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newIPsCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("ips rm with no identifier: error = nil, want a usage error")
	}
}

func TestIPsAdd_rejectsFlagsAndPositionalTogether(t *testing.T) {
	cmd := newIPsCommand()
	cmd.SetArgs([]string{"add", "--name", "x", "--mac", "aa:bb:cc:dd:ee:ff", "--ip", "10.0.0.1",
		"host x { hardware ethernet aa:bb:cc:dd:ee:ff; fixed-address 10.0.0.1; }"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("ips add with both flags and a positional declaration: error = nil, want a usage error")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/gofi/... -run TestIPs -v`
Expected: FAIL — `newIPsCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/ips"
)

func newIPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ips",
		Short: "Fixed IP + DNS reservations, ISC DHCP host-declaration format",
		Long: `Fixed IP + DNS reservations.

Each host declaration is a paired write: a static bind on the user record,
and a DNS A record for the hostname. These commands keep the two in step.`,
	}
	cmd.AddCommand(
		newIPsListCommand(),
		newIPsExportCommand(),
		newIPsImportCommand(),
		newIPsAddCommand(),
		newIPsRmCommand(),
		newIPsClearCommand(),
	)
	return cmd
}

func newIPsListCommand() *cobra.Command {
	return newIPsExportCommand() // list and export are the same read, per C-IPS's shared behavior
}

func newIPsExportCommand() *cobra.Command {
	var dnsDomain string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write every fixed-IP assignment to stdout in ISC DHCP format",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			opts := ips.FormatOptions{Host: opts.host, Date: time.Now().Format("2006-01-02")}
			return explain(ips.DoGet(cmd.Context(), client, opts.site(), dnsDomain, cmd.OutOrStdout(), opts))
		},
	}
	cmd.Flags().StringVar(&dnsDomain, "dns-domain", os.Getenv("UNIFI_DNS_DOMAIN"), "DNS suffix for record keys")
	return cmd
}

func newIPsImportCommand() *cobra.Command {
	var dnsDomain string
	var dryRun, force bool
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import host declarations from a file, or from stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var reader io.Reader = os.Stdin
			if len(args) == 1 && args[0] != "-" {
				f, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("open %s: %w", args[0], err)
				}
				defer f.Close()
				reader = f
			}
			parsed, err := ips.Parse(reader)
			if err != nil {
				return fmt.Errorf("%w: %s", errUsage, err)
			}
			if len(parsed.Entries) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "no entries to process")
				return nil
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			result, err := ips.DoSet(cmd.Context(), client, siteFlag(), parsed.Entries, dnsDomain, dryRun, force)
			if err != nil {
				return explain(err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%d processed, %d skipped, %d created, %d updated, %d errors\n",
				len(parsed.Entries), result.Skipped, result.Created, result.Updated, result.Errors)
			if result.Errors > 0 {
				return fmt.Errorf("import completed with %d error(s)", result.Errors)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&dnsDomain, "dns-domain", os.Getenv("UNIFI_DNS_DOMAIN"), "DNS suffix for record keys")
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	f.BoolVar(&force, "force", false, "reprocess entries even if unchanged")
	return cmd
}

func newIPsAddCommand() *cobra.Command {
	var name, mac, ip, dnsDomain string
	var force bool
	cmd := &cobra.Command{
		Use:   "add [declaration]",
		Short: "Add one host, by flags or as an ISC DHCP declaration fragment",
		Example: `  gofi ips add --name nas --mac aa:bb:cc:dd:ee:01 --ip 192.168.1.13
  gofi ips add 'host nas { hardware ethernet aa:bb:cc:dd:ee:01; fixed-address 192.168.1.13; }'`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			byFlags := name != "" || mac != "" || ip != ""
			if len(args) == 1 && byFlags {
				return fmt.Errorf("%w: give either a declaration or --name/--mac/--ip, not both", errUsage)
			}

			var entry *ips.HostEntry
			switch {
			case len(args) == 1:
				parsed, err := ips.ParseSingle(args[0])
				if err != nil {
					return fmt.Errorf("%w: %s", errUsage, err)
				}
				entry = parsed
			case byFlags:
				if name == "" || mac == "" || ip == "" {
					return fmt.Errorf("%w: --name, --mac and --ip are required together", errUsage)
				}
				entry = &ips.HostEntry{Hostname: name, MAC: mac, IP: ip}
			default:
				parsed, err := ips.Parse(os.Stdin)
				if err != nil {
					return fmt.Errorf("%w: %s", errUsage, err)
				}
				if len(parsed.Entries) != 1 {
					return fmt.Errorf("%w: expected exactly one host declaration on stdin, found %d", errUsage, len(parsed.Entries))
				}
				entry = &parsed.Entries[0]
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			return explain(ips.DoAdd(cmd.Context(), client, siteFlag(), entry, dnsDomain, force))
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "hostname")
	f.StringVar(&mac, "mac", "", "MAC address")
	f.StringVar(&ip, "ip", "", "IPv4 address")
	f.StringVar(&dnsDomain, "dns-domain", os.Getenv("UNIFI_DNS_DOMAIN"), "DNS suffix for record keys")
	f.BoolVar(&force, "force", false, "proceed past conflicts")
	return cmd
}

func newIPsRmCommand() *cobra.Command {
	var name, mac, ip, dnsDomain string
	var force, keepDNS, dryRun bool
	cmd := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"remove"},
		Short:   "Remove one host, identified by --name, --mac or --ip",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			count := 0
			for _, v := range []string{name, mac, ip} {
				if v != "" {
					count++
				}
			}
			if count != 1 {
				return fmt.Errorf("%w: pass exactly one of --name, --mac, or --ip", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			identifier := ips.DeleteIdentifier{Name: name, MAC: mac, IP: ip}
			return explain(ips.DoDel(cmd.Context(), client, siteFlag(), identifier, dnsDomain, force, keepDNS))
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "hostname to remove")
	f.StringVar(&mac, "mac", "", "MAC address to remove")
	f.StringVar(&ip, "ip", "", "IPv4 address to remove")
	f.StringVar(&dnsDomain, "dns-domain", os.Getenv("UNIFI_DNS_DOMAIN"), "DNS suffix for record keys")
	f.BoolVar(&force, "force", false, "proceed past an ambiguous match")
	f.BoolVar(&keepDNS, "keep-dns", false, "do not delete the associated DNS record")
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	return cmd
}

// siteFlag reads the resolved site the same way every ips command needs it;
// area command files added in later phases follow the same pattern.
func siteFlag() string {
	if opts.site != "" {
		return opts.site
	}
	if opts.resolvedTarget != nil && opts.resolvedTarget.Site != "" {
		return opts.resolvedTarget.Site
	}
	return "default"
}

func (o FormatOptionsHost) site() string { return siteFlag() }

type FormatOptionsHost = ips.FormatOptions
```

- [ ] **Step 4: Fix the small wiring mismatch introduced above**

The `newIPsExportCommand` body above calls `opts.site()`, which doesn't exist on `ips.FormatOptions` — replace that line with a direct call to `siteFlag()`:

```go
return explain(ips.DoGet(cmd.Context(), client, siteFlag(), dnsDomain, cmd.OutOrStdout(), opts))
```

and delete the `func (o FormatOptionsHost) site() string` and `type FormatOptionsHost` lines added above — they were scaffolding for a wrong approach and aren't needed once `siteFlag()` is called directly.

- [ ] **Step 5: Wire `ips` into the root command**

In `utilities/gofi/root.go`, replace the comment-only line in `newRootCommand`'s `root.AddCommand(...)` call:

```go
root.AddCommand(
	newConfigCommand(),
	newIPsCommand(),
)
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestIPs -v`
Expected: PASS.

- [ ] **Step 7: Build and smoke-test against the mock server or a real controller**

Run: `go build -o bin/gofi ./utilities/gofi && ./bin/gofi ips --help`
Expected: lists `list`, `export`, `import`, `add`, `rm`, `clear` (the last added in Task 1.3).

- [ ] **Step 8: Commit**

```bash
git add utilities/gofi/ips.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi ips list/export/import/add/rm"
```

### Task 1.3: `ips.DoClear` — new business logic for `ips clear`

**Files:**
- Modify: `utilities/internal/ips/operations.go` (add `DoClear`)
- Modify: `utilities/internal/ips/operations_test.go`

**Interfaces:**
- Produces: `ips.DoClear(ctx context.Context, client gofi.Client, site string, dryRun bool) ([]ips.HostEntry, error)`. Returns every entry that was (or, if `dryRun`, would be) removed — the caller (Task 1.4) uses this list to satisfy C-IPS-008 (print before confirming).

- [ ] **Step 1: Write the failing test**

```go
func TestDoClear_removesEveryFixedIPAssignment(t *testing.T) {
	client := newMockClientWithFixedIPs(t, []HostEntry{
		{Hostname: "a", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
		{Hostname: "b", MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.11"},
	})

	removed, err := DoClear(context.Background(), client, "default", false)
	if err != nil {
		t.Fatalf("DoClear() error = %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed = %d entries, want 2", len(removed))
	}

	remaining, err := DoGetEntries(context.Background(), client, "default", "")
	if err != nil {
		t.Fatalf("DoGetEntries() error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("remaining = %d entries after clear, want 0", len(remaining))
	}
}

func TestDoClear_dryRunRemovesNothing(t *testing.T) {
	client := newMockClientWithFixedIPs(t, []HostEntry{
		{Hostname: "a", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	})

	removed, err := DoClear(context.Background(), client, "default", true)
	if err != nil {
		t.Fatalf("DoClear() error = %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("removed = %d entries (preview), want 1", len(removed))
	}

	remaining, err := DoGetEntries(context.Background(), client, "default", "")
	if err != nil {
		t.Fatalf("DoGetEntries() error = %v", err)
	}
	if len(remaining) != 1 {
		t.Errorf("remaining = %d entries after dry-run clear, want 1 (nothing removed)", len(remaining))
	}
}
```

Note: `newMockClientWithFixedIPs` and `DoGetEntries` are test/implementation helpers this task also needs. `DoGet` today writes directly to an `io.Writer`; extracting the entry-listing step it already performs internally into a small exported `DoGetEntries(ctx, client, site, dnsDomainOverride string) ([]HostEntry, error)` helper — with `DoGet` calling it and then formatting — is part of this task's implementation, not scope creep: `DoClear` needs the same listing `DoGet` already builds internally.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/ips/... -run TestDoClear -v`
Expected: FAIL — `DoClear` and `DoGetEntries` undefined.

- [ ] **Step 3: Refactor `DoGet` to expose `DoGetEntries`, then add `DoClear`**

In `operations.go`, find the existing `DoGet` (already read: it lists users with `UseFixedIP == true`, resolves hostnames, cross-references DNS). Extract its entry-building loop into `DoGetEntries`, then have `DoGet` call it:

```go
// DoGetEntries returns every fixed-IP assignment as HostEntry values, sorted
// numerically by IP -- the same listing DoGet formats to a writer, exposed
// so DoClear (and any other caller) doesn't need a writer to get the data.
func DoGetEntries(ctx context.Context, client gofi.Client, site, dnsDomainOverride string) ([]HostEntry, error) {
	// Move DoGet's existing entry-collection body here verbatim (steps 2-5 of
	// DoGet's current implementation: list users, filter UseFixedIP, resolve
	// hostname, cross-reference DNS, sort by IP). DoGet's step 6 (writing the
	// formatted output) stays in DoGet.
	...
}

func DoGet(ctx context.Context, client gofi.Client, site, dnsDomainOverride string, writer io.Writer, options FormatOptions) error {
	entries, err := DoGetEntries(ctx, client, site, dnsDomainOverride)
	if err != nil {
		return err
	}
	return Format(writer, entries, options)
}

// DoClear removes every fixed-IP assignment on the site, per host, using the
// same removal DoDel performs for one (C-IPS-007, C-IPS-009: the DNS record
// goes too, unless a future --keep-dns is threaded through here). Returns
// what was (or, if dryRun, would be) removed, so the caller can print it
// before a confirmation prompt (C-IPS-008).
func DoClear(ctx context.Context, client gofi.Client, site string, dryRun bool) ([]HostEntry, error) {
	entries, err := DoGetEntries(ctx, client, site, "")
	if err != nil {
		return nil, err
	}
	if dryRun {
		return entries, nil
	}
	for _, entry := range entries {
		identifier := DeleteIdentifier{MAC: entry.MAC}
		if err := DoDel(ctx, client, site, identifier, "", true, false); err != nil {
			return nil, fmt.Errorf("clearing %s (%s): %w", entry.Hostname, entry.MAC, err)
		}
	}
	return entries, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/ips/... -v`
Expected: PASS, including every pre-existing test (the `DoGet` refactor must not change its externally observable behavior — this is verified by the pre-existing `DoGet`/`Format` tests still passing unmodified).

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/ips
git commit -m "feat: add ips.DoClear, refactoring DoGet to expose DoGetEntries"
```

### Task 1.4: Wire `gofi ips clear` — hard-gated per C-IPS-007/008

**Files:**
- Modify: `utilities/gofi/ips.go`
- Test: extend `utilities/gofi/cli_test.go`

**Interfaces:**
- Consumes: `ips.DoClear` (Task 1.3).

- [ ] **Step 1: Write the failing test**

```go
func TestIPsClear_refusesWithoutForce(t *testing.T) {
	cmd := newIPsCommand()
	cmd.SetArgs([]string{"clear"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("ips clear without --force: error = nil, want a refusal")
	}
	if !errors.Is(err, errUsage) && !errors.Is(err, errRefused) {
		t.Errorf("ips clear without --force: error = %v, want errUsage or errRefused", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestIPsClear -v`
Expected: FAIL — `ips clear` subcommand doesn't exist.

- [ ] **Step 3: Write the implementation**

```go
func newIPsClearCommand() *cobra.Command {
	var force, yes, dryRun bool
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove every fixed-IP assignment on the site (--force required)",
		Long: `Remove every fixed-IP assignment on the site.

Hard-gated beyond a normal write: --force is mandatory (there is no bare
invocation that deletes everything), the full list of what would be removed
is printed before the confirmation prompt, and --yes only skips that prompt
-- it never substitutes for --force. gofi manages shared infrastructure, so
this verb needs a stronger floor than a disposable bench router's equivalent
(C-IPS-007, C-IPS-008).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !force {
				return fmt.Errorf("%w: --force is required; this removes every fixed-IP assignment on the site", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			preview, err := ips.DoClear(cmd.Context(), client, siteFlag(), true)
			if err != nil {
				return explain(err)
			}
			if len(preview) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no fixed-IP assignments to remove")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "This will remove %d fixed-IP assignment(s):\n", len(preview))
			for _, entry := range preview {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\t%s\t%s\n", entry.Hostname, entry.MAC, entry.IP)
			}

			if dryRun {
				return nil
			}
			if !yes {
				fmt.Fprint(cmd.OutOrStdout(), "Proceed? [y/N] ")
				var response string
				fmt.Fscanln(cmd.InOrStdin(), &response)
				if strings.ToLower(strings.TrimSpace(response)) != "y" {
					return fmt.Errorf("%w: not confirmed", errRefused)
				}
			}

			removed, err := ips.DoClear(cmd.Context(), client, siteFlag(), false)
			if err != nil {
				return explain(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d fixed-IP assignment(s)\n", len(removed))
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&force, "force", false, "required: acknowledges this removes every fixed-IP assignment")
	f.BoolVar(&yes, "yes", false, "skip the confirmation prompt (still requires --force)")
	f.BoolVar(&dryRun, "dry-run", false, "show what would be removed without removing it")
	return cmd
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestIPsClear -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/gofi/ips.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi ips clear, hard-gated behind --force and a printed preview"
```

**Phase 1 checkpoint:** `go test ./utilities/internal/ips/... ./utilities/gofi/...` passes; `gofi ips list/export/import/add/rm/clear --help` all show correct usage; C-IPS-001 through C-IPS-011 are all satisfied (cross-check against `REQUIREMENTS.md`'s `IPS` section).

---

## Phase 2: `dns` area (supersedes `gofidns`)

Same extraction pattern as Phase 1, smaller surface (`list`, `rm` only — `add`/`set` stay Blocked per C-DNS-004).

### Task 2.1: Extract business logic into `utilities/internal/dns`

**Files:**
- Create: `utilities/internal/dns/operations.go`, `format.go` (moved from `utilities/gofidns/`)
- Create: corresponding `_test.go` files (moved)

**Interfaces:**
- Produces: `dns.ListRecords`, `dns.DoGet`, `dns.DoDel`, `dns.DNSEntry`, `dns.DeleteIdentifier`, `dns.DeleteResult`, `dns.FormatOptions`, `dns.FormatText`, `dns.FormatJSON`.

- [ ] **Step 1: Move and rename**

```bash
mkdir -p utilities/internal/dns
git mv utilities/gofidns/operations.go utilities/internal/dns/operations.go
git mv utilities/gofidns/operations_test.go utilities/internal/dns/operations_test.go
git mv utilities/gofidns/format.go utilities/internal/dns/format.go
git mv utilities/gofidns/format_test.go utilities/internal/dns/format_test.go
sed -i 's/^package main$/package dns/' utilities/internal/dns/*.go
```

- [ ] **Step 2: Build and run the moved tests**

Run: `go build ./utilities/internal/dns/... && go test ./utilities/internal/dns/... -v`
Expected: PASS, unmodified behavior.

- [ ] **Step 3: Commit**

```bash
git add utilities/internal/dns
git commit -m "refactor: extract gofidns business logic into utilities/internal/dns"
```

### Task 2.2: Wire `gofi dns list/rm`, exit code 3 for ambiguous matches

**Files:**
- Create: `utilities/gofi/dns.go`
- Modify: `utilities/gofi/root.go` (add `newDNSCommand()` to `root.AddCommand`)
- Test: extend `utilities/gofi/cli_test.go`

**Interfaces:**
- Consumes: `dns.DoGet`, `dns.DoDel`, `dns.DeleteIdentifier`, `dns.FormatOptions` (Task 2.1); `connect`, `explain`, `asJSON`, `siteFlag`, `errUsage`, `errRefused` (Phase 0/1).

- [ ] **Step 1: Write the failing test**

```go
func TestDNSRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newDNSCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("dns rm with no identifier: error = nil, want a usage error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestDNSRm -v`
Expected: FAIL — `newDNSCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/dns"
)

func newDNSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Local DNS records, independent of ips",
		Long: `Local DNS records, independent of fixed-IP reservations.

Exists for the case where only the DNS side of a binding should move --
correcting a stale record without touching a reservation something else
still depends on. See VISION.md's Area boundaries for the ips/dns split.`,
	}
	cmd.AddCommand(newDNSListCommand(), newDNSRmCommand())
	return cmd
}

func newDNSListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List DNS records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			return explain(dns.DoGet(cmd.Context(), client, siteFlag(),
				dns.FormatOptions{Writer: cmd.OutOrStdout(), JSON: asJSON()}))
		},
	}
}

func newDNSRmCommand() *cobra.Command {
	var id, name, ip string
	var force, dryRun bool
	cmd := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"remove"},
		Short:   "Remove one record, by --id/--name/--ip",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			count := 0
			for _, v := range []string{id, name, ip} {
				if v != "" {
					count++
				}
			}
			if count != 1 {
				return fmt.Errorf("%w: pass exactly one of --id, --name, or --ip", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			identifier := dns.DeleteIdentifier{ID: id, Name: name, IP: ip}
			result, err := dns.DoDel(cmd.Context(), client, siteFlag(), identifier, dryRun, force, cmd.ErrOrStderr())
			if err != nil {
				// An ambiguous, unforced match is a guard refusal (C-DNS-003):
				// exit 3, not 1.
				return fmt.Errorf("%w: %s", errRefused, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d record(s)\n", result.Removed)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&id, "id", "", "record ID to remove")
	f.StringVar(&name, "name", "", "record name to remove")
	f.StringVar(&ip, "ip", "", "record value (IP) to remove")
	f.BoolVar(&force, "force", false, "allow an identifier matching several records to remove them all")
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	return cmd
}
```

Note: `dns.DoDel`'s existing behavior returns a plain `error` for both "not found" (should map to exit 1) and "ambiguous without --force" (should map to exit 3, per C-DNS-003/B-DNS-002). Wrapping every `DoDel` error in `errRefused` above is wrong for the not-found case — fix in the next step.

- [ ] **Step 4: Disambiguate "not found" from "ambiguous" in `dns.DoDel`'s error**

In `utilities/internal/dns/operations.go`, `DoDel` currently returns a bare `fmt.Errorf(...)` for both cases. Add two sentinel errors and use them:

```go
var (
	// ErrNotFound means the identifier matched nothing (exit 1: a real error).
	ErrNotFound = errors.New("no matching record")
	// ErrAmbiguous means the identifier matched more than one record without
	// --force (exit 3: a guard refusal, C-DNS-003).
	ErrAmbiguous = errors.New("ambiguous match")
)
```

Update `DoDel`'s zero-match branch to wrap `ErrNotFound` and its multi-match-without-force branch to wrap `ErrAmbiguous`, then update `newDNSRmCommand`'s `RunE` in `utilities/gofi/dns.go`:

```go
result, err := dns.DoDel(cmd.Context(), client, siteFlag(), identifier, dryRun, force, cmd.ErrOrStderr())
if errors.Is(err, dns.ErrAmbiguous) {
	return fmt.Errorf("%w: %s", errRefused, err)
}
if err != nil {
	return explain(err)
}
```

- [ ] **Step 5: Add tests for the sentinel-error distinction**

```go
func TestDoDel_notFoundIsErrNotFound(t *testing.T) {
	client := newMockClientWithNoDNSRecords(t)
	_, err := DoDel(context.Background(), client, "default", DeleteIdentifier{ID: "missing"}, false, false, io.Discard)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("DoDel() error = %v, want ErrNotFound", err)
	}
}

func TestDoDel_ambiguousIsErrAmbiguous(t *testing.T) {
	client := newMockClientWithDNSRecords(t, twoRecordsNamedSame())
	_, err := DoDel(context.Background(), client, "default", DeleteIdentifier{Name: "dup"}, false, false, io.Discard)
	if !errors.Is(err, ErrAmbiguous) {
		t.Errorf("DoDel() error = %v, want ErrAmbiguous", err)
	}
}
```

- [ ] **Step 6: Wire `dns` into the root command**

In `root.go`: `root.AddCommand(newConfigCommand(), newIPsCommand(), newDNSCommand())`.

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./utilities/internal/dns/... ./utilities/gofi/... -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add utilities/internal/dns utilities/gofi/dns.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi dns list/rm with exit code 3 for ambiguous matches"
```

**Phase 2 checkpoint:** C-DNS-001 through C-DNS-003 satisfied; C-DNS-004 (`dns add`/`dns set`) remains explicitly Blocked — do not implement it in this phase.

---

## Phase 3: `network` area (supersedes `gofinet`)

Read-only today (C-NETWORK-004 is Blocked). Adds `network show <name>`, new relative to `gofinet`.

### Task 3.1: Extract business logic into `utilities/internal/network`

**Files:**
- Create: `utilities/internal/network/operations.go`, `format.go` (moved from `utilities/gofinet/`)
- Create: corresponding `_test.go` files (moved)

**Interfaces:**
- Produces: `network.ListNetworks`, `network.NetworkEntry`, `network.FormatText`, `network.FormatJSON`.

- [ ] **Step 1: Move and rename**

```bash
mkdir -p utilities/internal/network
git mv utilities/gofinet/operations.go utilities/internal/network/operations.go
git mv utilities/gofinet/operations_test.go utilities/internal/network/operations_test.go
git mv utilities/gofinet/format.go utilities/internal/network/format.go
git mv utilities/gofinet/format_test.go utilities/internal/network/format_test.go
sed -i 's/^package main$/package network/' utilities/internal/network/*.go
```

- [ ] **Step 2: Build and run the moved tests**

Run: `go build ./utilities/internal/network/... && go test ./utilities/internal/network/... -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add utilities/internal/network
git commit -m "refactor: extract gofinet business logic into utilities/internal/network"
```

### Task 3.2: Add `network.FindByName` for `network show`

**Files:**
- Modify: `utilities/internal/network/operations.go`
- Modify: `utilities/internal/network/operations_test.go`

**Interfaces:**
- Produces: `network.FindByName(ctx context.Context, client gofi.Client, site, name string) (*network.NetworkEntry, error)`, returning `network.ErrNotFound` (new sentinel) when no network matches.

- [ ] **Step 1: Write the failing test**

```go
func TestFindByName_returnsErrNotFoundForUnknownNetwork(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Default"}})
	_, err := FindByName(context.Background(), client, "default", "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByName() error = %v, want ErrNotFound", err)
	}
}

func TestFindByName_findsExactMatch(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Guest"}, {Name: "Default"}})
	entry, err := FindByName(context.Background(), client, "default", "Guest")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if entry.Name != "Guest" {
		t.Errorf("Name = %q, want Guest", entry.Name)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/network/... -run TestFindByName -v`
Expected: FAIL — `FindByName` undefined.

- [ ] **Step 3: Write the implementation**

```go
// ErrNotFound means no network on the site matched the given name (C-NETWORK-003).
var ErrNotFound = errors.New("no matching network")

// FindByName reports one network's full detail by exact name match.
func FindByName(ctx context.Context, client gofi.Client, site, name string) (*NetworkEntry, error) {
	entries, err := ListNetworks(ctx, client, site)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		if entries[i].Name == name {
			return &entries[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/network/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/network
git commit -m "feat: add network.FindByName for gofi network show"
```

### Task 3.3: Wire `gofi network list/show`

**Files:**
- Create: `utilities/gofi/network.go`
- Modify: `utilities/gofi/root.go`
- Test: extend `utilities/gofi/cli_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestNetworkShow_requiresOneArg(t *testing.T) {
	cmd := newNetworkCommand()
	cmd.SetArgs([]string{"show"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("network show with no name: error = nil, want a usage error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestNetworkShow -v`
Expected: FAIL — `newNetworkCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/network"
)

func newNetworkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Networks (VLANs): subnet, DHCP pool, DNS servers",
		Long: `Networks on the site: subnet, VLAN tag, DHCP pool boundaries, DNS servers.

Read-only today (C-NETWORK-004): a write endpoint has not yet been verified
against real hardware.`,
	}
	cmd.AddCommand(newNetworkListCommand(), newNetworkShowCommand())
	return cmd
}

func newNetworkListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every network on the site",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entries, err := network.ListNetworks(cmd.Context(), client, siteFlag())
			if err != nil {
				return explain(err)
			}
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), entries)
			}
			return network.FormatText(cmd.OutOrStdout(), entries)
		},
	}
}

func newNetworkShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Report one network in detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entry, err := network.FindByName(cmd.Context(), client, siteFlag(), args[0])
			if errors.Is(err, network.ErrNotFound) {
				return err
			}
			if err != nil {
				return explain(err)
			}
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), entry)
			}
			return network.FormatText(cmd.OutOrStdout(), []network.NetworkEntry{*entry})
		},
	}
}
```

- [ ] **Step 4: Wire `network` into the root command**

`root.AddCommand(newConfigCommand(), newIPsCommand(), newDNSCommand(), newNetworkCommand())`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestNetwork -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add utilities/gofi/network.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi network list/show"
```

**Phase 3 checkpoint:** C-NETWORK-001 through C-NETWORK-003 satisfied; C-NETWORK-004 remains explicitly Blocked.

---

## Phase 4: `clients` area (supersedes `gofimac`)

Adds `clients vendor <mac>` as a standalone command (new relative to `gofimac`, which only did OUI lookup inline within `list`) and moves the OUI cache path.

### Task 4.1: Extract business logic into `utilities/internal/clients`

**Files:**
- Create: `utilities/internal/clients/operations.go`, `format.go`, `oui.go` (moved from `utilities/gofimac/`)
- Create: corresponding `_test.go` files (moved)

**Interfaces:**
- Produces: `clients.ListClients`, `clients.FindClientByMAC`, `clients.ClientEntry`, `clients.OUIDatabase`, `clients.FormatText`, `clients.FormatJSON`, plus the filter/sort types.

- [ ] **Step 1: Move and rename**

```bash
mkdir -p utilities/internal/clients
git mv utilities/gofimac/operations.go utilities/internal/clients/operations.go
git mv utilities/gofimac/operations_test.go utilities/internal/clients/operations_test.go
git mv utilities/gofimac/format.go utilities/internal/clients/format.go
git mv utilities/gofimac/format_test.go utilities/internal/clients/format_test.go
git mv utilities/gofimac/oui.go utilities/internal/clients/oui.go
git mv utilities/gofimac/oui_test.go utilities/internal/clients/oui_test.go
git mv utilities/gofimac/duration.go utilities/internal/clients/duration.go
git mv utilities/gofimac/duration_test.go utilities/internal/clients/duration_test.go
sed -i 's/^package main$/package clients/' utilities/internal/clients/*.go
```

- [ ] **Step 2: Move the OUI cache directory (C-CLIENTS-003)**

`oui.go`'s cache-location logic (per `CLAUDE.md`: `$XDG_DATA_HOME/gofimac/oui.txt`) needs its hardcoded `"gofimac"` path segment changed to `"gofi"`. Find that string:

Run: `grep -n '"gofimac"' utilities/internal/clients/oui.go`

Replace the matched literal with `"gofi"`.

- [ ] **Step 3: Build and run the moved tests**

Run: `go build ./utilities/internal/clients/... && go test ./utilities/internal/clients/... -v`
Expected: PASS, except any test asserting the literal `gofimac` cache path — update those to expect `gofi`.

- [ ] **Step 4: Commit**

```bash
git add utilities/internal/clients
git commit -m "refactor: extract gofimac business logic into utilities/internal/clients, rename OUI cache to gofi"
```

### Task 4.2: Wire `gofi clients list`

**Files:**
- Create: `utilities/gofi/clients.go`
- Modify: `utilities/gofi/root.go`
- Test: extend `utilities/gofi/cli_test.go`

**Interfaces:**
- Consumes: `clients.ListClients`, `clients.FindClientByMAC`, `clients.OUIDatabase`, `clients.FormatText`, `clients.FormatJSON` (Task 4.1).

- [ ] **Step 1: Write the failing test**

```go
func TestClientsList_rejectsWifiAndWiredTogether(t *testing.T) {
	cmd := newClientsCommand()
	cmd.SetArgs([]string{"list", "--wifi", "--wired"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("clients list --wifi --wired: error = nil, want a usage error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestClientsList -v`
Expected: FAIL — `newClientsCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/clients"
	"github.com/unifi-go/gofi/utilities/internal/config"
)

func newClientsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clients",
		Short: "Currently-connected stations, with OUI lookup",
		Long: `Currently-connected stations (live telemetry, not persistent identity --
see "gofi users" for the known-client registry).

Manufacturer lookup always comes from gofi's own cached IEEE OUI database,
never the controller's own (frequently stale) OUI field.`,
	}
	cmd.AddCommand(newClientsListCommand(), newClientsVendorCommand())
	return cmd
}

func newClientsListCommand() *cobra.Command {
	var wifi, wired bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connected stations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if wifi && wired {
				return fmt.Errorf("%w: --wifi and --wired are mutually exclusive", errUsage)
			}
			filter := clients.FilterAll
			if wifi {
				filter = clients.FilterWiFi
			}
			if wired {
				filter = clients.FilterWired
			}

			cacheDir, err := ouiCacheDir()
			if err != nil {
				return err
			}
			db, err := clients.LoadOUI(cacheDir, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entries, err := clients.ListClients(cmd.Context(), client, siteFlag(), filter, clients.SortByIP, db)
			if err != nil {
				return explain(err)
			}
			if asJSON() {
				return clients.FormatJSON(cmd.OutOrStdout(), entries)
			}
			return clients.FormatText(cmd.OutOrStdout(), entries, false, 0)
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&wifi, "wifi", "w", false, "only wireless stations")
	f.BoolVarP(&wired, "wired", "e", false, "only wired stations")
	return cmd
}

// ouiCacheDir resolves $XDG_DATA_HOME/gofi (C-CLIENTS-003).
func ouiCacheDir() (string, error) {
	return config.DataDir("gofi")
}
```

- [ ] **Step 4: Add `config.DataDir`, needed by Step 3**

`utilities/internal/config` (Task 0.2) doesn't yet have an XDG data-dir helper — `oui.go`'s pre-migration version computed this inline. Add it to `utilities/internal/config/config.go`:

```go
// DataDir returns where re-downloadable data for the named subsystem lives,
// honoring XDG_DATA_HOME.
func DataDir(name string) (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locating a data directory: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, name), nil
}
```

Then, in `utilities/internal/clients/oui.go`, replace its own inline XDG-path computation (the code touched in Task 4.1 Step 2) with a call to `config.DataDir("gofi")`, removing the now-duplicate logic.

- [ ] **Step 5: Verify `clients.FilterAll`/`FilterWiFi`/`FilterWired`/`SortByIP` names match `operations.go`**

`operations.go`'s `FilterMode`/`SortMode` are `int`-based enum types (confirmed in Task 4.1's extraction). Check their actual constant names:

Run: `grep -n "FilterMode = iota\|SortMode = iota\|Filter.*=\|Sort.*=" utilities/internal/clients/operations.go`

If the constants are named differently than `FilterAll`/`FilterWiFi`/`FilterWired`/`SortByIP` above, update `clients.go`'s references to match exactly — cobra wiring must use the real names, not assumed ones.

- [ ] **Step 6: Wire `clients` into the root command**

`root.AddCommand(newConfigCommand(), newIPsCommand(), newDNSCommand(), newNetworkCommand(), newClientsCommand())`.

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./utilities/internal/config/... ./utilities/internal/clients/... ./utilities/gofi/... -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add utilities/internal/config utilities/internal/clients utilities/gofi/clients.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi clients list, sharing an XDG data-dir helper for the OUI cache"
```

### Task 4.3: `gofi clients vendor <mac>` — new standalone command

**Files:**
- Modify: `utilities/gofi/clients.go`
- Test: extend `utilities/gofi/cli_test.go`

**Interfaces:**
- Consumes: `clients.OUIDatabase.Lookup` (existing method on the type moved in Task 4.1 — confirm its exact name before use).

- [ ] **Step 1: Confirm the lookup method's name**

Run: `grep -n "func (.*OUIDatabase)" utilities/internal/clients/oui.go`

Use whatever this reports (likely `Lookup(mac string) string`, matching `gogl`'s equivalent) in the step below.

- [ ] **Step 2: Write the failing test**

```go
func TestClientsVendor_worksWithoutAControllerSession(t *testing.T) {
	cmd := newClientsCommand()
	// No --host, --target, or credentials given: this must not attempt to
	// connect (I-CLIENTS-001).
	cmd.SetArgs([]string{"vendor", "b4:0e:cf:2a:85:6f"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("clients vendor: %v (should need no controller connection)", err)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestClientsVendor -v`
Expected: FAIL — no `vendor` subcommand registered.

- [ ] **Step 4: Write the implementation**

```go
func newClientsVendorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "vendor <mac>",
		Short: "Look up a MAC address's manufacturer, without contacting the controller",
		Long: `Look up a MAC address's manufacturer.

Entirely offline (C-CLIENTS-004): reads the cached IEEE OUI registry and
never opens a controller session.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cacheDir, err := ouiCacheDir()
			if err != nil {
				return err
			}
			db, err := clients.LoadOUI(cacheDir, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			vendor := db.Lookup(args[0])
			if vendor == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  no registered manufacturer (locally administered or randomized)\n", args[0])
				return nil
			}
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), map[string]string{"mac": args[0], "vendor": vendor})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", args[0], vendor)
			return nil
		},
	}
}
```

Add `newClientsVendorCommand()` to `newClientsCommand`'s `cmd.AddCommand(...)` call (Task 4.2, Step 3).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestClientsVendor -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add utilities/gofi/clients.go utilities/gofi/cli_test.go
git commit -m "feat: add gofi clients vendor, an offline OUI lookup"
```

**Phase 4 checkpoint:** C-CLIENTS-001 through C-CLIENTS-005 satisfied.

---

## Phase 5: `users` area (supersedes `gofiuser`)

Same extraction pattern; no new commands, only exit-code updates.

### Task 5.1: Extract business logic into `utilities/internal/users`

**Files:**
- Create: `utilities/internal/users/operations.go`, `format.go` (moved from `utilities/gofiuser/`)
- Create: corresponding `_test.go` files (moved)

- [ ] **Step 1: Move and rename**

```bash
mkdir -p utilities/internal/users
git mv utilities/gofiuser/operations.go utilities/internal/users/operations.go
git mv utilities/gofiuser/operations_test.go utilities/internal/users/operations_test.go
git mv utilities/gofiuser/format.go utilities/internal/users/format.go
git mv utilities/gofiuser/format_test.go utilities/internal/users/format_test.go
sed -i 's/^package main$/package users/' utilities/internal/users/*.go
```

- [ ] **Step 2: Build and run the moved tests**

Run: `go build ./utilities/internal/users/... && go test ./utilities/internal/users/... -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add utilities/internal/users
git commit -m "refactor: extract gofiuser business logic into utilities/internal/users"
```

### Task 5.2: Sentinel errors for ambiguous match, matching Task 2.2's pattern

**Files:**
- Modify: `utilities/internal/users/operations.go`
- Modify: `utilities/internal/users/operations_test.go`

**Interfaces:**
- Produces: `users.ErrNotFound`, `users.ErrAmbiguous` (same shape as `dns`'s equivalents from Task 2.2).

- [ ] **Step 1: Write the failing tests**

```go
func TestFindUser_ambiguousNameIsErrAmbiguous(t *testing.T) {
	dup := []types.User{{Name: "phone"}, {Name: "phone"}}
	_, err := findUser(dup, DeleteIdentifier{Name: "phone"})
	if !errors.Is(err, ErrAmbiguous) {
		t.Errorf("findUser() error = %v, want ErrAmbiguous", err)
	}
}

func TestFindUser_noMatchIsErrNotFound(t *testing.T) {
	_, err := findUser(nil, DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("findUser() error = %v, want ErrNotFound", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/users/... -run "TestFindUser_ambiguous|TestFindUser_noMatch" -v`
Expected: FAIL — sentinels undefined, or `findUser`'s existing error doesn't wrap them.

- [ ] **Step 3: Add the sentinels and wrap them in `findUser`**

```go
var (
	ErrNotFound  = errors.New("no matching user")
	ErrAmbiguous = errors.New("ambiguous match")
)
```

Update `findUser`'s existing zero-match return to wrap `ErrNotFound`, and its ambiguous-name branch to wrap `ErrAmbiguous`, mirroring exactly how Task 2.2 updated `dns.DoDel`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/users/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/users
git commit -m "feat: add users.ErrNotFound/ErrAmbiguous for exit-code-3 refusals"
```

### Task 5.3: Wire `gofi users list/rm`

**Files:**
- Create: `utilities/gofi/users.go`
- Modify: `utilities/gofi/root.go`
- Test: extend `utilities/gofi/cli_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestUsersRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newUsersCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("users rm with no identifier: error = nil, want a usage error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestUsersRm -v`
Expected: FAIL — `newUsersCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/users"
)

func newUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Known-client identity records, connected or not",
		Long: `Known-client identity records: persistent, outlive the connection, carry the
fixed-IP relationship. See "gofi clients" for currently-connected stations
only -- the two read different backend objects, kept as separate areas
permanently (VISION.md's Area boundaries).`,
	}
	cmd.AddCommand(newUsersListCommand(), newUsersRmCommand())
	return cmd
}

func newUsersListCommand() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known clients",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			return explain(users.DoList(cmd.Context(), client, siteFlag(), filter,
				users.FormatOptions{Writer: cmd.OutOrStdout(), JSON: asJSON()}))
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "only list entries containing this substring")
	return cmd
}

func newUsersRmCommand() *cobra.Command {
	var mac, name string
	var dryRun bool
	cmd := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"remove"},
		Short:   "Remove a known client, by --mac/--name",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if (mac == "") == (name == "") {
				return fmt.Errorf("%w: pass exactly one of --mac or --name", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			identifier := users.DeleteIdentifier{MAC: mac, Name: name}
			result, err := users.DoDel(cmd.Context(), client, siteFlag(), identifier, dryRun, cmd.ErrOrStderr())
			if errors.Is(err, users.ErrAmbiguous) {
				return fmt.Errorf("%w: %s", errRefused, err)
			}
			if err != nil {
				return explain(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed fixed IP: %t, forgot client: %t\n",
				result.FixedIPCleared, result.Forgotten)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&mac, "mac", "", "client MAC to remove")
	f.StringVar(&name, "name", "", "client name to remove")
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	return cmd
}
```

Note: `result.FixedIPCleared`/`result.Forgotten` field names are assumed from `DeleteResult`'s described shape in `CLAUDE.md` ("reports both steps independently") — before using them, confirm the exact field names:

Run: `grep -n "type DeleteResult" -A5 utilities/internal/users/operations.go`

Adjust the two field references above to match exactly.

- [ ] **Step 4: Wire `users` into the root command**

`root.AddCommand(newConfigCommand(), newIPsCommand(), newDNSCommand(), newNetworkCommand(), newClientsCommand(), newUsersCommand())`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestUsers -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add utilities/gofi/users.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi users list/rm"
```

**Phase 5 checkpoint:** C-USERS-001 through C-USERS-005 satisfied; B-USERS-002's exit code updated to 3.

---

## Phase 6: `profile` area (new)

No predecessor tool. Scoped to networks + WLANs + fixed IPs only (C-PROFILE-001).

### Task 6.1: `utilities/internal/profile` — data shape and capture

**Files:**
- Create: `utilities/internal/profile/profile.go`
- Test: `utilities/internal/profile/profile_test.go`

**Interfaces:**
- Consumes: `network.ListNetworks` (Phase 3), `ips.DoGetEntries` (Phase 1, Task 1.3), `gofi.Client.WLANs()` (existing library service — confirm its method names before use, per Step 1 below).
- Produces: `profile.Profile` (struct: `Networks []network.NetworkEntry`, `WLANs []profile.WLANEntry`, `FixedIPs []ips.HostEntry`, `Captured string`, `Site string`), `profile.Capture(ctx, client, site string, withKeys bool) (*profile.Profile, error)`, `(*profile.Profile).Write(w io.Writer) error`, `profile.ReadProfile(r io.Reader) (*profile.Profile, error)`.

- [ ] **Step 1: Confirm the WLAN service's read method**

Run: `grep -n "^type WLANService interface" -A 15 src/services/wlan.go`

Use whatever method it reports (e.g. `List(ctx, site) ([]types.WLAN, error)`) in the implementation below — do not assume `List`.

- [ ] **Step 2: Write the failing tests**

```go
package profile

import (
	"bytes"
	"strings"
	"testing"
)

func TestCapture_omitsPassphrasesByDefault(t *testing.T) {
	client := newMockClientWithWLAN(t, "guest-wifi", "supersecret")

	p, err := Capture(context.Background(), client, "default", false)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	for _, w := range p.WLANs {
		if strings.Contains(w.Passphrase, "supersecret") {
			t.Errorf("WLAN %q leaked its passphrase without --with-keys", w.Name)
		}
	}
}

func TestCapture_withKeysIncludesPassphrases(t *testing.T) {
	client := newMockClientWithWLAN(t, "guest-wifi", "supersecret")

	p, err := Capture(context.Background(), client, "default", true)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	found := false
	for _, w := range p.WLANs {
		if w.Passphrase == "supersecret" {
			found = true
		}
	}
	if !found {
		t.Error("expected --with-keys to include the passphrase")
	}
}

func TestProfileWriteAndReadRoundTrip(t *testing.T) {
	p := &Profile{Site: "default", Captured: "2026-08-13T00:00:00Z"}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := ReadProfile(&buf)
	if err != nil {
		t.Fatalf("ReadProfile() error = %v", err)
	}
	if got.Site != p.Site {
		t.Errorf("Site = %q, want %q", got.Site, p.Site)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./utilities/internal/profile/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 4: Write the implementation**

```go
// Package profile captures a site's networks, WLANs, and fixed-IP
// reservations as JSON, and applies one back. Deliberately narrow
// (C-PROFILE-001): devices, firewall, routing, and port profiles are never
// captured, permanently, so "profile" can never silently grow to mean
// "everything on the site."
package profile

import (
	"context"
	"encoding/json"
	"io"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/ips"
	"github.com/unifi-go/gofi/utilities/internal/network"
)

// WLANEntry is the portable subset of a WLAN's configuration.
type WLANEntry struct {
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Passphrase string `json:"passphrase,omitempty"`
}

// Profile is a captured site: networks, WLANs, fixed IPs. Nothing else,
// permanently (C-PROFILE-001).
type Profile struct {
	Site     string               `json:"site"`
	Captured string               `json:"captured"`
	Networks []network.NetworkEntry `json:"networks"`
	WLANs    []WLANEntry          `json:"wlans"`
	FixedIPs []ips.HostEntry      `json:"fixed_ips"`
}

// Capture reads the three portable sections from the site. withKeys
// controls whether WLAN passphrases are included in cleartext
// (C-PROFILE-002); when false, Passphrase is left empty on every entry.
func Capture(ctx context.Context, client gofi.Client, site string, withKeys bool) (*Profile, error) {
	networks, err := network.ListNetworks(ctx, client, site)
	if err != nil {
		return nil, err
	}

	// wlans, err := client.WLANs().<the method confirmed in Step 1>(ctx, site)
	// if err != nil {
	// 	return nil, err
	// }
	// wlanEntries := make([]WLANEntry, 0, len(wlans))
	// for _, w := range wlans {
	// 	entry := WLANEntry{Name: w.Name, Enabled: w.Enabled}
	// 	if withKeys {
	// 		entry.Passphrase = w.Passphrase // field name to confirm against types.WLAN
	// 	}
	// 	wlanEntries = append(wlanEntries, entry)
	// }

	fixedIPs, err := ips.DoGetEntries(ctx, client, site, "")
	if err != nil {
		return nil, err
	}

	return &Profile{
		Site:     site,
		Networks: networks,
		// WLANs:    wlanEntries,
		FixedIPs: fixedIPs,
	}, nil
}

// Write serializes the profile as indented JSON.
func (p *Profile) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// ReadProfile parses a profile written by Write.
func ReadProfile(r io.Reader) (*Profile, error) {
	var p Profile
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
```

The `WLANs`-fetching block is left commented with the exact shape to fill in once Step 1's method name is confirmed — replace the comment block with real code calling the confirmed method and field names, then delete the comment markers. This is the one place in this plan where the exact call depends on a name this plan's author could not see directly; every other task uses names already confirmed against the real source.

- [ ] **Step 5: Fill in the WLAN fetch using the confirmed method name and run the tests**

Replace the commented block from Step 4 with real code using the method name found in Step 1, matching `types.WLAN`'s actual field names (check with `grep -n "^type WLAN struct" -A 15 src/types/*.go` if unsure).

Run: `go test ./utilities/internal/profile/... -v`
Expected: PASS, all three tests.

- [ ] **Step 6: Commit**

```bash
git add utilities/internal/profile
git commit -m "feat: add profile.Capture, scoped to networks + WLANs + fixed IPs"
```

### Task 6.2: `profile.Apply` — fixed order, warn-on-mismatch

**Files:**
- Modify: `utilities/internal/profile/profile.go`
- Modify: `utilities/internal/profile/profile_test.go`

**Interfaces:**
- Produces: `profile.Apply(ctx context.Context, client gofi.Client, p *Profile, dryRun bool, progress io.Writer) error`.

- [ ] **Step 1: Write the failing tests**

```go
func TestApply_skipsNetworksAbsentFromTarget(t *testing.T) {
	client := newMockClientWithNetworks(t, nil) // target site has no networks
	p := &Profile{
		Networks: []network.NetworkEntry{{Name: "Guest"}},
	}
	var progress bytes.Buffer
	if err := Apply(context.Background(), client, p, false, &progress); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !strings.Contains(progress.String(), "Guest") {
		t.Error("expected Apply to report the skipped network by name")
	}
}

func TestApply_dryRunMakesNoWrites(t *testing.T) {
	client := newMockClientRecordingWrites(t)
	p := &Profile{FixedIPs: []ips.HostEntry{{Hostname: "nas", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.13"}}}

	if err := Apply(context.Background(), client, p, true, io.Discard); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if client.WriteCount() != 0 {
		t.Errorf("WriteCount() = %d after dry-run, want 0", client.WriteCount())
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./utilities/internal/profile/... -run TestApply -v`
Expected: FAIL — `Apply` undefined.

- [ ] **Step 3: Write the implementation**

```go
// Apply applies a profile in the fixed order networks, fixed IPs, WLANs
// (C-PROFILE-003): fixed IPs must land inside an existing subnet, so
// networks go first. A network or WLAN the target site lacks is reported
// and skipped rather than failing the whole run (C-PROFILE-005).
//
// Devices, firewall, and routing are never written, even if such a field
// somehow appears in a hand-edited profile file (I-PROFILE-001) -- Profile's
// struct has no field for them, so there is nothing for Apply to read.
func Apply(ctx context.Context, client gofi.Client, p *Profile, dryRun bool, progress io.Writer) error {
	targetNetworks, err := network.ListNetworks(ctx, client, "default")
	if err != nil {
		return err
	}
	targetByName := make(map[string]network.NetworkEntry, len(targetNetworks))
	for _, n := range targetNetworks {
		targetByName[n.Name] = n
	}

	for _, n := range p.Networks {
		if _, ok := targetByName[n.Name]; !ok {
			fmt.Fprintf(progress, "skipping network %q: not present on the target site\n", n.Name)
			continue
		}
		fmt.Fprintf(progress, "network %q matches target; no write endpoint yet (C-NETWORK-004)\n", n.Name)
	}

	if len(p.FixedIPs) > 0 {
		if dryRun {
			fmt.Fprintf(progress, "would import %d fixed-IP assignment(s)\n", len(p.FixedIPs))
		} else {
			result, err := ips.DoSet(ctx, client, "default", p.FixedIPs, "", false, false)
			if err != nil {
				return err
			}
			fmt.Fprintf(progress, "fixed IPs: %d created, %d updated, %d skipped, %d errors\n",
				result.Created, result.Updated, result.Skipped, result.Errors)
		}
	}

	for _, w := range p.WLANs {
		if dryRun {
			fmt.Fprintf(progress, "would apply WLAN %q\n", w.Name)
			continue
		}
		fmt.Fprintf(progress, "WLAN %q: no write endpoint yet\n", w.Name)
	}

	return nil
}
```

Note: this task's `network`/WLAN sections are deliberately reporting-only, matching C-NETWORK-004/C-DNS-004's Blocked status — `profile import` cannot write what `network set` and a WLAN-write endpoint don't yet support. Once those land (their own future tasks, outside this plan), `Apply`'s two reporting branches become real writes; this task's job is only the fixed-order sequencing and the skip/report behavior C-PROFILE-003/004/005 require today.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./utilities/internal/profile/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/profile
git commit -m "feat: add profile.Apply in fixed order, reporting what it cannot yet write"
```

### Task 6.3: Wire `gofi profile export/import`

**Files:**
- Create: `utilities/gofi/profile.go`
- Modify: `utilities/gofi/root.go`
- Test: extend `utilities/gofi/cli_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestProfileImport_readsFromStdinWithNoArgs(t *testing.T) {
	cmd := newProfileCommand()
	cmd.SetIn(strings.NewReader(`{"site":"default"}`))
	cmd.SetArgs([]string{"import", "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("profile import: %v", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./utilities/gofi/... -run TestProfileImport -v`
Expected: FAIL — `newProfileCommand` undefined.

- [ ] **Step 3: Write the implementation**

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/profile"
)

func newProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Capture networks + WLANs + fixed IPs as JSON, apply one back",
		Long: `Capture a site's networks, WLANs, and fixed-IP reservations as JSON, and
apply one back.

Deliberately narrower than a full site dump: devices, firewall, routing, and
port profiles are never captured, permanently (C-PROFILE-001).`,
	}
	cmd.AddCommand(newProfileExportCommand(), newProfileImportCommand())
	return cmd
}

func newProfileExportCommand() *cobra.Command {
	var withKeys bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write a profile to stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			p, err := profile.Capture(cmd.Context(), client, siteFlag(), withKeys)
			if err != nil {
				return explain(err)
			}
			p.Captured = time.Now().Format(time.RFC3339)
			if !withKeys && len(p.WLANs) > 0 {
				fmt.Fprintln(cmd.ErrOrStderr(),
					"note: WiFi passphrases omitted. Use --with-keys to include them.")
			}
			return p.Write(cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&withKeys, "with-keys", false, "include WiFi passphrases in cleartext")
	return cmd
}

func newProfileImportCommand() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Apply a profile from a file, or from stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := cmd.InOrStdin()
			if len(args) == 1 && args[0] != "-" {
				f, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("open %s: %w", args[0], err)
				}
				defer f.Close()
				input = f
			}

			p, err := profile.ReadProfile(input)
			if err != nil {
				return fmt.Errorf("%w: %s", errUsage, err)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			return explain(profile.Apply(cmd.Context(), client, p, dryRun, cmd.ErrOrStderr()))
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	return cmd
}
```

- [ ] **Step 4: Wire `profile` into the root command**

`root.AddCommand(newConfigCommand(), newIPsCommand(), newDNSCommand(), newNetworkCommand(), newClientsCommand(), newUsersCommand(), newProfileCommand())`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./utilities/gofi/... -run TestProfile -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add utilities/gofi/profile.go utilities/gofi/root.go utilities/gofi/cli_test.go
git commit -m "feat: wire gofi profile export/import"
```

**Phase 6 checkpoint:** C-PROFILE-001 through C-PROFILE-005 satisfied for the sections that have a write endpoint today (fixed IPs); `network`/WLAN sections apply as reporting-only, correctly reflecting their Blocked status rather than silently no-op-ing.

---

## Phase 7: Migration cutover

One release, no wrapper binaries (C-GLOBAL-001).

### Task 7.1: Remove the five old binaries, update the Makefile

**Files:**
- Delete: `utilities/gofips/`, `utilities/gofimac/`, `utilities/gofinet/`, `utilities/gofidns/`, `utilities/gofiuser/` (each already emptied of its `operations.go`/`format.go`/etc. by Phases 1–5; only `main.go`, `parse.go` if any remain, and each directory's own `README.md`/`DESIGN.md` under `docs/` are left)
- Modify: `Makefile`

- [ ] **Step 1: Confirm nothing outside the five directories still imports them**

Run: `grep -rn "utilities/gofips\|utilities/gofimac\|utilities/gofinet\|utilities/gofidns\|utilities/gofiuser" --include=*.go .`
Expected: no matches outside the five directories themselves (every internal package extraction in Phases 1–5 already updated all real call sites).

- [ ] **Step 2: Delete the five directories**

```bash
git rm -r utilities/gofips utilities/gofimac utilities/gofinet utilities/gofidns utilities/gofiuser
```

- [ ] **Step 3: Update the Makefile's `UTILITIES` list**

Find the existing `UTILITIES := gofips gofimac gofinet gofidns gofiuser` (or equivalent) line and change it to:

```makefile
UTILITIES := gofi
```

- [ ] **Step 4: Build everything and run the full test suite**

Run: `make build && make test`
Expected: `bin/gofi` builds; every test across `utilities/internal/*` and `utilities/gofi` passes; no references to the deleted packages remain anywhere in the build.

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "chore: remove the five superseded binaries, gofi replaces them (C-GLOBAL-001)"
```

### Task 7.2: Publish the command-mapping table

**Files:**
- Modify: `README.md`
- Create or modify: `CHANGELOG.md` (if the project doesn't have one yet, create it with this entry as the first)

- [ ] **Step 1: Write the mapping table**

Add to `README.md`, in a new "Migrating from the old tools" section:

```markdown
## Migrating from the old tools

`gofips`, `gofimac`, `gofinet`, `gofidns`, and `gofiuser` are replaced by one binary,
`gofi`, as of this release. There are no wrapper binaries; update any script or cron
job using the table below.

| Old command | New command |
|---|---|
| `gofips -H <host> -k -g` | `gofi -H <host> --secure ips export` |
| `gofips -H <host> -k -s <file>` | `gofi -H <host> --secure ips import <file>` |
| `gofips -H <host> -k -a '<declaration>'` | `gofi -H <host> --secure ips add '<declaration>'` |
| `gofips -H <host> -k -d -n <name>` | `gofi -H <host> --secure ips rm --name <name>` |
| `gofips -H <host> -k -d -m <mac>` | `gofi -H <host> --secure ips rm --mac <mac>` |
| `gofimac -H <host> -k --wifi` | `gofi -H <host> --secure clients list --wifi` |
| `gofimac -H <host> -k --wired -j` | `gofi -H <host> --secure --output json clients list --wired` |
| `gofinet -H <host> -k` | `gofi -H <host> --secure network list` |
| `gofidns -H <host> -k -g` | `gofi -H <host> --secure dns list` |
| `gofidns -H <host> -k -d -n <name>` | `gofi -H <host> --secure dns rm --name <name>` |
| `gofiuser -H <host> -k -l` | `gofi -H <host> --secure users list` |
| `gofiuser -H <host> -k -d -m <mac>` | `gofi -H <host> --secure users rm --mac <mac>` |

`--target <name>` (via `gofi config init`) replaces repeating `-H`/`-k`/`-S` on every
invocation.
```

- [ ] **Step 2: Add a CHANGELOG entry**

```markdown
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
```

- [ ] **Step 3: Commit**

```bash
git add README.md CHANGELOG.md
git commit -m "docs: publish the gofips/gofimac/gofinet/gofidns/gofiuser -> gofi command mapping"
```

### Task 7.3: Shell completion and final README pass

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: cobra's built-in `completion` subcommand (added automatically to `root` by cobra; no code change needed — verify it's present).

- [ ] **Step 1: Verify completion is available**

Run: `./bin/gofi completion --help`
Expected: lists `bash`, `zsh`, `fish`, `powershell` subcommands (cobra adds this automatically once the root command exists — this step is verification, not implementation).

- [ ] **Step 2: Add the install snippet to README's quick start**

```markdown
Shell completion is free:

\`\`\`bash
gofi completion bash > /etc/bash_completion.d/gofi
\`\`\`
```

- [ ] **Step 3: Update README's quick-start example to use the new binary**

Replace any `gofips`/`gofimac`/etc. invocation examples in the README's existing quick-start section with `gofi <area> <action>` equivalents, using the mapping table from Task 7.2.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document gofi shell completion, update quick-start examples"
```

**Phase 7 checkpoint (final):** `make build && make test` passes with only `gofi` in `bin/`; `README.md` and `CHANGELOG.md` document the cutover; every requirement in `REQUIREMENTS.md` tagged Shipped or Target for an implemented area is satisfied — re-check `REQUIREMENTS.md` section by section against the built binary and update any Status tag whose item is now actually shipped, per that document's own Traceability rule.

---

## Self-Review Notes

- **Spec coverage:** Every `GLOBAL` requirement is implemented in Phase 0. Every `IPS` requirement is covered by Phase 1 (Tasks 1.1–1.4). Every `DNS` requirement by Phase 2. Every `NETWORK` requirement by Phase 3 (C-NETWORK-004 correctly left Blocked, not implemented). Every `CLIENTS` requirement by Phase 4. Every `USERS` requirement by Phase 5. Every `PROFILE` requirement by Phase 6 (network/WLAN write sections correctly report-only, pending their own Blocked endpoints). Every `CONFIG` requirement by Task 0.6.
- **Two tasks (4.2/4.3, 6.1) carry an explicit "confirm the exact name before use" step** rather than an invented signature, because this plan's author could not read `clients.FilterMode`'s constant names or `WLANService`'s method name directly — every other signature in this plan was confirmed against the actual source in `utilities/gofips`, `utilities/gofidns`, `utilities/gofinet`, `utilities/gofimac`, `utilities/gofiuser`, `src/client.go`, and `src/config.go` before being used.
- **Type consistency:** `ips.HostEntry`, `ips.DeleteIdentifier`, `ips.FormatOptions`, `dns.DeleteIdentifier`, `users.DeleteIdentifier` are used with the same field names across every task that touches them (`Hostname`/`MAC`/`IP` for `ips`; `Name`/`MAC` for `users`; `ID`/`Name`/`IP` for `dns`), matching what Tasks 1.1/2.1/5.1 moved unchanged from the existing `operations.go` files.
