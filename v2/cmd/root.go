package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

const (
	doesNotRequireValidConfig    = "chezmoi_annotation_does_not_require_valid_config"
	modifiesConfigFile           = "chezmoi_annotation_modifies_config_file"
	modifiesDestinationDirectory = "chezmoi_annotation_modifies_destination_directory"
	modifiesSourceDirectory      = "chezmoi_annotation_modifies_source_directory"
	requiresConfigDirectory      = "chezmoi_annotation_requires_config_directory"
	requiresSourceDirectory      = "chezmoi_annotation_requires_source_directory"
)

var config = mustNewConfig()

// Version information.
var (
	VersionStr string
	Commit     string
	Date       string
	BuiltBy    string
	Version    *semver.Version
)

var rootCmd = &cobra.Command{
	Use:                "chezmoi",
	Short:              "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:      true,
	SilenceUsage:       true,
	PersistentPreRunE:  config.persistentPreRunRootE,
	PersistentPostRunE: config.persistentPostRunRootE,
}

var (
	errExitFailure = errors.New("")
	initErr        error
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

func getBoolAnnotation(cmd *cobra.Command, key string) bool {
	value, ok := cmd.Annotations[key]
	if !ok {
		return false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}
	return boolValue
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
		panic(fmt.Sprintf("%s: no long help", command))
	}
	return help.long
}
