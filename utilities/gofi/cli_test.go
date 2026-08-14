package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestUsersRm_requiresExactlyOneIdentifier(t *testing.T) {
	cmd := newUsersCommand()
	cmd.SetArgs([]string{"rm"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("users rm with no identifier: error = nil, want a usage error")
	}
}

func TestClientsList_rejectsWifiAndWiredTogether(t *testing.T) {
	cmd := newClientsCommand()
	cmd.SetArgs([]string{"list", "--wifi", "--wired"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("clients list --wifi --wired: error = nil, want a usage error")
	}
}

func TestClientsVendor_worksWithoutAControllerSession(t *testing.T) {
	seedOUIFixture(t)

	var out bytes.Buffer
	cmd := newClientsCommand()
	cmd.SetOut(&out)
	// No --host, --target, or credentials given: this must not attempt to
	// connect (I-CLIENTS-001).
	cmd.SetArgs([]string{"vendor", "b4:0e:cf:2a:85:6f"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("clients vendor: %v (should need no controller connection)", err)
	}
	// Also proves the seeded fixture is what answered, rather than the test
	// silently falling through to a downloaded registry.
	if !strings.Contains(out.String(), "Texas Instruments") {
		t.Errorf("clients vendor output = %q, want the manufacturer from the seeded OUI fixture", out.String())
	}
}

// seedOUIFixture points XDG_DATA_HOME at a temp dir holding a hand-written,
// freshly-modified oui.txt in the IEEE "(hex)" line format. Without it,
// LoadOUIDatabase would find no cache and download the real registry from
// standards-oui.ieee.org, making `go test ./...` fail offline and crawl on a
// cold CI run.
func seedOUIFixture(t *testing.T) {
	t.Helper()

	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	dir := filepath.Join(dataHome, "gofi")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	fixture := "OUI/MA-L                                                    Organization\n" +
		"company_id                                                  Organization\n" +
		"                                                            Address\n\n" +
		"B4-0E-CF     (hex)\t\tTexas Instruments\n" +
		"B40ECF     (base 16)\t\tTexas Instruments\n" +
		"\t\t\t\t12500 TI Blvd\n" +
		"\t\t\t\tDallas TX 75243\n" +
		"\t\t\t\tUS\n\n" +
		"00-1A-2B     (hex)\t\tAyecom Technology Co., Ltd.\n" +
		"001A2B     (base 16)\t\tAyecom Technology Co., Ltd.\n"
	if err := os.WriteFile(filepath.Join(dir, "oui.txt"), []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestProfileImport_rejectsMalformedJSON(t *testing.T) {
	cmd := newProfileCommand()
	cmd.SetIn(strings.NewReader(`not json`))
	cmd.SetArgs([]string{"import", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("profile import with malformed JSON: error = nil, want a usage error")
	}
	if !errors.Is(err, errUsage) {
		t.Errorf("profile import with malformed JSON: error = %v, want errUsage", err)
	}
}

// TestProfileImport_readsFromStdinWithNoArgs exercises the same wiring as
// the brief's suggested test (well-formed JSON on stdin, no positional
// file), but clears every credential env var first. This sandbox happens to
// carry real UNIFI_API_KEY/UNIFI_CONSOLE_ID values, and connect() honors
// them unconditionally (C-GLOBAL-006..010) -- without clearing them, this
// test would make a live network call to api.ui.com with real credentials.
// With credentials cleared, ReadProfile/JSON-decode wiring is still
// exercised (a usage error would surface first if it weren't), and connect()
// fails deterministically before any network I/O.
func TestProfileImport_readsFromStdinWithNoArgs(t *testing.T) {
	for _, k := range []string{"UNIFI_API_KEY", "UNIFI_CONSOLE_ID", "UNIFI_USERNAME", "UNIFI_PASSWORD", "UNIFI_CONTROLLER_IP"} {
		t.Setenv(k, "")
	}
	t.Setenv("GOFI_CONFIG", filepath.Join(t.TempDir(), "config.toml"))

	cmd := newProfileCommand()
	cmd.SetIn(strings.NewReader(`{"site":"default"}`))
	cmd.SetArgs([]string{"import", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("profile import: error = nil, want a connection error (no credentials in test env)")
	}
	if errors.Is(err, errUsage) {
		t.Errorf("profile import: error = %v, want a connection failure, not a usage error (JSON parsed fine)", err)
	}
}

// TestProfileImport_dnsDomainFlagIsAccepted mirrors
// TestProfileImport_readsFromStdinWithNoArgs: credentials are cleared so
// connect() fails deterministically before any network I/O, which is enough
// to prove --dns-domain parses as a flag (a usage error would surface first
// otherwise) and reaches the point of connecting, without depending on
// whatever real credentials this sandbox happens to carry.
func TestProfileImport_dnsDomainFlagIsAccepted(t *testing.T) {
	for _, k := range []string{"UNIFI_API_KEY", "UNIFI_CONSOLE_ID", "UNIFI_USERNAME", "UNIFI_PASSWORD", "UNIFI_CONTROLLER_IP"} {
		t.Setenv(k, "")
	}
	t.Setenv("GOFI_CONFIG", filepath.Join(t.TempDir(), "config.toml"))

	cmd := newProfileCommand()
	cmd.SetIn(strings.NewReader(`{"site":"default"}`))
	cmd.SetArgs([]string{"import", "--dry-run", "--dns-domain", "bench.test"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("profile import --dns-domain: error = nil, want a connection error (no credentials in test env)")
	}
	if errors.Is(err, errUsage) {
		t.Errorf("profile import --dns-domain: error = %v, want a connection failure, not a usage error (flag should be accepted)", err)
	}
}

// TestExitCodes_cobraGeneratedUsageErrors covers the invocations whose errors
// come from cobra itself rather than from a hand-written check: an unknown
// command, an unknown flag, and positional-arg validators across several
// areas. Each must exit 2 under C-GLOBAL-012, not cobra's default 1.
func TestExitCodes_cobraGeneratedUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown area", []string{"badarea"}},
		{"unknown subcommand of an area", []string{"network", "bogus"}},
		{"unknown flag", []string{"network", "list", "--bogus"}},
		{"flag missing its value", []string{"network", "list", "--output"}},
		{"too many args (MaximumNArgs)", []string{"network", "show", "a", "b"}},
		{"too few args (ExactArgs)", []string{"clients", "vendor"}},
		{"args on a NoArgs command", []string{"config", "targets", "extra"}},
		{"too many args on ips import", []string{"ips", "import", "a", "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOFI_CONFIG", filepath.Join(t.TempDir(), "config.toml"))

			root := newRootCommand()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(tc.args)

			err := root.Execute()
			if err == nil {
				t.Fatalf("gofi %v: error = nil, want a usage error", tc.args)
			}
			if got := codeFor(err); got != exitUsage {
				t.Errorf("gofi %v: exit code = %d (%v), want %d", tc.args, got, err, exitUsage)
			}
		})
	}
}

// TestSecureAgainstConnectorTarget_exitsUsage is the end-to-end half of
// C-GLOBAL-007: conn returns a sentinel, connect() classifies it, codeFor maps
// it to 2.
func TestSecureAgainstConnectorTarget_exitsUsage(t *testing.T) {
	t.Setenv("GOFI_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("UNIFI_API_KEY", "k123")
	t.Setenv("UNIFI_CONSOLE_ID", "console-abc")

	root := newRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"network", "list", "--secure"})

	err := root.Execute()
	if err == nil {
		t.Fatal("gofi --secure network list against a connector target: error = nil, want a usage error")
	}
	if got := codeFor(err); got != exitUsage {
		t.Errorf("exit code = %d (%v), want %d", got, err, exitUsage)
	}
	if !strings.Contains(err.Error(), "--secure") {
		t.Errorf("error = %v, want it to name --secure", err)
	}
	if strings.Contains(err.Error(), "--insecure") {
		t.Errorf("error = %v, names --insecure, a flag gofi does not have", err)
	}
}

func TestNetworkShow_requiresOneArg(t *testing.T) {
	cmd := newNetworkCommand()
	cmd.SetArgs([]string{"show"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("network show with no name: error = nil, want a usage error")
	}
}
