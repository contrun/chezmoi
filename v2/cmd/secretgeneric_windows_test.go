// +build windows

package cmd

import (
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func getSecretTestConfig() (*Config, []string) {
	// Windows doesn't (usually) have "date", but powershell is included with
	// all versions of Windows v7 or newer.
	return newConfig(
			withSystem(chezmoi.NewNullSystem()),
			withGenericSecretCmdConfig(genericSecretCmdConfig{
				Command: "powershell.exe",
			}),
		), []string{
			"-NoProfile",
			"-NonInteractive",
			"-Command", "Get-Date",
		}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return newConfig(
			withSystem(chezmoi.NewNullSystem()),
			withGenericSecretCmdConfig(genericSecretCmdConfig{
				Command: "powershell.exe",
			}),
		), []string{
			"-NoProfile",
			"-NonInteractive",
			"-Command", "Get-Date | ConvertTo-Json",
		}
}
