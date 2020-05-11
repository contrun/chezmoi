package chezmoi

// A PersistentState is a persistent state.
type PersistentState interface {
	Get(bucket, key []byte) ([]byte, error)
	Delete(bucket, key []byte) error
	Set(bucket, key, value []byte) error
}
