package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion shell",
	Args:      cobra.ExactArgs(1),
	Short:     "Generate shell completion code for the specified shell (bash, fish, or zsh)",
	Long:      mustGetLongHelp("completion"),
	Example:   getExample("completion"),
	ValidArgs: []string{"bash", "fish", "zsh"},
	RunE:      config.runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func (c *Config) runCompletion(cmd *cobra.Command, args []string) error {
	sb := &strings.Builder{}
	switch args[0] {
	case "bash":
		if err := rootCmd.GenBashCompletion(sb); err != nil {
			return err
		}
	case "fish":
		if err := rootCmd.GenFishCompletion(sb, true); err != nil {
			return err
		}
	case "zsh":
		if err := rootCmd.GenZshCompletion(sb); err != nil {
			return err
		}
	default:
		return errors.New("unsupported shell")
	}
	return c.writeOutputString(sb.String())
}
