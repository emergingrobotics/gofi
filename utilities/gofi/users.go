package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/users"
)

func newUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Known-client identity records, connected or not",
		Long: `Known-client identity records: persistent, outlive the connection, carry the
fixed-IP relationship. See "gofi clients" for currently-connected stations
only -- the two read different backend objects, kept as separate areas
permanently (VISION.md's Area boundaries).`,
	}
	cmd.AddCommand(newUsersListCommand(), newUsersRmCommand())
	return cmd
}

func newUsersListCommand() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known clients",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			return explain(users.DoList(cmd.Context(), client, siteFlag(), filter,
				users.FormatOptions{Writer: cmd.OutOrStdout(), JSON: asJSON()}))
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "only list entries containing this substring")
	return cmd
}

func newUsersRmCommand() *cobra.Command {
	var mac, name string
	var dryRun bool
	cmd := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"remove"},
		Short:   "Remove a known client, by --mac/--name",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if (mac == "") == (name == "") {
				return fmt.Errorf("%w: pass exactly one of --mac or --name", errUsage)
			}

			client, err := connect()
			if err != nil {
				return err
			}
			defer client.Disconnect(cmd.Context())

			identifier := users.DeleteIdentifier{MAC: mac, Name: name}
			result, err := users.DoDel(cmd.Context(), client, siteFlag(), identifier, dryRun, cmd.ErrOrStderr())
			if errors.Is(err, users.ErrAmbiguous) {
				// An ambiguous match is a guard refusal (C-USERS exit-code
				// contract): exit 3, not 1.
				return fmt.Errorf("%w: %s", errRefused, err)
			}
			if err != nil {
				return explain(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed fixed IP: %t, forgot client: %t\n",
				result.ClearedFixedIP, result.Forgot)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&mac, "mac", "", "client MAC to remove")
	f.StringVar(&name, "name", "", "client name to remove")
	f.BoolVar(&dryRun, "dry-run", false, "show what would change without changing it")
	return cmd
}
