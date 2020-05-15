package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/v2/cmd"
)

//nolint:interfacer
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"chezmoi": func() int {
			if err := cmd.Execute(cmd.VersionInfo{
				Version: "v2.0.0",
				Commit:  "HEAD",
				Date:    time.Now().Format(time.RFC3339),
				BuiltBy: "testscript",
			}); err != nil {
				if s := err.Error(); s != "" {
					fmt.Printf("chezmoi: %s\n", s)
				}
				return 1
			}
			return 0
		},
	}))
}

func TestChezmoi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"chhome":      cmdChHome,
			"edit":        cmdEdit,
			"mkfile":      cmdMkFile,
			"mkhomedir":   cmdMkHomeDir,
			"mksourcedir": cmdMkSourceDir,
		},
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "windows":
				return runtime.GOOS == "windows", nil
			case "v2":
				return true, nil
			default:
				return false, fmt.Errorf("%s: unknown condition", cond)
			}
		},
		Setup:         setup,
		UpdateScripts: os.Getenv("CHEZMOIUPDATESCRIPTS") != "",
	})
}

// cmdChHome changes the home directory to its argument, creating the directory
// if it does not already exists. It updates the HOME environment variable, and,
// if running on Windows, USERPROFILE too.
func cmdChHome(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! chhome")
	}
	if len(args) != 1 {
		ts.Fatalf("usage: chhome dir")
	}
	homeDir := ts.MkAbs(args[0])
	ts.Check(os.MkdirAll(homeDir, 0o777))
	ts.Setenv("HOME", homeDir)
	ts.Setenv("HOMESLASH", filepath.ToSlash(homeDir))
	if runtime.GOOS == "windows" {
		ts.Setenv("USERPROFILE", homeDir)
	}
}

// cmdEdit edits all of its arguments by appending "# edited\n" to them.
func cmdEdit(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! edit")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			ts.Fatalf("edit: %v", err)
		}
		data = append(data, []byte("# edited\n")...)
		//nolint:gosec
		if err := ioutil.WriteFile(filename, data, 0o666); err != nil {
			ts.Fatalf("edit: %v", err)
		}
	}
}

// cmdMkFile creates empty files.
func cmdMkFile(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkfile")
	}
	perm := os.FileMode(0o666)
	if len(args) >= 1 && strings.HasPrefix(args[0], "-perm=") {
		permStr := strings.TrimPrefix(args[0], "-perm=")
		permUint32, err := strconv.ParseUint(permStr, 8, 32)
		if err != nil {
			ts.Fatalf("%s: bad permissions", permStr)
		}
		perm = os.FileMode(permUint32)
		args = args[1:]
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		_, err := os.Lstat(filename)
		switch {
		case err == nil:
			ts.Fatalf("%s: already exists", arg)
		case !os.IsNotExist(err):
			ts.Fatalf("%s: %v", arg, err)
		}
		if err := ioutil.WriteFile(filename, nil, perm); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
	}
}

// cmdMkHomeDir makes and populates a home directory.
func cmdMkHomeDir(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkhomedir")
	}
	if len(args) > 1 {
		ts.Fatalf(("usage: mkhomedir [path]"))
	}
	path := ts.Getenv("HOME")
	if len(args) > 0 {
		path = args[0]
	}
	if err := vfst.NewBuilder().Build(vfs.HostOSFS, map[string]interface{}{
		ts.MkAbs(path): map[string]interface{}{
			".bashrc": "# contents of .bashrc\n",
			".binary": &vfst.File{
				Perm:     0o755,
				Contents: []byte("#/bin/sh\n"),
			},
			".gitconfig": "[user]\n  email = user@home.org\n",
			".hushlogin": "",
			".ssh": &vfst.Dir{
				Perm: 0o700,
				Entries: map[string]interface{}{
					"config": "# contents of .ssh/config\n",
				},
			},
			".symlink": &vfst.Symlink{
				Target: ".bashrc",
			},
		}}); err != nil {
		ts.Fatalf("mkhomedir: %v", err)
	}
}

