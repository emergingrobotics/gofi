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
	// list and export are the same read, per C-IPS's shared behavior; build a
	// fresh export command (so flags aren't shared) and rename it, since
	// reusing the command verbatim would register two "export" entries and
	// leave "list" unrecognized as a subcommand name.
	cmd := newIPsExportCommand()
	cmd.Use = "list"
	cmd.Short = "List every fixed-IP assignment to stdout in ISC DHCP format"
	return cmd
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
			return explain(ips.DoGet(cmd.Context(), client, siteFlag(), dnsDomain, cmd.OutOrStdout(), opts))
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
			return explain(ips.DoDel(cmd.Context(), client, siteFlag(), identifier, dnsDomain, force, keepDNS, dryRun))
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
