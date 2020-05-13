package cmd

import (
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:     "data",
	Args:    cobra.NoArgs,
	Short:   "Print the template data",
	Long:    mustGetLongHelp("data"),
	Example: getExample("data"),
	RunE:    config.runDataCmd,
}

func init() {
	rootCmd.AddCommand(dataCmd)
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string) error {
	s, err := c.getSourceState()
	if err != nil {
		return err
	}
	return c.marshal(s.TemplateData())
}
