package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

const eolStr = "\r\n"

// enableVirtualTerminalProcessing enables virtual terminal processing. See
// https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences.
func enableVirtualTerminalProcessing(w io.Writer) error {
	f, ok := w.(*os.File)
	if !ok {
		return nil
	}
	var dwMode uint32
	if err := windows.GetConsoleMode(windows.Handle(f.Fd()), &dwMode); err != nil {
		return nil // Ignore error in the case that fd is not a terminal.
	}
	return windows.SetConsoleMode(windows.Handle(f.Fd()), dwMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func getUmask() int {
	return 0
}

func trimExecutableSuffix(s string) string {
	if strings.EqualFold(filepath.Ext(s), ".exe") {
		return s[:len(s)-4]
	}
	return s
}
