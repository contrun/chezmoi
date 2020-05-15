package cmd

import (
	"fmt"
	"os"
	"strconv"
)

// An permFlag represents permissions. It implements the
// github.com/spf13/pflag.Value interface for use as a command line flag.
type permFlag os.FileMode

func (p *permFlag) Set(s string) error {
	v, err := strconv.ParseUint(s, 8, 32)
	if os.FileMode(v)&os.ModePerm != os.FileMode(v) {
		return fmt.Errorf("%s: invalid mode", s)
	}
	*p = permFlag(v)
	return err
}

func (p *permFlag) String() string {
	return fmt.Sprintf("%03o", *p)
}

func (p *permFlag) Type() string {
	return "mode"
}
