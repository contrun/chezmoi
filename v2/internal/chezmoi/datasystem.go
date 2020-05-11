package chezmoi

import (
	"os"
)

type dataType string

const (
	dataTypeDir     dataType = "dir"
	dataTypeFile    dataType = "file"
	dataTypeScript  dataType = "script"
	dataTypeSymlink dataType = "symlink"
)

// A DataSystem is a System that writes to a data file.
type DataSystem struct {
	nullSystem
	data map[string]interface{}
}

type dirData struct {
	Type dataType    `json:"type" toml:"type" yaml:"type"`
	Name string      `json:"name" toml:"name" yaml:"name"`
	Perm os.FileMode `json:"perm" toml:"perm" yaml:"perm"`
}

type fileData struct {
	Type     dataType    `json:"type" toml:"type" yaml:"type"`
	Name     string      `json:"name" toml:"name" yaml:"name"`
	Contents string      `json:"contents" toml:"contents" yaml:"contents"`
	Perm     os.FileMode `json:"perm" toml:"perm" yaml:"perm"`
}

type scriptData struct {
	Type     dataType `json:"type" toml:"type" yaml:"type"`
	Name     string   `json:"name" toml:"name" yaml:"name"`
	Contents string   `json:"contents" toml:"contents" yaml:"contents"`
}

type symlinkData struct {
	Type     dataType `json:"type" toml:"type" yaml:"type"`
	Name     string   `json:"name" toml:"name" yaml:"name"`
	Linkname string   `json:"linkname" toml:"linkname" yaml:"linkname"`
}

// NewDataSystem returns a new DataSystem that accumulates data.
func NewDataSystem() *DataSystem {
	return &DataSystem{
		data: make(map[string]interface{}),
	}
}

// Chmod implements System.Chmod.
func (s *DataSystem) Chmod(name string, mode os.FileMode) error {
	return os.ErrPermission
}

// Data returns s's data.
func (s *DataSystem) Data() interface{} {
	return s.data
}

// Mkdir implements System.Mkdir.
func (s *DataSystem) Mkdir(dirname string, perm os.FileMode) error {
	if _, exists := s.data[dirname]; exists {
		return os.ErrExist
	}
	s.data[dirname] = &dirData{
		Type: dataTypeDir,
		Name: dirname,
		Perm: perm,
	}
	return nil
}

// RemoveAll implements System.RemoveAll.
func (s *DataSystem) RemoveAll(name string) error {
	return os.ErrPermission
}

// RunScript implements System.RunScript.
func (s *DataSystem) RunScript(scriptname string, data []byte) error {
	if _, exists := s.data[scriptname]; exists {
		return os.ErrExist
	}
	s.data[scriptname] = &scriptData{
		Type:     dataTypeScript,
		Name:     scriptname,
		Contents: string(data),
	}
	return nil
}

// WriteFile implements System.WriteFile.
func (s *DataSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if _, exists := s.data[filename]; exists {
		return os.ErrExist
	}
	s.data[filename] = &fileData{
		Type:     dataTypeFile,
		Name:     filename,
		Contents: string(data),
		Perm:     perm,
	}
	return nil
}

// WriteSymlink implements System.WriteSymlink.
func (s *DataSystem) WriteSymlink(oldname, newname string) error {
	if _, exists := s.data[newname]; exists {
		return os.ErrExist
	}
	s.data[newname] = &symlinkData{
		Type:     dataTypeSymlink,
		Name:     newname,
		Linkname: oldname,
	}
	return nil
}
