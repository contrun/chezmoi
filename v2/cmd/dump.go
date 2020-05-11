package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var dumpCmd = &cobra.Command{
	Use:     "dump [targets...]",
	Short:   "Generate a dump of the target state",
	Long:    mustGetLongHelp("dump"),
	Example: getExample("dump"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDumpCmd,
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	markRemainingZshCompPositionalArgumentsAsFiles(dumpCmd, 1)
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dataSystem := chezmoi.NewDataSystem()
	if err := c.applyArgs(dataSystem, "", args); err != nil {
		return err
	}
	return c.marshal(dataSystem.Data())
}
