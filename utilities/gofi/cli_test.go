package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInit_writesStarterFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOFI_CONFIG", filepath.Join(dir, "config.toml"))

	var out bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"config", "init"})
	if err := root.Execute(); err != nil {
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

	var out bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"config", "init"})
	if err := root.Execute(); err == nil {
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
	root := newRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"config", "targets"})
	if err := root.Execute(); err != nil {
		t.Fatalf("config targets: %v", err)
	}
}

func TestIPsRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newIPsCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("ips rm with no identifier: error = nil, want a usage error")
	}
}

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

func TestIPsAdd_rejectsFlagsAndPositionalTogether(t *testing.T) {
	cmd := newIPsCommand()
	cmd.SetArgs([]string{"add", "--name", "x", "--mac", "aa:bb:cc:dd:ee:ff", "--ip", "10.0.0.1",
		"host x { hardware ethernet aa:bb:cc:dd:ee:ff; fixed-address 10.0.0.1; }"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("ips add with both flags and a positional declaration: error = nil, want a usage error")
	}
}

func TestDNSRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newDNSCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("dns rm with no identifier: error = nil, want a usage error")
	}
}

func TestNetworkShow_requiresOneArg(t *testing.T) {
	cmd := newNetworkCommand()
	cmd.SetArgs([]string{"show"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("network show with no name: error = nil, want a usage error")
	}
}
