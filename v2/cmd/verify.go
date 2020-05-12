package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var verifyCmd = &cobra.Command{
	Use:     "verify [targets...]",
	Short:   "Exit with success if the destination state matches the target state, fail otherwise",
	Long:    mustGetLongHelp("verify"),
	Example: getExample("verify"),
	PreRunE: config.ensureNoError,
	RunE:    config.runVerifyCmd,
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(verifyCmd, 1)
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	canarySystem := chezmoi.NewCanarySystem(chezmoi.NewNullSystem())
	if err := c.applyArgs(canarySystem, "", args); err != nil {
		return err
	}
	if canarySystem.Mutated() {
		if c.Debug {
			fmt.Println(strings.Join(canarySystem.Mutations(), " "))
		}
		return errExitFailure
	}
	return nil
}
