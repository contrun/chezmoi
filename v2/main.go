//go:generate go run ../internal/generate-assets -o cmd/docs.gen.go -tags=!noembeddocs -trimprefix=../ ../docs/CHANGES.md ../docs/CONTRIBUTING.md ../docs/FAQ.md ../docs/HOWTO.md ../docs/INSTALL.md ../docs/MEDIA.md ../docs/QUICKSTART.md ../docs/REFERENCE.md
//go:generate go run ../internal/generate-assets -o cmd/templates.gen.go -trimprefix=../ ../assets/templates/COMMIT_MESSAGE.tmpl
//go:generate go run ../internal/generate-helps -o cmd/helps.gen.go -i ../docs/REFERENCE.md

package main

import (
	"fmt"
	"os"

	"github.com/twpayne/chezmoi/v2/cmd"
)

var (
	version = ""
	commit  = ""
	date    = ""
	builtBy = ""
)

func run() error {
	cmd.VersionStr = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.BuiltBy = builtBy
	return cmd.Execute()
}

func main() {
	if err := run(); err != nil {
		if s := err.Error(); s != "" {
			fmt.Printf("chezmoi: %s\n", s)
		}
		os.Exit(1)
	}
}
