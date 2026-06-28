package cavy

import "github.com/Atharv3221/cavy/internal"

func NewCache[K comparable, V any](policy Policy, capacity int) Cache[K, V] {
	switch policy {
	case LRU:
		return internal.NewLRU[K, V](capacity)
	case LFU:
		return internal.NewLFU[K, V](capacity)
	default:
		panic("cavy: invalid cache policy")
	}
}
