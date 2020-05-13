package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-xdg/v3"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type sourceVCSConfig struct {
	Command    string
	AutoCommit bool
	AutoPush   bool
	Init       interface{}
	Pull       interface{}
}

type templateConfig struct {
	Options []string
}

// A Config represents a configuration.
type Config struct {
	homeDir    string
	workingDir string
	bds        *xdg.BaseDirectorySpecification

	configFile string
	err        error
	fs         vfs.FS
	system     chezmoi.System
	colored    bool

	// Global configuration, settable in the config file.
	SourceDir string
	DestDir   string
	Umask     permValue
	Format    string
	Follow    bool
	Remove    bool
	Color     string
	SourceVCS sourceVCSConfig
	Data      map[string]interface{}
	Template  templateConfig

	// Global configuration, not settable in the config file.
	debug         bool
	dryRun        bool
	force         bool
	output        string
	recursive     bool
	verbose       bool
	templateFuncs template.FuncMap

	// Password manager configurations, settable in the config file.
	Bitwarden     bitwardenCmdConfig
	GenericSecret genericSecretCmdConfig
	Gopass        gopassCmdConfig
	Keepassxc     keepassxcCmdConfig
	Keyring       keyringCmdConfig
	Lastpass      lastpassCmdConfig
	Onepassword   onepasswordCmdConfig
	Pass          passCmdConfig
	Vault         vaultCmdConfig

	// Command configurations, settable in the config file.
	CD   cdCmdConfig
	Diff diffCmdConfig

	// Command configurations, not settable in the config file.
	add             addCmdConfig
	edit            editCmdConfig
	executeTemplate executeTemplateCmdConfig
	managed         managedCmdConfig

	scriptStateBucket []byte
	Stdin             io.Reader
	Stdout            io.WriteCloser
	Stderr            io.WriteCloser
}

// A configOption sets and option on a Config.
type configOption func(*Config)

var (
	persistentStateFilename    = "chezmoistate.boltdb"
	commitMessageTemplateAsset = "assets/templates/COMMIT_MESSAGE.tmpl"

	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": {},
		"CPE":  {},
		"ID":   {},
		"URL":  {},
	}

	identifierRegexp = regexp.MustCompile(`\A[\pL_][\pL\p{Nd}_]*\z`)
	whitespaceRegexp = regexp.MustCompile(`\s+`)

	assets = make(map[string][]byte)
)

// newConfig creates a new Config with the given options.
func newConfig(options ...configOption) (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	bds, err := xdg.NewBaseDirectorySpecification()
	if err != nil {
		return nil, err
	}

	c := &Config{
		homeDir:    filepath.ToSlash(homeDir),
		workingDir: filepath.ToSlash(workingDir),
		bds:        bds,
		configFile: getDefaultConfigFile(bds),
		DestDir:    filepath.ToSlash(homeDir),
		SourceDir:  getDefaultSourceDir(bds),
		Umask:      permValue(getUmask()),
		Color:      "auto",
		Format:     "json",
		recursive:  true,
		SourceVCS: sourceVCSConfig{
			Command: "git",
		},
		Template: templateConfig{
			Options: chezmoi.DefaultTemplateOptions,
		},
		templateFuncs: sprig.TxtFuncMap(),
		Bitwarden: bitwardenCmdConfig{
			Command: "bw",
		},
		Gopass: gopassCmdConfig{
			Command: "gopass",
		},
		Keepassxc: keepassxcCmdConfig{
			Command: "keepassxc-cli",
		},
		Lastpass: lastpassCmdConfig{
			Command: "lpass",
		},
		Onepassword: onepasswordCmdConfig{
			Command: "op",
		},
		Pass: passCmdConfig{
			Command: "pass",
		},
		Vault: vaultCmdConfig{
			Command: "vault",
		},
		managed: managedCmdConfig{
			include: []string{"dirs", "files", "symlinks"},
		},
		scriptStateBucket: []byte("script"),
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
	}
	for _, option := range options {
		option(c)
	}
	return c, nil
}

func mustNewConfig(options ...configOption) *Config {
	c, err := newConfig(options...)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Config) addTemplateFunc(key string, value interface{}) {
	if c.templateFuncs == nil {
		c.templateFuncs = make(template.FuncMap)
	}
	if _, ok := c.templateFuncs[key]; ok {
		panic(fmt.Sprintf("Config.addTemplateFunc: %s already defined", key))
	}
	c.templateFuncs[key] = value
}

