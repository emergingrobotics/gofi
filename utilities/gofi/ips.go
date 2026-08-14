package main

import (
	"fmt"
	"io"
	"os"
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
