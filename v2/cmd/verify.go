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
	RunE:    config.runVerifyCmd,
}

type verifyCmdConfig struct {
	include *chezmoi.IncludeBits
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	persistentFlags := verifyCmd.PersistentFlags()
	persistentFlags.VarP(config.verify.include, "include", "i", "include entry types")

	markRemainingZshCompPositionalArgumentsAsFiles(verifyCmd, 1)
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	canarySystem := chezmoi.NewCanarySystem(chezmoi.NewNullSystem())
	if err := c.applyArgs(canarySystem, "", args, c.verify.include); err != nil {
		return err
	}
	if canarySystem.Mutated() {
		if c.debug {
			fmt.Println(strings.Join(canarySystem.Mutations(), " "))
		}
		return errExitFailure
	}
	return nil
}
