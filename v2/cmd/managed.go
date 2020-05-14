package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var managedCmd = &cobra.Command{
	Use:     "managed",
	Args:    cobra.NoArgs,
	Short:   "List the managed entries in the destination directory",
	Long:    mustGetLongHelp("managed"),
	Example: getExample("managed"),
	RunE:    config.runManagedCmd,
}

type managedCmdConfig struct {
	include []string
}

func init() {
	rootCmd.AddCommand(managedCmd)

	persistentFlags := managedCmd.PersistentFlags()
	persistentFlags.StringSliceVarP(&config.managed.include, "include", "i", config.managed.include, "include")
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string) error {
	var (
		includeAll      = false
		includeAbsent   = false
		includeDirs     = false
		includeFiles    = false
		includeScripts  = false
		includeSymlinks = false
	)
	for _, what := range c.managed.include {
		switch what {
		case "all":
			includeAll = true
		case "absent", "a":
			includeAbsent = true
		case "dirs", "d":
			includeDirs = true
		case "files", "f":
			includeFiles = true
		case "scripts":
			includeScripts = true
		case "symlinks", "s":
			includeSymlinks = true
		default:
			return fmt.Errorf("%s: unrecognized include", what)
		}
	}

	s, err := c.getSourceState()
	if err != nil {
		return err
	}

	entries := s.Entries()
	targetNames := make([]string, 0, len(entries))
	for targetName, sourceStateEntry := range entries {
		targetStateEntry, err := sourceStateEntry.TargetStateEntry()
		if err != nil {
			return err
		}
		if !includeAll {
			switch targetStateEntry.(type) {
			case *chezmoi.TargetStateAbsent:
				if !includeAbsent {
					continue
				}
			case *chezmoi.TargetStateDir:
				if !includeDirs {
					continue
				}
			case *chezmoi.TargetStateFile:
				if !includeFiles {
					continue
				}
			case *chezmoi.TargetStateScript:
				if !includeScripts {
					continue
				}
			case *chezmoi.TargetStateSymlink:
				if !includeSymlinks {
					continue
				}
			}
		}
		targetNames = append(targetNames, targetName)
	}

	sort.Strings(targetNames)
	sb := &strings.Builder{}
	for _, targetName := range targetNames {
		sb.WriteString(filepath.FromSlash(filepath.Join(c.DestDir, targetName)) + eolStr)
	}
	return c.writeOutputString(sb.String())
}
