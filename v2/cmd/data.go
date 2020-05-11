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
	PreRunE: config.ensureNoError,
	RunE:    config.runDataCmd,
}

func init() {
	rootCmd.AddCommand(dataCmd)
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	templateData, err := c.getTemplateData()
	if err != nil {
		return err
	}
	return c.marshal(templateData)
}
