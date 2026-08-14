package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/utilities/internal/config"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "gofi's own configuration file",
		Long: `gofi's own configuration file.

Acts on your machine, never on a controller. Holds named targets and an
output preference. Secrets never appear here (C-CONFIG's invariant): a
password or API key comes from the environment, a command the file names,
or an interactive prompt.`,

		// Runnable + Args so an unknown subcommand under this area is a
		// usage error (exit 2) rather than cobra's silent help-with-exit-0
		// for a non-runnable parent (C-GLOBAL-012).
		Args: wrapArgsError(unknownSubcommandArgs),
		RunE: showHelp,
	}
	cmd.AddCommand(newConfigShowCommand(), newConfigTargetsCommand(), newConfigInitCommand())
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Report the config file's location and what it resolves to",
		Args:  wrapArgsError(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			exists := "yes"
			if _, err := os.Stat(opts.file.Path()); err != nil {
				exists = "no (flags and the environment still work)"
			}

			if asJSON() {
				out := map[string]any{
					"path":   opts.file.Path(),
					"exists": exists == "yes",
					"output": opts.output,
				}
				if opts.resolvedTarget != nil {
					out["target"] = opts.resolvedTarget.Name()
					out["mode"] = opts.resolvedTarget.Mode()
				}
				return writeJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "PATH       %s\nEXISTS     %s\nOUTPUT     %s\n",
				opts.file.Path(), exists, opts.output)
			if opts.resolvedTarget != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "TARGET     %s (%s mode)\n",
					opts.resolvedTarget.Name(), opts.resolvedTarget.Mode())
			}
			return nil
		},
	}
}

func newConfigTargetsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List the configured targets",
		Args:  wrapArgsError(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			names := opts.file.Names()
			if asJSON() {
				return writeJSON(cmd.OutOrStdout(), names)
			}
			if len(names) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no targets configured in %s\n", opts.file.Path())
				fmt.Fprintln(cmd.OutOrStdout(), "run `gofi config init` to write a starting point")
				return nil
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tMODE")
			for _, name := range names {
				t := opts.file.Targets[name]
				marker := name
				if name == opts.file.Default {
					marker += " (default)"
				}
				fmt.Fprintf(tw, "%s\t%s\n", marker, t.Mode())
			}
			return tw.Flush()
		},
	}
}

func newConfigInitCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starting configuration file",
		Args:  wrapArgsError(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.Path()
			if _, err := os.Stat(path); err == nil && !force {
				return fmt.Errorf("%w: %s already exists; pass --force to overwrite", errUsage, path)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(starterConfig), 0o600); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
			fmt.Fprintln(cmd.OutOrStdout(), "edit it, then run `gofi config show`")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")
	return cmd
}

const starterConfig = `# gofi configuration. See ` + "`gofi config show`" + ` for where this was read from.
#
# Secrets never belong here. A password or API key comes from the environment,
# from a command named below, or from a prompt -- in that order.

# Which target to use when --target is not given. Optional with only one defined.
default = "home"

# Output format for every command: "text" or "json". Overridden by --output.
output = "text"

# Local mode: connect directly to a controller.
[targets.home]
host             = "192.168.1.1"
site             = "default"
# password_command = "pass show unifi/home"

# Connector mode: connect through api.ui.com. Uncomment and fill in console_id.
# [targets.cloud]
# site             = "default"
# console_id       = "your-console-id"
# api_key_command  = "pass show unifi/cloud-key"
`
