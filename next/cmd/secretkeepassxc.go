package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var keepassxcCmd = &cobra.Command{
	Use:   "keepassxc [args...]",
	Short: "Execute the Keepassxc CLI (keepassxc-cli)",
	RunE:  config.runKeepassxcCmd,
}

type keepassxcCmdConfig struct {
	Command  string
	Database string
	Args     []string
}

type keepassxcAttributeCacheKey struct {
	entry     string
	attribute string
}

var (
	keepassxcVersion                     *semver.Version
	keepassxcCache                       = make(map[string]map[string]string)
	keepassxcAttributeCache              = make(map[keepassxcAttributeCacheKey]string)
	keepassxcPairRegexp                  = regexp.MustCompile(`^([^:]+): (.*)$`)
	keepassxcPassword                    string
	keepassxcNeedShowProtectedArgVersion = semver.Version{Major: 2, Minor: 5, Patch: 1}
)

func init() {
	config.Keepassxc.Command = "keepassxc-cli"
	config.addTemplateFunc("keepassxc", config.keepassxcFunc)
	config.addTemplateFunc("keepassxcAttribute", config.keepassxcAttributeFunc)

	secretCmd.AddCommand(keepassxcCmd)
}

func (c *Config) runKeepassxcCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Keepassxc.Command, args)
}

func (c *Config) getKeepassxcVersion() *semver.Version {
	if keepassxcVersion != nil {
		return keepassxcVersion
	}
	name := c.Keepassxc.Command
	args := []string{"--version"}
	cmd := exec.Command(name, args...)
	output, err := c.system.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	keepassxcVersion, err = semver.NewVersion(string(bytes.TrimSpace(output)))
	if err != nil {
		panic(fmt.Errorf("keepassxc: cannot parse version %q: %w", output, err))
	}
	return keepassxcVersion
}

func (c *Config) keepassxcFunc(entry string) map[string]string {
	if data, ok := keepassxcCache[entry]; ok {
		return data
	}
	if c.Keepassxc.Database == "" {
		panic(errors.New("keepassxc: keepassxc.database not set"))
	}
	name := c.Keepassxc.Command
	args := []string{"show"}
	if c.getKeepassxcVersion().Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database, entry)
	output, err := c.runKeepassxcCLICommand(name, args)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	data, err := parseKeyPassXCOutput(output)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	keepassxcCache[entry] = data
	return data
}

func (c *Config) keepassxcAttributeFunc(entry, attribute string) string {
	key := keepassxcAttributeCacheKey{
		entry:     entry,
		attribute: attribute,
	}
	if data, ok := keepassxcAttributeCache[key]; ok {
		return data
	}
	if c.Keepassxc.Database == "" {
		panic(errors.New("keepassxc: keepassxc.database not set"))
	}
	name := c.Keepassxc.Command
	args := []string{"show", "--attributes", attribute, "--quiet"}
	if c.getKeepassxcVersion().Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database, entry)
	output, err := c.runKeepassxcCLICommand(name, args)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	outputStr := strings.TrimSpace(string(output))
	keepassxcAttributeCache[key] = outputStr
	return outputStr
}

func (c *Config) runKeepassxcCLICommand(name string, args []string) ([]byte, error) {
	if keepassxcPassword == "" {
		fmt.Printf("Insert password to unlock %s: ", c.Keepassxc.Database)
		password, err := terminal.ReadPassword(int(os.Stdout.Fd()))
		fmt.Println()
		if err != nil {
			return nil, err
		}
		keepassxcPassword = string(password)
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(keepassxcPassword + "\n")
	cmd.Stderr = c.stderr
	return c.system.IdempotentCmdOutput(cmd)
}

func parseKeyPassXCOutput(output []byte) (map[string]string, error) {
	data := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(output))
	for i := 0; s.Scan(); i++ {
		if i == 0 {
			continue
		}
		match := keepassxcPairRegexp.FindStringSubmatch(s.Text())
		if match == nil {
			return nil, fmt.Errorf("cannot parse %q", s.Text())
		}
		data[match[1]] = match[2]
	}
	return data, s.Err()
}
