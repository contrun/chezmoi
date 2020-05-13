//+build windows

package cmd

import (
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
