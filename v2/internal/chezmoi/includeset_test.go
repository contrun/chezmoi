package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncludeMaskSet(t *testing.T) {
	for _, tc := range []struct {
		s           string
		expected    *IncludeBits
		expectedErr bool
	}{
		{
			s:        "",
			expected: NewIncludeBits(IncludeNone),
		},
		{
			s:        "none",
			expected: NewIncludeBits(IncludeNone),
		},
		{
			s:        "dirs,files",
			expected: NewIncludeBits(IncludeDirs | IncludeFiles),
		},
		{
			s:        "all",
			expected: NewIncludeBits(IncludeAll),
		},
		{
			s:        "all,!scripts",
			expected: NewIncludeBits(IncludeAbsent | IncludeDirs | IncludeFiles | IncludeSymlinks),
		},
		{
			s:        "a,s",
			expected: NewIncludeBits(IncludeAbsent | IncludeSymlinks),
		},
		{
			s:        "symlinks,,",
			expected: NewIncludeBits(IncludeSymlinks),
		},
		{
			s:           "devices",
			expectedErr: true,
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			actual := NewIncludeBits(IncludeNone)
			err := actual.Set(tc.s)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestIncludeMaskStringSlice(t *testing.T) {
	for _, tc := range []struct {
		includeBits IncludeBits
		expected    string
	}{
		{
			includeBits: IncludeAll,
			expected:    "all",
		},
		{
			includeBits: IncludeAbsent,
			expected:    "absent",
		},
		{
			includeBits: IncludeDirs,
			expected:    "dirs",
		},
		{
			includeBits: IncludeFiles,
			expected:    "files",
		},
		{
			includeBits: IncludeScripts,
			expected:    "scripts",
		},
		{
			includeBits: IncludeSymlinks,
			expected:    "symlinks",
		},
		{
			includeBits: IncludeNone,
			expected:    "none",
		},
		{
			includeBits: IncludeDirs | IncludeFiles,
			expected:    "dirs,files",
		},
	} {
		assert.Equal(t, tc.expected, tc.includeBits.String())
	}
}
