package chezmoi

import (
	"os"
	"os/exec"
)

// A System reads from and writes to a filesystem, executes idempotent commands,
// runs scripts, and persists state.
type System interface {
	PersistentState
	Chmod(name string, mode os.FileMode) error
	Glob(pattern string) ([]string, error)
	IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)
	Lstat(filename string) (os.FileInfo, error)
	Mkdir(name string, perm os.FileMode) error
	ReadDir(dirname string) ([]os.FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	Readlink(name string) (string, error)
	RemoveAll(name string) error
	Rename(oldpath, newpath string) error
	RunScript(scriptname string, data []byte) error
	Stat(name string) (os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	WriteSymlink(oldname, newname string) error
}
