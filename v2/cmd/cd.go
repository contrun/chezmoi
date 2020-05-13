package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
)

var cdCmd = &cobra.Command{
	Use:     "cd",
	Args:    cobra.NoArgs,
	Short:   "Launch a shell in the source directory",
	Long:    mustGetLongHelp("cd"),
	Example: getExample("cd"),
	RunE:    config.runCDCmd,
	Annotations: map[string]string{
		doesNotRequireValidConfig: "true",
		modifiesSourceDirectory:   "true",
	},
}

type cdCmdConfig struct {
	Command string
	Args    []string
}

func init() {
	rootCmd.AddCommand(cdCmd)
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	if err := c.ensureSourceDirectory(); err != nil {
		return err
	}

	shellCommand := c.CD.Command
	if shellCommand == "" {
		shellCommand, _ = shell.CurrentUserShell()
	}
	return c.run(c.SourceDir, shellCommand, c.CD.Args)
}
