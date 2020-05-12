// +build !windows

package cmd

import (
	"io"
	"syscall"
)

const eolStr = "\n"

// enableVirtualTerminalProcessingOnWindows does nothing on POSIX systems.
func enableVirtualTerminalProcessingOnWindows(w io.Writer) error {
	return nil
}

func getUmask() int {
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return umask
}

func trimExecutableSuffix(s string) string {
	return s
}
