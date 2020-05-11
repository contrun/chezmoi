package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var sourcePathCmd = &cobra.Command{
	Use:     "source-path [targets...]",
	Short:   "Print the path of a target in the source state",
	Long:    mustGetLongHelp("source-path"),
	Example: getExample("source-path"),
	PreRunE: config.ensureNoError,
	RunE:    config.runSourcePathCmd,
}

func init() {
	rootCmd.AddCommand(sourcePathCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(sourcePathCmd, 1)
}

func (c *Config) runSourcePathCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	if len(args) == 0 {
		return c.writeOutputString(c.SourceDir)
	}

	s, err := c.getSourceState()
	if err != nil {
		return err
	}

	targetNames, err := c.getTargetNames(s, args, getTargetNamesOptions{
		recursive:           false,
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	sb := &strings.Builder{}
	for _, targetName := range targetNames {
		sb.WriteString(s.Entries[targetName].Path())
		sb.WriteByte('\n')
	}
	return c.writeOutputString(sb.String())
}
