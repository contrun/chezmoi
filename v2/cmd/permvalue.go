package cmd

import (
	"fmt"
	"strconv"
)

// An permFlag represents permissions. It implements the pflag.Value interface
// for use as a command line flag.
type permFlag uint

func (p *permFlag) Set(s string) error {
	v, err := strconv.ParseUint(s, 8, 64)
	*p = permFlag(v)
	return err
}

func (p *permFlag) String() string {
	return fmt.Sprintf("%03o", *p)
}

func (p *permFlag) Type() string {
	return "uint"
}
