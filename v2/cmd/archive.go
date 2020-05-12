package cmd

// FIXME add -j option to bzip2 compress output
// FIXME add -z option to gzip compress output

import (
	"archive/tar"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var archiveCmd = &cobra.Command{
	Use:     "archive [targets...]",
	Short:   "Generate a tar archive of the target state",
	Long:    mustGetLongHelp("archive"),
	Example: getExample("archive"),
	PreRunE: config.ensureNoError,
	RunE:    config.runArchiveCmd,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	c.readOnly()

	sb := &strings.Builder{}
	tarSystem := chezmoi.NewTARSystem(sb, tarHeaderTemplate())
	if err := c.applyArgs(tarSystem, "", args); err != nil {
		return err
	}
	if err := tarSystem.Close(); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}

// tarHeaderTemplate returns a tar.Header template populated with the current
// user and time.
func tarHeaderTemplate() tar.Header {
	// Attempt to lookup the current user. Ignore errors because the default
	// zero values are reasonable.
	var (
		uid   int
		gid   int
		Uname string
		Gname string
	)
	if currentUser, err := user.Current(); err == nil {
		uid, _ = strconv.Atoi(currentUser.Uid)
		gid, _ = strconv.Atoi(currentUser.Gid)
		Uname = currentUser.Username
		if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
			Gname = group.Name
		}
	}

	now := time.Now()
	return tar.Header{
		Uid:        uid,
		Gid:        gid,
		Uname:      Uname,
		Gname:      Gname,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
}
