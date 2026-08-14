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
