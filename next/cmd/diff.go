package cmd

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Print the diff between the target state and the destination state",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	RunE:    config.runDiffCmd,
}

type diffCmdConfig struct {
	include   *chezmoi.IncludeBits
	recursive bool
	NoPager   bool
	Pager     string
}

func init() {
	rootCmd.AddCommand(diffCmd)

	persistentFlags := diffCmd.PersistentFlags()
	persistentFlags.VarP(config.Diff.include, "include", "i", "include entry types")
	persistentFlags.BoolVar(&config.Diff.NoPager, "no-pager", config.Diff.NoPager, "disable pager")
	persistentFlags.BoolVarP(&config.Diff.recursive, "recursive", "r", config.Diff.recursive, "recursive")

	markRemainingZshCompPositionalArgumentsAsFiles(diffCmd, 1)
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	sb := &strings.Builder{}
	unifiedEncoder := diff.NewUnifiedEncoder(sb, diff.DefaultContextLines)
	if c.colored {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	gitDiffSystem := chezmoi.NewGitDiffSystem(unifiedEncoder, c.system, c.DestDir+chezmoi.PathSeparatorStr)
	if err := c.applyArgs(gitDiffSystem, c.DestDir, args, c.Diff.include, c.Diff.recursive); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}
