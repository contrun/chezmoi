// +build !windows

package cmd

import (
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func getSecretTestConfig() (*Config, []string) {
	return newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "date",
		}),
	), []string{"+%Y-%M-%DT%H:%M:%SZ"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return newConfig(
		withSystem(chezmoi.NewNullSystem()),
		withGenericSecretCmdConfig(genericSecretCmdConfig{
			Command: "date",
		}),
	), []string{`+{"date":"%Y-%M-%DT%H:%M:%SZ"}`}
}
