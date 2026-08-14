package main

import (
	"errors"
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

		// Runnable + Args so an unknown subcommand under this area is a
		// usage error (exit 2) rather than cobra's silent help-with-exit-0
		// for a non-runnable parent (C-GLOBAL-012).
		Args: wrapArgsError(unknownSubcommandArgs),
		RunE: showHelp,
	}
	cmd.AddCommand(newProfileExportCommand(), newProfileImportCommand())
	return cmd
}

func newProfileExportCommand() *cobra.Command {
	var withKeys bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write a profile to stdout",
		Args:  wrapArgsError(cobra.NoArgs),
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
	var dnsDomain string
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Apply a profile from a file, or from stdin",
		Args:  wrapArgsError(cobra.MaximumNArgs(1)),
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

			// siteFlag(), not the profile's own "site" field: a profile
			// is applied to the site the invocation resolves, so
			// -S/--site and the target's site both take effect.
			err = profile.Apply(cmd.Context(), client, siteFlag(), p, dryRun, dnsDomain, cmd.ErrOrStderr())
			if errors.Is(err, profile.ErrNoSite) {
				return fmt.Errorf("%w: %s", errUsage, err)
			}
			return explain(err)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	f.StringVar(&dnsDomain, "dns-domain", os.Getenv("UNIFI_DNS_DOMAIN"), "DNS suffix for fixed-IP hostnames, overriding the profile's own network domain")
	return cmd
}
