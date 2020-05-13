package chezmoi

import (
	"gopkg.in/yaml.v2"
)

type yamlFormat struct{}

// YAMLFormat is the YAML serialization format.
var YAMLFormat yamlFormat

func (yamlFormat) Name() string {
	return "yaml"
}

func (yamlFormat) Marshal(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

func (yamlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	Formats[YAMLFormat.Name()] = YAMLFormat
}
