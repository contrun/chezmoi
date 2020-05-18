package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var keyringSetCmd = &cobra.Command{
	Use:   "set",
	Args:  cobra.NoArgs,
	Short: "Set a password in keyring",
	RunE:  config.runKeyringSetCmd,
}

func init() {
	keyringCmd.AddCommand(keyringSetCmd)

	persistentFlags := keyringSetCmd.PersistentFlags()
	persistentFlags.StringVar(&config.Keyring.password, "password", "", "password")
}

func (c *Config) runKeyringSetCmd(cmd *cobra.Command, args []string) error {
	passwordString := c.Keyring.password
	if passwordString == "" {
		fmt.Print("Password: ")
		password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		passwordString = string(password)
	}
	return keyring.Set(c.Keyring.service, c.Keyring.user, passwordString)
}
