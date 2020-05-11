package cmd

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type diffCmdConfig struct {
	NoPager bool
	Pager   string
}

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Print the diff between the target state and the destination state",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDiffCmd,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	persistentFlags := diffCmd.PersistentFlags()
	persistentFlags.BoolVar(&config.Diff.NoPager, "no-pager", config.Diff.NoPager, "disable pager")

	markRemainingZshCompPositionalArgumentsAsFiles(diffCmd, 1)
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	sb := &strings.Builder{}
	unifiedEncoder := diff.NewUnifiedEncoder(sb, diff.DefaultContextLines)
	if c.colored {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	gitDiffSystem := chezmoi.NewGitDiffSystem(unifiedEncoder, c.system, c.DestDir+chezmoi.PathSeparatorStr)
	if err := c.applyArgs(gitDiffSystem, c.DestDir, args); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}
