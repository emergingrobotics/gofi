package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unifi-go/gofi/skill"
)

func newPrintSkillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print-skill",
		Short: "Print the agent skill that teaches gofi to stdout",
		Long: `Print the agent skill that teaches an AI agent to use gofi.

Writes the carried skill's SKILL.md -- frontmatter and body, exact bytes -- to
stdout and nothing else: no banner, path, or heading. The content comes from
inside this binary, never from the skills store or an installed copy, so the
output is the same whether or not the skill has ever been installed with
` + "`gofi install-skill`" + `.

Needs no config file, credentials, or controller, and changes nothing on the
filesystem -- unlike install-skill, print-skill never touches the store.

Exit codes:
  0  SKILL.md written to stdout
  1  failure: the embedded skill could not be read (unreachable in a correctly
     built binary)`,
		Args: wrapArgsError(cobra.NoArgs),

		// Replaces the root's PersistentPreRunE, which loads config.toml and
		// resolves a target: printing the skill needs neither, and a broken
		// config must not block it (C-SKILL-008, mirroring install-skill's
		// C-SKILL-004).
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },

		RunE: func(cmd *cobra.Command, _ []string) error {
			data, err := skill.MD()
			if err != nil {
				return fmt.Errorf("reading embedded skill: %w", err)
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
	return cmd
}
