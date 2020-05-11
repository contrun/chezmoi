package cmd

import (
	"github.com/pelletier/go-toml"
)

type tomlFormat struct{}

// TOMLFormat is the TOML serialization format.
var TOMLFormat tomlFormat

func (tomlFormat) Name() string {
	return "toml"
}

func (tomlFormat) Marshal(data interface{}) ([]byte, error) {
	return toml.Marshal(data)
}

func (tomlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	Formats[TOMLFormat.Name()] = TOMLFormat
}
