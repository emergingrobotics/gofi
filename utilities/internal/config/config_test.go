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
