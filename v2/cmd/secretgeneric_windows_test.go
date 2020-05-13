// +build windows

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func getSecretTestConfig(t *testing.T) (*Config, []string) {
	// Windows doesn't (usually) have "date", but powershell is included with
	// all versions of Windows v7 or newer.
	c, err := newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "powershell.exe",
		}),
	)
	require.NoError(t, c)
	return c, []string{
		"-NoProfile",
		"-NonInteractive",
		"-Command", "Get-Date",
	}
}

func getSecretJSONTestConfig(t *testing.T) (*Config, []string) {
	c, err := newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "powershell.exe",
		}),
	)
	require.NoError(t, err)
	return c, []string{
		"-NoProfile",
		"-NonInteractive",
		"-Command", "Get-Date | ConvertTo-Json",
	}
}
