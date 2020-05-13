// +build !windows

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func getSecretTestConfig(t *testing.T) (*Config, []string) {
	c, err := newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "date",
		}),
	)
	require.NoError(t, err)
	return c, []string{"+%Y-%M-%DT%H:%M:%SZ"}
}

func getSecretJSONTestConfig(t *testing.T) (*Config, []string) {
	c, err := newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "date",
		}),
	)
	require.NoError(t, err)
	return c, []string{`+{"date":"%Y-%M-%DT%H:%M:%SZ"}`}
}
