package cmd

import (
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:     "apply [targets...]",
	Short:   "Update the destination directory to match the target state",
	Long:    mustGetLongHelp("apply"),
	Example: getExample("apply"),
	RunE:    config.runApplyCmd,
	Annotations: map[string]string{
		modifiesDestinationDirectory: "true",
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	markRemainingZshCompPositionalArgumentsAsFiles(applyCmd, 1)
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(c.system, c.DestDir, args)
}
