package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/config"
	"github.com/unifi-go/gofi/utilities/internal/conn"
)

// globals holds every persistent flag, shared by every command (C-GLOBAL-009).
type globals struct {
	target string
	host   string
	port   int
	site   string
	secure *bool // nil = flag not given, distinguishing "not passed" for C-GLOBAL-007
	output string

	file           *config.File
	resolvedTarget *config.Target
}

var opts globals

const (
	outputText = "text"
	outputJSON = "json"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "gofi",
		Short: "Manage a UniFi controller's networks, DNS, reservations, and clients",
		Long: `gofi manages a UniFi controller (UDM Pro and compatible consoles) over its
local or cloud-connector API.

It exists to make a site's addressing, DNS names, networks, and known clients
reproducible from a file rather than from click history in the UniFi app.`,

		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadConfig(cmd)
		},

		// RunE makes root "runnable" (cobra.Command.Runnable()) so `--help`
		// renders the Usage/Flags block even before any subcommand is
		// registered (Task 0.6 adds the first one, "config"). Bare `gofi`
		// invocation just shows help, same as cobra's default for a
		// non-runnable multi-command root.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	flags := root.PersistentFlags()
	flags.StringVar(&opts.target, "target", "", "named controller+site from the config file")
	flags.StringVarP(&opts.host, "host", "H", "", "controller address")
	flags.IntVarP(&opts.port, "port", "p", 0, "controller port (default 443)")
	flags.StringVarP(&opts.site, "site", "S", "", "site name (default \"default\")")
	flags.StringVar(&opts.output, "output", outputText, "output format: text or json")
	flags.Bool("secure", false, "enforce TLS certificate verification (local mode only)")

	root.AddCommand(
		newConfigCommand(),
		newIPsCommand(),
		newDNSCommand(),
		newNetworkCommand(),
		newClientsCommand(),
	)
	// newUsersCommand(), newProfileCommand() are added by Phases 5-6.
	return root
}

func loadConfig(cmd *cobra.Command) error {
	file, err := config.Load()
	if err != nil {
		return err
	}
	opts.file = file

	if opts.output != outputText && opts.output != outputJSON {
		return fmt.Errorf("%w: --output %q: want %q or %q", errUsage, opts.output, outputText, outputJSON)
	}
	if !cmd.Flags().Changed("output") && file.Output != "" {
		opts.output = file.Output
	}

	if cmd.Flags().Changed("secure") {
		v, _ := cmd.Flags().GetBool("secure")
		opts.secure = &v
	}

	target, err := file.Resolve(opts.target)
	if err != nil {
		return fmt.Errorf("%w: %s", errUsage, err)
	}
	opts.resolvedTarget = target
	return nil
}

// connect resolves the connection, authenticates, and returns a live client.
func connect() (gofi.Client, error) {
	cfg, err := conn.ResolveTargetConfig(os.Stderr, opts.resolvedTarget, opts.host, opts.port, opts.site, opts.secure,
		func() (string, error) {
			prompt := "password: "
			if opts.resolvedTarget != nil && opts.resolvedTarget.Mode() == "connector" {
				prompt = "API key: "
			}
			return conn.ReadSecret(prompt, "")
		})
	if err != nil {
		return nil, err
	}

	client, err := gofi.New(cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}

// explain is a pass-through today; kept as the single seam every command
// routes errors through, matching gogl's convention, so a future
// backend-specific error-annotation pass has one place to land.
func explain(err error) error { return err }

func asJSON() bool { return opts.output == outputJSON }

// errRefused marks an error as a guard refusal (C-GLOBAL-012, exit 3) rather
// than a failure (exit 1).
var errRefused = errors.New("refused")

// errUsage marks an error caused by how the command was invoked (exit 2)
// rather than by controller/site state.
var errUsage = errors.New("usage")
