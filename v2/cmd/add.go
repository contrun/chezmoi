package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add targets...",
	Aliases: []string{"manage"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Add an existing file, directory, or symlink to the source state",
	Long:    mustGetLongHelp("add"),
	Example: getExample("add"),
	// RunE:     config.runAddCmd, // FIXME
	Annotations: map[string]string{
		modifiesSourceDirectory: "true",
	},
}

type addCmdConfig struct{}
