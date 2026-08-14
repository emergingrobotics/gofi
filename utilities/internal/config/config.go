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
