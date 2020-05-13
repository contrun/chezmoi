package cmd

import (
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit targets...",
	Short:   "Edit the source state of a target",
	Long:    mustGetLongHelp("edit"),
	Example: getExample("edit"),
	RunE:    config.runEditCmd,
	Annotations: map[string]string{
		requiresSourceDirectory: "true",
	},
}

type editCmdConfig struct {
	apply  bool
	diff   bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(editCmd)

	persistentFlags := editCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.edit.apply, "apply", "a", false, "apply edit after editing")
	persistentFlags.BoolVarP(&config.edit.diff, "diff", "d", false, "print diff after editing")
	persistentFlags.BoolVarP(&config.edit.prompt, "prompt", "p", false, "prompt before applying (implies --diff)")

	markRemainingZshCompPositionalArgumentsAsFiles(editCmd, 1)
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string) error {
	// FIXME
	return nil
}
