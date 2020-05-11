package cmd

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, want := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, want, upperSnakeCaseToCamelCase(s))
	}
}

func TestValidateKeys(t *testing.T) {
	for _, tc := range []struct {
		data    interface{}
		wantErr bool
	}{
		{
			data:    nil,
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo":                    "bar",
				"a":                      0,
				"_x9":                    false,
				"ThisVariableIsExported": nil,
				"αβ":                     "",
			},
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo-foo": "bar",
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar-bar": "baz",
				},
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar-bar": "baz",
					},
				},
			},
			wantErr: true,
		},
	} {
		if tc.wantErr {
			assert.Error(t, validateKeys(tc.data, identifierRegexp))
		} else {
			assert.NoError(t, validateKeys(tc.data, identifierRegexp))
		}
	}
}

func newTestConfig(fs vfs.FS, options ...configOption) *Config {
	return newConfig(append(
		[]configOption{
			withTestFS(fs),
			withTestUser("user"),
		},
		options...,
	)...)
}

func withAddCmdConfig(add addCmdConfig) configOption {
	return func(c *Config) {
		c.add = add
	}
}

func withData(data map[string]interface{}) configOption {
	return func(c *Config) {
		c.Data = data
	}
}

func withDestDir(destDir string) configOption {
	return func(c *Config) {
		c.DestDir = destDir
	}
}

func withFollow(follow bool) configOption {
	return func(c *Config) {
		c.Follow = follow
	}
}

func withGenericSecretCmdConfig(genericSecretCmdConfig genericSecretCmdConfig) configOption {
	return func(c *Config) {
		c.GenericSecret = genericSecretCmdConfig
	}
}

func withSystem(system chezmoi.System) configOption {
	return func(c *Config) {
		c.system = system
	}
}

func withRemove(remove bool) configOption {
	return func(c *Config) {
		c.Remove = remove
	}
}

func withStdin(stdin io.Reader) configOption {
	return func(c *Config) {
		c.Stdin = stdin
	}
}

func withStdout(stdout io.WriteCloser) configOption {
	return func(c *Config) {
		c.Stdout = stdout
	}
}

func withTestFS(fs vfs.FS) configOption {
	return func(c *Config) {
		c.fs = fs
		// c.system = chezmoi.NewRealSystem() // FIXME
		c.Verbose = true
	}
}

func withTestUser(username string) configOption {
	return func(c *Config) {
		homeDir := filepath.Join("/", "home", username)
		c.SourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
		c.DestDir = homeDir
		c.Umask = 0o22
		c.bds = &xdg.BaseDirectorySpecification{
			ConfigHome: filepath.Join(homeDir, ".config"),
			DataHome:   filepath.Join(homeDir, ".local"),
			CacheHome:  filepath.Join(homeDir, ".cache"),
			RuntimeDir: filepath.Join(homeDir, ".run"),
		}
	}
}
