package chezmoi

// A Format is a serialization format.
type Format interface {
	Marshal(data interface{}) ([]byte, error)
	Name() string
	Unmarshal(data []byte) (interface{}, error)
}

// Formats is a map of all Formats by name.
var Formats = make(map[string]Format)
