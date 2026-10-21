package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gherlein/gofi/skill"
	"github.com/gherlein/gofi/utilities/internal/installskill"
)

func newInstallSkillCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the agent skill that teaches gofi into the skills store",
		Long: `Install the agent skill that teaches an AI agent to use gofi.

The skill is embedded in this binary and written to $XDG_DATA_HOME/skills/` + skill.Name + `
(default ~/.local/share/skills/` + skill.Name + `), per the Skill Self-Installation
Specification. Nothing else in the store is touched. Agents that read only their own
skills directory see it after a mirror step such as ` + "`sm sync`" + `.

Needs no config file, credentials, or controller.

An existing, different copy is left alone unless -f/--force is given. An identical
copy is a successful no-op. Never prompts.

Exit codes:
  0  installed, replaced, or already up to date
  1  failure: store not writable, I/O error, or store location unresolvable
  2  usage error
  3  refused: a different copy is installed and --force was not given`,
		Args: wrapArgsError(cobra.NoArgs),

		// Replaces the root's PersistentPreRunE, which loads config.toml and
		// resolves a target: installing a skill needs neither, and a broken
		// config must not block it (C-SKILL-004).
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },

		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := installskill.StorePath()
			if err != nil {
				return err
			}
			outcome, path, err := installskill.Install(skill.FS(), skill.Name, store, force)
			if errors.Is(err, installskill.ErrConflict) {
				return fmt.Errorf("%w: %s", errRefused, err)
			}
			if err != nil {
				return err
			}
			switch outcome {
			case installskill.Unchanged:
				fmt.Fprintf(cmd.OutOrStdout(), "skill %s already up to date at %s; nothing changed\n", skill.Name, path)
			case installskill.Replaced:
				fmt.Fprintf(cmd.OutOrStdout(), "replaced skill %s at %s\n", skill.Name, path)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "installed skill %s at %s\n", skill.Name, path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "replace an existing, different copy")
	return cmd
}
