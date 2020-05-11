package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove targets...",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state and the destination directory",
	Long:    mustGetLongHelp("remove"),
	Example: getExample("remove"),
	PreRunE: config.ensureNoError,
	RunE:    config.runRemoveCmd,
	// PostRunE: config.autoCommitAndAutoPush,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	markRemainingZshCompPositionalArgumentsAsFiles(removeCmd, 1)
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string) error {
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

	for _, targetName := range targetNames {
		destPath := path.Join(c.DestDir, targetName)
		sourcePath := s.Entries[targetName].Path()
		if !c.Force {
			choice, err := c.prompt(fmt.Sprintf("Remove %s and %s", destPath, sourcePath), "ynqa")
			if err != nil {
				return err
			}
			switch choice {
			case 'y':
			case 'n':
				continue
			case 'q':
				return nil
			case 'a':
				c.Force = true
			}
		}
		if err := c.system.RemoveAll(destPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := c.system.RemoveAll(sourcePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
