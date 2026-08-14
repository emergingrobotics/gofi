package main

import (
	"errors"
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

		// Runnable + Args so an unknown subcommand under this area is a
		// usage error (exit 2) rather than cobra's silent help-with-exit-0
		// for a non-runnable parent (C-GLOBAL-012).
		Args: wrapArgsError(unknownSubcommandArgs),
		RunE: showHelp,
	}
	cmd.AddCommand(newDNSListCommand(), newDNSRmCommand())
	return cmd
}

func newDNSListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List DNS records",
		Args:  wrapArgsError(cobra.NoArgs),
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
		Args:    wrapArgsError(cobra.NoArgs),
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
			if errors.Is(err, dns.ErrAmbiguous) {
				// An ambiguous, unforced match is a guard refusal
				// (C-DNS-003): exit 3, not 1.
				return fmt.Errorf("%w: %s", errRefused, err)
			}
			if err != nil {
				return explain(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d record(s)\n", result.Deleted)
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
