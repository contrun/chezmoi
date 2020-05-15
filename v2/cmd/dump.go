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
	RunE:    config.runDumpCmd,
}

type dumpCmdConfig struct {
	include *chezmoi.IncludeBits
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	persistentFlags := dumpCmd.PersistentFlags()
	persistentFlags.VarP(config.dump.include, "include", "i", "include entry types")

	markRemainingZshCompPositionalArgumentsAsFiles(dumpCmd, 1)
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dataSystem := chezmoi.NewDataSystem()
	if err := c.applyArgs(dataSystem, "", args, c.dump.include); err != nil {
		return err
	}
	return c.marshal(dataSystem.Data())
}
