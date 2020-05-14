package cmd

import (
	"fmt"
	"strconv"
)

// An permValue is an uint that is scanned and printed in octal. It implements
// the pflag.Value interface for use as a command line flag.
type permValue uint

func (p *permValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 8, 64)
	*p = permValue(v)
	return err
}

func (p *permValue) String() string {
	return fmt.Sprintf("%03o", *p)
}

func (p *permValue) Type() string {
	return "uint"
}
