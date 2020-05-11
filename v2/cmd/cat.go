package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var catCmd = &cobra.Command{
	Use:     "cat targets...",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Print the target contents of a file or symlink",
	Long:    mustGetLongHelp("cat"),
	Example: getExample("cat"),
	PreRunE: config.ensureNoError,
	RunE:    config.runCatCmd,
}

func init() {
	rootCmd.AddCommand(catCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(catCmd, 1)
}

func (c *Config) runCatCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	s, err := c.getSourceState()
	if err != nil {
		return err
	}
	targetNames, err := c.getTargetNames(s, args, getTargetNamesOptions{
		recursive:           c.Recursive,
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	sb := &strings.Builder{}
	for _, targetName := range targetNames {
		targetStateEntry, err := s.Entries[targetName].TargetStateEntry()
		if err != nil {
			return fmt.Errorf("%s: %w", targetName, err)
		}
		switch targetStateEntry := targetStateEntry.(type) {
		case *chezmoi.TargetStateFile:
			contents, err := targetStateEntry.Contents()
			if err != nil {
				return fmt.Errorf("%s: %w", targetName, err)
			}
			sb.Write(contents)
		case *chezmoi.TargetStateSymlink:
			linkname, err := targetStateEntry.Linkname()
			if err != nil {
				return fmt.Errorf("%s: %w", targetName, err)
			}
			sb.WriteString(linkname)
		default:
			return fmt.Errorf("%s: not a file or symlink", targetName)
		}
	}
	return c.writeOutputString(sb.String())
}
