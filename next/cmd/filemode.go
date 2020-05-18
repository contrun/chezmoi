package cmd

import (
	"fmt"
	"os"
	"strconv"
)

// A fileMode represents a file mode. It implements the
// github.com/spf13/pflag.Value interface for use as a command line flag.
type fileMode os.FileMode

func (p *fileMode) FileMode() os.FileMode {
	return os.FileMode(*p)
}

func (p *fileMode) Set(s string) error {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil || os.FileMode(v)&os.ModePerm != os.FileMode(v) {
		return fmt.Errorf("%s: invalid mode", s)
	}
	*p = fileMode(v)
	return nil
}

func (p *fileMode) String() string {
	return fmt.Sprintf("%03o", *p)
}

func (p *fileMode) Type() string {
	return "mode"
}
