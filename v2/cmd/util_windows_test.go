//+build windows

package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimExecutableSuffix(t *testing.T) {
	for _, filename := range []string{
		"filename",
		"filename.exe",
		"filename.EXE",
	} {
		assert.Equal(t, "filename", trimExecutableSuffix(filename))
	}
}

func lines(s string) string {
	return strings.Replace(s, "\n", "\r\n", -1)
}
