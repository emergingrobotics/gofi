package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/network"
)

func newNetworkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Networks (VLANs): subnet, DHCP pool, DNS servers",
		Long: `Networks on the site: subnet, VLAN tag, DHCP pool boundaries, DNS servers.

Read-only today (C-NETWORK-004): a write endpoint has not yet been verified
against real hardware.`,
	}
	cmd.AddCommand(newNetworkListCommand(), newNetworkShowCommand())
	return cmd
}

func newNetworkListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every network on the site",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entries, err := network.ListNetworks(cmd.Context(), client, siteFlag())
			if err != nil {
				return explain(err)
			}
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), entries)
			}
			return network.FormatText(cmd.OutOrStdout(), entries)
		},
	}
}

func newNetworkShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Report one network in detail",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("%w: expected exactly one network name", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			entry, err := network.FindByName(cmd.Context(), client, siteFlag(), args[0])
			if errors.Is(err, network.ErrNotFound) {
				return err
			}
			if err != nil {
				return explain(err)
			}
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), entry)
			}
			return network.FormatText(cmd.OutOrStdout(), []network.NetworkEntry{*entry})
		},
	}
}