func (c *Config) applyArgs(targetSystem chezmoi.System, targetDir string, args []string) error {
	s, err := c.getSourceState()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return s.ApplyAll(targetSystem, os.FileMode(c.Umask), targetDir)
	}

	targetNames, err := c.getTargetNames(s, args, getTargetNamesOptions{
		recursive:           c.recursive,
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	for _, targetName := range targetNames {
		if err := s.ApplyOne(targetSystem, os.FileMode(c.Umask), targetDir, targetName); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ensureSourceDirectory() error {
	info, err := c.fs.Stat(c.SourceDir)
	switch {
	case err == nil && info.IsDir():
		if chezmoi.POSIXFileModes && info.Mode()&os.ModePerm&0o77 != 0 {
			return c.system.Chmod(c.SourceDir, 0o700&^os.FileMode(c.Umask))
		}
		return nil
	case os.IsNotExist(err):
		if err := vfs.MkdirAll(c.system, filepath.Dir(c.SourceDir), 0o777&^os.FileMode(c.Umask)); err != nil {
			return err
		}
		return c.system.Mkdir(c.SourceDir, 0o700&^os.FileMode(c.Umask))
	case err == nil:
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	default:
		return err
	}
}

func (c *Config) getDefaultTemplateData() (map[string]interface{}, error) {
	data := map[string]interface{}{
		"arch":      runtime.GOARCH,
		"os":        runtime.GOOS,
		"sourceDir": c.SourceDir,
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	data["username"] = currentUser.Username

	// user.LookupGroupId is generally unreliable:
	//
	// If CGO is enabled, then this uses an underlying C library call (e.g.
	// getgrgid_r on Linux) and is trustworthy, except on recent versions of Go
	// on Android, where LookupGroupId is not implemented.
	//
	// If CGO is disabled then the fallback implementation only searches
	// /etc/group, which is typically empty if an external directory service is
	// being used, and so the lookup fails.
	//
	// So, only set group if user.LookupGroupId does not return an error.
	group, err := user.LookupGroupId(currentUser.Gid)
	if err == nil {
		data["group"] = group.Name
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data["homedir"] = homedir

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	data["fullHostname"] = hostname
	data["hostname"] = strings.SplitN(hostname, ".", 2)[0]

	osRelease, err := getOSRelease(c.fs)
	if err == nil {
		if osRelease != nil {
			data["osRelease"] = upperSnakeCaseToCamelCaseMap(osRelease)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	kernelInfo, err := getKernelInfo(c.fs)
	if err == nil && kernelInfo != nil {
		data["kernel"] = kernelInfo
	} else if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Config) getPersistentState(options *bolt.Options) (chezmoi.PersistentState, error) {
	persistentStateFile := c.getPersistentStateFile()
	if c.dryRun {
		if options == nil {
			options = &bolt.Options{}
		}
		options.ReadOnly = true
	}
	return chezmoi.NewBoltPersistentState(c.fs, persistentStateFile, options)
}

func (c *Config) getPersistentStateFile() string {
	if c.configFile != "" {
		return filepath.Join(filepath.Dir(c.configFile), persistentStateFilename)
	}
	for _, configDir := range c.bds.ConfigDirs {
		persistentStateFile := filepath.Join(configDir, "chezmoi", persistentStateFilename)
		if _, err := os.Stat(persistentStateFile); err == nil {
			return persistentStateFile
		}
	}
	return filepath.Join(filepath.Dir(getDefaultConfigFile(c.bds)), persistentStateFilename)
}

func (c *Config) getSourceState() (*chezmoi.SourceState, error) {
	templateData, err := c.getTemplateData()
	if err != nil {
		return nil, err
	}

	s := chezmoi.NewSourceState(
		chezmoi.WithSourcePath(c.SourceDir),
		chezmoi.WithSystem(c.system),
		chezmoi.WithTemplateData(templateData),
		chezmoi.WithTemplateFuncs(c.templateFuncs),
		chezmoi.WithTemplateOptions(c.Template.Options),
	)

	if err := s.Read(); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Config) getTargetName(arg string) (string, error) {
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(workingDir, arg)
	}
	arg = filepath.ToSlash(filepath.Clean(arg))
	destDirPrefix := c.DestDir + chezmoi.PathSeparatorStr
	if !strings.HasPrefix(arg, destDirPrefix) {
		return "", fmt.Errorf("%s: not in destination directory", arg)
	}
	return strings.TrimPrefix(arg, destDirPrefix), nil
}

type getTargetNamesOptions struct {
	recursive           bool
	mustBeInSourceState bool
}

func (c *Config) getTargetNames(s *chezmoi.SourceState, args []string, options getTargetNamesOptions) ([]string, error) {
	targetNames := make([]string, 0, len(args))
	for _, arg := range args {
		targetName, err := c.getTargetName(arg)
		if err != nil {
			return nil, err
		}
		if options.mustBeInSourceState {
			if _, ok := s.Entries[targetName]; !ok {
				return nil, fmt.Errorf("%s: not in source state", arg)
			}
		}
		targetNames = append(targetNames, targetName)
		if options.recursive {
			targetNamePrefix := targetName + chezmoi.PathSeparatorStr
			for targetName := range s.Entries {
				if strings.HasPrefix(targetName, targetNamePrefix) {
					targetNames = append(targetNames, targetName)
				}
			}
		}
	}

	if len(targetNames) == 0 {
		return nil, nil
	}

	// Sort and de-duplicate targetNames in place.
	sort.Strings(targetNames)
	n := 1
	for i := 1; i < len(targetNames); i++ {
		if targetNames[i] != targetNames[i-1] {
			targetNames[n] = targetNames[i]
			n++
		}
	}
	return targetNames[:n], nil
}

func (c *Config) getTemplateData() (map[string]interface{}, error) {
	defaultData, err := c.getDefaultTemplateData()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"chezmoi": defaultData,
	}
	for key, value := range c.Data {
		data[key] = value
	}
	return data, nil
}

func (c *Config) init(rootCmd *cobra.Command) error {
	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVar(&c.Color, "color", c.Color, "colorize diffs")
	persistentFlags.StringVarP(&c.DestDir, "destination", "D", c.DestDir, "destination directory")
	persistentFlags.BoolVar(&c.Follow, "follow", c.Follow, "follow symlinks")
	persistentFlags.StringVar(&c.Format, "format", c.Format, "format ("+serializationFormatNamesStr()+")")
	persistentFlags.BoolVar(&c.Remove, "remove", c.Remove, "remove targets")
	persistentFlags.StringVarP(&c.SourceDir, "source", "S", c.SourceDir, "source directory")
	for _, key := range []string{
		"color",
		"destination",
		"follow",
		"format",
		"remove",
		"source",
	} {
		if err := viper.BindPFlag(key, persistentFlags.Lookup(key)); err != nil {
			return err
		}
	}

	persistentFlags.StringVarP(&c.configFile, "config", "c", c.configFile, "config file")
	persistentFlags.BoolVarP(&c.dryRun, "dry-run", "n", c.dryRun, "dry run")
	persistentFlags.BoolVar(&c.force, "force", c.force, "force")
	persistentFlags.BoolVarP(&c.recursive, "recursive", "r", c.recursive, "recursive")
	persistentFlags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "verbose")
	persistentFlags.StringVarP(&c.output, "output", "o", c.output, "output file")
	persistentFlags.BoolVar(&c.debug, "debug", c.debug, "write debug logs")

	for _, err := range []error{
		rootCmd.MarkPersistentFlagDirname("destination"),
		rootCmd.MarkPersistentFlagFilename("output"),
		rootCmd.MarkPersistentFlagDirname("source"),
	} {
		if err != nil {
			return err
		}
	}

	cobra.OnInitialize(func() {
		_, err := os.Stat(c.configFile)
		switch {
		case err == nil:
			viper.SetConfigFile(c.configFile)
			c.err = viper.ReadInConfig()
			if c.err == nil {
				c.err = viper.Unmarshal(&config)
			}
			if c.err == nil {
				c.err = c.validateData()
			}
			if c.err != nil {
				rootCmd.Printf("warning: %s: %v\n", c.configFile, c.err)
			}
		case os.IsNotExist(err):
		default:
			initErr = err
		}
	})

	return nil
}

func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	if !getBoolAnnotation(cmd, doesNotRequireValidConfig) {
		if c.err != nil {
			return errors.New("config contains errors, aborting")
		}
	}

	if colored, err := strconv.ParseBool(c.Color); err == nil {
		c.colored = colored
	} else {
		switch c.Color {
		case "on":
			c.colored = true
		case "off":
			c.colored = false
		case "auto":
			if _, ok := os.LookupEnv("NO_COLOR"); ok {
				c.colored = false
			} else if stdout, ok := c.Stdout.(*os.File); ok {
				c.colored = terminal.IsTerminal(int(stdout.Fd()))
			} else {
				c.colored = false
			}
		default:
			return fmt.Errorf("%s: invalid color value", c.Color)
		}
	}

	if c.colored {
		if err := enableVirtualTerminalProcessingOnWindows(c.Stdout); err != nil {
			return err
		}
	}

	c.fs = vfs.OSFS
	persistentState, err := c.getPersistentState(nil)
	if err != nil {
		return initErr
	}
	c.system = chezmoi.NewRealSystem(c.fs, persistentState)
	if !getBoolAnnotation(cmd, modifiesConfigFile) &&
		!getBoolAnnotation(cmd, modifiesDestinationDirectory) &&
		!getBoolAnnotation(cmd, modifiesSourceDirectory) {
		c.system = chezmoi.NewReadOnlySystem(c.system)
	}
	if c.dryRun {
		c.system = chezmoi.NewDryRunSystem(c.system)
	}
	if c.debug {
		c.system = chezmoi.NewDebugSystem(c.system)
	}
	// FIXME verbose

	info, err := c.fs.Stat(c.SourceDir)
	switch {
	case err == nil && !info.IsDir():
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	case err == nil:
		if chezmoi.POSIXFileModes && info.Mode()&os.ModePerm&0o77 != 0 {
			cmd.Printf("%s: not private, but should be\n", c.SourceDir)
		}
	case !os.IsNotExist(err):
		return err
	}

	// Apply any fixes for snap, if needed.
	return c.snapFix()
}

//nolint:unparam
func (c *Config) prompt(s, choices string) (byte, error) {
	r := bufio.NewReader(c.Stdin)
	for {
		_, err := fmt.Printf("%s [%s]? ", s, strings.Join(strings.Split(choices, ""), ","))
		if err != nil {
			return 0, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if len(line) == 1 && strings.IndexByte(choices, line[0]) != -1 {
			return line[0], nil
		}
	}
}

//nolint:unparam
func (c *Config) run(dir, name string, args []string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		var err error
		cmd.Dir, err = c.fs.RawPath(dir)
		if err != nil {
			return err
		}
	}
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stdout
	// FIXME use c.system
	return cmd.Run()
}

func (c *Config) runEditor(args []string) error {
	editorName, editorArgs := getEditor()
	return c.run("", editorName, append(editorArgs, args...))
}

func (c *Config) marshal(data interface{}) error {
	format, ok := Formats[strings.ToLower(c.Format)]
	if !ok {
		return fmt.Errorf("unknown format: %s", c.Format)
	}
	marshaledData, err := format.Marshal(data)
	if err != nil {
		return err
	}
	return c.writeOutput(marshaledData)
}

func (c *Config) validateData() error {
	return validateKeys(c.Data, identifierRegexp)
}

func (c *Config) writeOutput(data []byte) error {
	if c.output == "" || c.output == "-" {
		_, err := c.Stdout.Write(data)
		return err
	}
	return c.fs.WriteFile(c.output, data, 0o666)
}

func (c *Config) writeOutputString(data string) error {
	return c.writeOutput([]byte(data))
}

func getDefaultConfigFile(bds *xdg.BaseDirectorySpecification) string {
	// Search XDG Base Directory Specification config directories first.
	for _, configDir := range bds.ConfigDirs {
		for _, extension := range viper.SupportedExts {
			configFilePath := filepath.Join(configDir, "chezmoi", "chezmoi."+extension)
			if _, err := os.Stat(configFilePath); err == nil {
				return configFilePath
			}
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return filepath.Join(bds.ConfigHome, "chezmoi", "chezmoi.toml")
}

func getDefaultSourceDir(bds *xdg.BaseDirectorySpecification) string {
	// Check for XDG Base Directory Specification data directories first.
	for _, dataDir := range bds.DataDirs {
		sourceDir := filepath.Join(dataDir, "chezmoi")
		if _, err := os.Stat(sourceDir); err == nil {
			return sourceDir
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return filepath.Join(bds.DataHome, "chezmoi")
}

func getEditor() (string, []string) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	components := whitespaceRegexp.Split(editor, -1)
	return components[0], components[1:]
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func serializationFormatNamesStr() string {
	names := make([]string, 0, len(Formats))
	for name := range Formats {
		names = append(names, strings.ToLower(name))
	}
	sort.Strings(names)
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " or " + names[1]
	default:
		names[len(names)-1] = "or " + names[len(names)-1]
		return strings.Join(names, ", ")
	}
}

// titleize returns s, titleized.
func titleize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	return string(append([]rune{unicode.ToTitle(runes[0])}, runes[1:]...))
}

// upperSnakeCaseToCamelCase converts a string in UPPER_SNAKE_CASE to
// camelCase.
func upperSnakeCaseToCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else if !isWellKnownAbbreviation(word) {
			words[i] = titleize(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// upperSnakeCaseToCamelCaseKeys returns m with all keys converted from
// UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}

// validateKeys ensures that all keys in data match re.
func validateKeys(data interface{}, re *regexp.Regexp) error {
	switch data := data.(type) {
	case map[string]interface{}:
		for key, value := range data {
			if !re.MatchString(key) {
				return fmt.Errorf("invalid key: %q", key)
			}
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, value := range data {
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	}
	return nil
}
