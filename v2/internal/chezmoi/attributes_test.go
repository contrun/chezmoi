package chezmoi

import (
	"testing"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDirAttributes tests DirAttributes by round-tripping between directory
// names and DirAttributes.
func TestDirAttributes(t *testing.T) {
	testData := struct {
		Name    []string
		Exact   []bool
		Private []bool
	}{
		Name: []string{
			".dir",
			"dir.tmpl",
			"dir",
			"empty_dir",
			"encrypted_dir",
			"executable_dir",
			"once_dir",
			"run_dir",
			"run_once_dir",
			"symlink_dir",
		},
		Exact:   []bool{false, true},
		Private: []bool{false, true},
	}
	var dirAttributes []DirAttributes
	require.NoError(t, combinator.Generate(&dirAttributes, testData))
	for _, da := range dirAttributes {
		actualSourceName := da.SourceName()
		actualDA := ParseDirAttributes(actualSourceName)
		assert.Equal(t, da, actualDA)
		assert.Equal(t, actualSourceName, actualDA.SourceName())
	}
}

// TestFileAttributes tests FileAttributes by round-tripping between file names
// and FileAttributes.
func TestFileAttributes(t *testing.T) {
	var fileAttributes []FileAttributes
	require.NoError(t, combinator.Generate(&fileAttributes, struct {
		Type       SourceFileTargetType
		Name       []string
		Empty      []bool
		Encrypted  []bool
		Executable []bool
		Private    []bool
		Template   []bool
	}{
		Type: SourceFileTypeFile,
		Name: []string{
			".name",
			"exact_name",
			"name",
		},
		Empty:      []bool{false, true},
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		Template:   []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fileAttributes, struct {
		Type SourceFileTargetType
		Name []string
		Once []bool
	}{
		Type: SourceFileTypeScript,
		Name: []string{
			"exact_name",
			"name",
		},
		Once: []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fileAttributes, struct {
		Type SourceFileTargetType
		Name []string
		Once []bool
	}{
		Type: SourceFileTypeSymlink,
		Name: []string{
			"exact_name",
			"name",
		},
	}))
	for _, fa := range fileAttributes {
		actualSourceName := fa.SourceName()
		actualFA := ParseFileAttributes(actualSourceName)
		assert.Equal(t, fa, actualFA)
		assert.Equal(t, actualSourceName, actualFA.SourceName())
	}
}
