package cavy

// Cache interface to support different cache policies
type Cache[K comparable, V any] interface {
	Contains(key K) bool

	Get(key K) (V, bool)

	GetAll() map[K]V

	GetMaxCapacity() int

	Put(key K, value V)

	SetCapacity(capacity int) bool

	Delete(key K) bool

	Len() int
}
