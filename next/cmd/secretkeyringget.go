package cmd

import (
	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
)

var keyringGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get a password from keyring",
	RunE:  config.runKeyringGetCmd,
}

func init() {
	keyringCmd.AddCommand(keyringGetCmd)
}

func (c *Config) runKeyringGetCmd(cmd *cobra.Command, args []string) error {
	password, err := keyring.Get(c.Keyring.service, c.Keyring.user)
	if err != nil {
		return err
	}
	return c.writeOutputString(password)
}
