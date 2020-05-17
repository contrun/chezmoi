package chezmoi

import (
	"fmt"
	"strings"
)

// An IncludeBits controls what types of entries to include. It parses and
// prints as a comma-separated list of strings, but is internally represented as
// a bitmask. *IncludeBits implements the github.com/spf13/pflag.Value
// interface.
type IncludeBits int

// Include bits.
const (
	IncludeAbsent IncludeBits = 1 << iota
	IncludeDirs
	IncludeFiles
	IncludeScripts
	IncludeSymlinks

	IncludeAll = IncludeAbsent | IncludeDirs | IncludeFiles | IncludeScripts | IncludeSymlinks

	IncludeNone IncludeBits = 0
)

var includeBits = map[string]IncludeBits{
	"a":        IncludeAbsent,
	"absent":   IncludeAbsent,
	"all":      IncludeAll,
	"d":        IncludeDirs,
	"dirs":     IncludeDirs,
	"f":        IncludeFiles,
	"files":    IncludeFiles,
	"scripts":  IncludeScripts,
	"s":        IncludeSymlinks,
	"symlinks": IncludeSymlinks,
}

// NewIncludeBits returns a new IncludeBits.
func NewIncludeBits(includeBits IncludeBits) *IncludeBits {
	return &includeBits
}

// Include returns true if v should be included.
func (m *IncludeBits) Include(v interface{}) bool {
	switch v.(type) {
	case *TargetStateAbsent:
		return *m&IncludeAbsent != 0
	case *DestStateDir, *SourceStateDir, *TargetStateDir:
		return *m&IncludeDirs != 0
	case *DestStateFile, *SourceStateFile, *TargetStateFile:
		return *m&IncludeFiles != 0
	case *TargetStateScript:
		return *m&IncludeScripts != 0
	case *DestStateSymlink, *TargetStateSymlink:
		return *m&IncludeSymlinks != 0
	default:
		panic(fmt.Sprintf("%T: unsupported type", v))
	}
}

// Set implements github.com/spf13/pflag.Value.Set.
func (m *IncludeBits) Set(s string) error {
	if s == "none" {
		*m = IncludeNone
		return nil
	}

	v := IncludeNone
	for _, s := range strings.Split(s, ",") {
		if s == "" {
			continue
		}
		exclude := false
		if strings.HasPrefix(s, "!") {
			exclude = true
			s = s[1:]
		}
		bit, ok := includeBits[s]
		if !ok {
			return fmt.Errorf("%s: unknown include element", s)
		}
		if exclude {
			v &^= bit
		} else {
			v |= bit
		}
	}
	*m = v
	return nil
}

func (m *IncludeBits) String() string {
	switch *m {
	case IncludeAll:
		return "all"
	case IncludeNone:
		return "none"
	}
	var ss []string
	for i, s := range []string{
		"absent",
		"dirs",
		"files",
		"scripts",
		"symlinks",
	} {
		if *m&(1<<i) != 0 {
			ss = append(ss, s)
		}
	}
	return strings.Join(ss, ",")
}

// Type implements github.com/spf13/pflag.Value.Type.
func (m *IncludeBits) Type() string {
	return "include set"
}
