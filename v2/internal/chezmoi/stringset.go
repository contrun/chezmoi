package chezmoi

// A stringSet is a set of strings.
type stringSet map[string]struct{}

// newStringSet returns a new StringSet containing elements.
func newStringSet(elements ...string) stringSet {
	s := make(stringSet)
	s.Add(elements...)
	return s
}

// Add adds elements to s.
func (s stringSet) Add(elements ...string) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// Elements returns all the elements of s.
func (s stringSet) Elements() []string {
	elements := make([]string, 0, len(s))
	for element := range s {
		elements = append(elements, element)
	}
	return elements
}
