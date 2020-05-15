package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var forgetCmd = &cobra.Command{
	Use:     "forget targets...",
	Aliases: []string{"unmanage"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state",
	Long:    mustGetLongHelp("forget"),
	Example: getExample("forget"),
	RunE:    config.runForgetCmd,
	Annotations: map[string]string{
		modifiesSourceDirectory: "true",
	},
}

func init() {
	rootCmd.AddCommand(forgetCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(forgetCmd, 1)
}

func (c *Config) runForgetCmd(cmd *cobra.Command, args []string) error {
	s, err := c.getSourceState()
	if err != nil {
		return err
	}

	sourcePaths, err := c.getSourcePaths(s, args)
	if err != nil {
		return err
	}

	for _, sourcePath := range sourcePaths {
		if !c.force {
			choice, err := c.prompt(fmt.Sprintf("Remove %s", sourcePath), "ynqa")
			if err != nil {
				return err
			}
			switch choice {
			case 'y':
			case 'n':
				continue
			case 'q':
				return nil
			case 'a':
				c.force = false
			}
		}
		if err := c.system.RemoveAll(sourcePath); err != nil {
			return err
		}
	}

	return nil
}
