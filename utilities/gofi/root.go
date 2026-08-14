package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

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
	// opts is package-level (every command reads it), so start each command
	// tree from a clean slate rather than inheriting a previous one's flag
	// values -- notably opts.secure, which loadConfig only writes when the
	// flag was actually given.
	opts = globals{}

	root := &cobra.Command{
		Use:   "gofi",
		Short: "Manage a UniFi controller's networks, DNS, reservations, and clients",
		Long: `gofi manages a UniFi controller (UDM Pro and compatible consoles) over its
local or cloud-connector API.

It exists to make a site's addressing, DNS names, networks, and known clients
reproducible from a file rather than from click history in the UniFi app.`,

		SilenceErrors: true,
		SilenceUsage:  true,

		// Replaces cobra's implicit legacyArgs for a root with
		// subcommands, which reports an unknown command with a bare
		// error that would exit 1 instead of 2 (C-GLOBAL-012).
		Args: wrapArgsError(unknownSubcommandArgs),

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadConfig(cmd)
		},

		// RunE makes root "runnable" (cobra.Command.Runnable()) so `--help`
		// renders the Usage/Flags block even before any subcommand is
		// registered (Task 0.6 adds the first one, "config"). Bare `gofi`
		// invocation just shows help, same as cobra's default for a
		// non-runnable multi-command root.
		RunE: showHelp,
	}

	// Cobra's own parse failures (unknown flag, malformed value) are plain
	// errors; without this they would exit 1 instead of the required 2
	// (C-GLOBAL-012). SetFlagErrorFunc is inherited by every subcommand.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %s", errUsage, err)
	})

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
		newUsersCommand(),
		newProfileCommand(),
	)
	return root
}

// wrapArgsError adapts a positional-argument validator so its rejection
// carries errUsage, and therefore exit 2 (C-GLOBAL-012). Cobra's built-in
// validators (ExactArgs, NoArgs, MaximumNArgs) return bare errors, and
// SetFlagErrorFunc only covers flag parsing, so every `Args:` assignment in
// the command tree goes through here.
func wrapArgsError(validator cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validator(cmd, args); err != nil {
			return fmt.Errorf("%w: %s", errUsage, err)
		}
		return nil
	}
}

// showHelp is the RunE of every command that only groups subcommands. It
// makes those commands Runnable, which is what lets cobra reach their Args
// validator instead of short-circuiting to help with exit 0.
func showHelp(cmd *cobra.Command, _ []string) error { return cmd.Help() }

// unknownSubcommandArgs reproduces cobra's unexported legacyArgs behavior for
// a command that has subcommands: any leftover positional argument is an
// unknown command, reported with cobra's own "did you mean" suggestions.
func unknownSubcommandArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if cmd.DisableSuggestions {
		return errors.New(msg)
	}
	// SuggestionsFor uses the field as-is; cobra's own call path applies
	// this default first, so mirror it or every distance test fails.
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += fmt.Sprintf("\n\nDid you mean this?\n\t%s", strings.Join(suggestions, "\n\t"))
	}
	return errors.New(msg)
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
		// A rejected --secure is a mistake in how the command was invoked,
		// not a controller/site failure: exit 2, not 1 (C-GLOBAL-007).
		if errors.Is(err, conn.ErrSecureFlagNotApplicable) {
			return nil, fmt.Errorf("%w: %s", errUsage, err)
		}
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
