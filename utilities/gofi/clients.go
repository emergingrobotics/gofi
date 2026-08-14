package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/clients"
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
	cmd.AddCommand(newClientsListCommand())
	cmd.AddCommand(newClientsVendorCommand())
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
				filter = clients.FilterWifi
			}
			if wired {
				filter = clients.FilterWired
			}

			db, err := clients.LoadOUIDatabase()
			if err != nil {
				return err
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entries, err := clients.ListClients(cmd.Context(), client, siteFlag(), filter, clients.SortIP, db)
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

func newClientsVendorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "vendor <mac>",
		Short: "Look up a MAC address's manufacturer, without contacting the controller",
		Long: `Look up a MAC address's manufacturer.

Entirely offline (C-CLIENTS-004): reads the cached IEEE OUI registry and
never opens a controller session.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := clients.LoadOUIDatabase()
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