// cmdMkSourceDir makes and populates a source directory.
func cmdMkSourceDir(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mksourcedir")
	}
	if len(args) != 0 {
		ts.Fatalf(("usage: mksourcedir"))
	}
	err := vfst.NewBuilder().Build(vfs.HostOSFS, map[string]interface{}{
		ts.MkAbs(ts.Getenv("CHEZMOISOURCEDIR")): map[string]interface{}{
			"dot_absent":            "",
			"empty_dot_hushlogin":   "",
			"executable_dot_binary": "#/bin/sh\n",
			"dot_bashrc":            "# contents of .bashrc\n",
			"dot_gitconfig.tmpl":    "[user]\n  email = {{ \"user@home.org\" }}\n",
			"private_dot_ssh": map[string]interface{}{
				"config": "# contents of .ssh/config\n",
			},
			"run_script":          "#/bin/sh\necho script\n",
			"symlink_dot_symlink": ".bashrc\n",
		},
	})
	if err != nil {
		ts.Fatalf("mksourcedir: %v", err)
	}
}

func setup(env *testscript.Env) error {
	var (
		binDir           = filepath.Join(env.WorkDir, "bin")
		homeDir          = filepath.Join(env.WorkDir, "home", "user")
		chezmoiConfigDir = filepath.Join(homeDir, ".config", "chezmoi")
		chezmoiSourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
	)

	env.Setenv("HOME", homeDir)
	env.Setenv("HOMESLASH", filepath.ToSlash(homeDir))
	env.Setenv("PATH", prependDirToPath(binDir, env.Getenv("PATH")))
	env.Setenv("CHEZMOICONFIGDIR", chezmoiConfigDir)
	env.Setenv("CHEZMOISOURCEDIR", chezmoiSourceDir)
	switch runtime.GOOS {
	case "windows":
		env.Setenv("EDITOR", filepath.Join(binDir, "editor.cmd"))
		env.Setenv("USERPROFILE", homeDir)
		env.Setenv("WORKSLASH", filepath.ToSlash(env.WorkDir))
		// There is not currently a convenient way to override the shell on
		// Windows.
	default:
		env.Setenv("EDITOR", filepath.Join(binDir, "editor"))
		env.Setenv("SHELL", filepath.Join(binDir, "shell"))
		env.Setenv("WORKSLASH", env.WorkDir)
	}

	// Fix permissions on the source directory, if it exists.
	_ = os.Chmod(chezmoiSourceDir, 0o700)

	// Fix permissions on any files in the bin directory.
	infos, err := ioutil.ReadDir(binDir)
	if err == nil {
		for _, info := range infos {
			if err := os.Chmod(filepath.Join(binDir, info.Name()), 0o755); err != nil {
				return err
			}
		}
	}

	root := make(map[string]interface{})
	switch runtime.GOOS {
	case "windows":
		root["/bin"] = map[string]interface{}{
			// editor.cmd a non-interactive script that appends "# edited\n" to
			// the end of each file.
			"editor.cmd": &vfst.File{
				Perm:     0o755,
				Contents: []byte(`@for %%x in (%*) do echo # edited>>%%x`),
			},
		}
	default:
		root["/bin"] = map[string]interface{}{
			// editor a non-interactive script that appends "# edited\n" to the
			// end of each file.
			"editor": &vfst.File{
				Perm: 0o755,
				Contents: []byte(strings.Join([]string{
					`#!/bin/sh`,
					``,
					`for filename in $*; do`,
					`    echo "# edited" >> $filename`,
					`done`,
				}, "\n")),
			},
			// shell is a non-interactive script that appends the directory in
			// which it was launched to $WORK/shell.log.
			"shell": &vfst.File{
				Perm: 0o755,
				Contents: []byte(strings.Join([]string{
					`#!/bin/sh`,
					``,
					`echo $PWD >> ` + filepath.Join(env.WorkDir, "shell.log"),
				}, "\n")),
			},
		}
	}

	return vfst.NewBuilder().Build(vfs.NewPathFS(vfs.HostOSFS, env.WorkDir), root)
}

func prependDirToPath(dir, path string) string {
	return strings.Join(append([]string{dir}, filepath.SplitList(path)...), string(os.PathListSeparator))
}
