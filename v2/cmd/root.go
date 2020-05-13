package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

var config = newConfig()

// Version information.
var (
	VersionStr string
	Commit     string
	Date       string
	BuiltBy    string
	Version    *semver.Version
)

var rootCmd = &cobra.Command{
	Use:               "chezmoi",
	Short:             "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: config.persistentPreRunRootE,
}

var (
	errExitFailure = errors.New("")

	workingDir string
	initErr    error
)

func init() {
	if err := config.init(rootCmd); err != nil {
		initErr = err
		return
	}
}

// Execute executes the root command.
func Execute() error {
	if initErr != nil {
		return initErr
	}

	var versionComponents []string
	if VersionStr != "" {
		var err error
		Version, err = semver.NewVersion(strings.TrimPrefix(VersionStr, "v"))
		if err != nil {
			return err
		}
		versionComponents = append(versionComponents, VersionStr)
	} else {
		versionComponents = append(versionComponents, "dev")
	}
	if Commit != "" {
		versionComponents = append(versionComponents, "commit "+Commit)
	}
	if Date != "" {
		versionComponents = append(versionComponents, "built at "+Date)
	}
	if BuiltBy != "" {
		versionComponents = append(versionComponents, "built by "+BuiltBy)
	}
	rootCmd.Version = strings.Join(versionComponents, ", ")

	return rootCmd.Execute()
}

func getAsset(name string) ([]byte, error) {
	asset, ok := assets[name]
	if !ok {
		return nil, fmt.Errorf("%s: not found", name)
	}
	return asset, nil
}

func getExample(command string) string {
	return helps[command].example
}

func markRemainingZshCompPositionalArgumentsAsFiles(cmd *cobra.Command, from int) {
	// As far as I can tell, there is no way to mark all remaining positional
	// arguments as files. Marking the first eight positional arguments as files
	// should be enough for everybody.
	// FIXME mark all remaining positional arguments as files
	for i := 0; i < 8; i++ {
		panicOnError(cmd.MarkZshCompPositionalArgumentFile(from + i))
	}
}

func mustGetLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("no long help for %s", command))
	}
	return help.long
}
