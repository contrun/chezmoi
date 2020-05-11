package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-xdg/v3"
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
	var err error
	workingDir, err = os.Getwd()
	if err != nil {
		initErr = err
		return
	}
	workingDir = filepath.ToSlash(workingDir)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		initErr = err
		return
	}
	homeDir = filepath.ToSlash(homeDir)

	config.bds, err = xdg.NewBaseDirectorySpecification()
	if err != nil {
		initErr = err
		return
	}

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVarP(&config.configFile, "config", "c", getDefaultConfigFile(config.bds), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	panicOnError(viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run")))

	persistentFlags.BoolVar(&config.Follow, "follow", false, "follow symlinks")
	panicOnError(viper.BindPFlag("follow", persistentFlags.Lookup("follow")))

	persistentFlags.BoolVar(&config.Force, "force", config.Force, "force")
	panicOnError(viper.BindPFlag("force", persistentFlags.Lookup("force")))

	persistentFlags.StringVar(&config.Format, "format", config.Format, "format ("+serializationFormatNamesStr()+")")
	panicOnError(viper.BindPFlag("format", persistentFlags.Lookup("format")))

	persistentFlags.BoolVarP(&config.Recursive, "recursive", "r", config.Recursive, "recursive")
	panicOnError(viper.BindPFlag("recursive", persistentFlags.Lookup("recursive")))

	persistentFlags.BoolVar(&config.Remove, "remove", false, "remove targets")
	panicOnError(viper.BindPFlag("remove", persistentFlags.Lookup("remove")))

	persistentFlags.StringVarP(&config.SourceDir, "source", "S", getDefaultSourceDir(config.bds), "source directory")
	panicOnError(rootCmd.MarkPersistentFlagDirname("source"))
	panicOnError(viper.BindPFlag("source", persistentFlags.Lookup("source")))

	persistentFlags.StringVarP(&config.DestDir, "destination", "D", homeDir, "destination directory")
	panicOnError(rootCmd.MarkPersistentFlagDirname("destination"))
	panicOnError(viper.BindPFlag("destination", persistentFlags.Lookup("destination")))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	panicOnError(viper.BindPFlag("verbose", persistentFlags.Lookup("verbose")))

	persistentFlags.StringVar(&config.Color, "color", "auto", "colorize diffs")
	panicOnError(viper.BindPFlag("color", persistentFlags.Lookup("color")))

	persistentFlags.StringVarP(&config.Output, "output", "o", "", "output file")
	panicOnError(rootCmd.MarkPersistentFlagFilename("output"))
	panicOnError(viper.BindPFlag("output", persistentFlags.Lookup("output")))

	persistentFlags.BoolVar(&config.Debug, "debug", false, "write debug logs")
	panicOnError(viper.BindPFlag("debug", persistentFlags.Lookup("debug")))

	cobra.OnInitialize(func() {
		_, err := os.Stat(config.configFile)
		switch {
		case err == nil:
			viper.SetConfigFile(config.configFile)
			config.err = viper.ReadInConfig()
			if config.err == nil {
				config.err = viper.Unmarshal(&config)
			}
			if config.err == nil {
				config.err = config.validateData()
			}
			if config.err != nil {
				rootCmd.Printf("warning: %s: %v\n", config.configFile, config.err)
			}
		case os.IsNotExist(err):
		default:
			initErr = err
		}
	})
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
