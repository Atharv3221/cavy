# cavy

A lightweight, generic, pluggable in-memory cache library for Go, supporting multiple eviction policies (LRU, LFU) behind a single common interface.

> **Status:** Snapshot — unreleased, early stage. Core functionality is implemented and tested; concurrency safety, benchmarks, and additional policies are planned (see [Roadmap](#roadmap)).

## Why cavy?

Most cache libraries hard-code you into one eviction strategy. `cavy` separates the **interface** you code against from the **policy** you choose at construction time, so swapping LRU for LFU (or adding a new policy later) doesn't touch any calling code.

## Features

- Generic — works with any `comparable` key and any value type (`Cache[K comparable, V any]`)
- Pluggable eviction policies via a simple factory: `LRU`, `LFU`
- Minimal, predictable interface: `Get`, `Put`, `Delete`, `Contains`, `GetAll`, `Len`, `GetMaxCapacity`, `SetCapacity`
- `SetCapacity` is data-loss-safe: it refuses to shrink below the current item count rather than silently evicting items

## Installation

```bash
go get github.com/Atharv3221/cavy
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/Atharv3221/cavy"
)

func main() {
	// Create an LRU cache with a capacity of 2.
	cache := cavy.NewCache[string, int](cavy.LRU, 2)

	cache.Put("a", 1)
	cache.Put("b", 2)

	// "a" is evicted here once it's the least recently used and a 3rd item is added.
	cache.Put("c", 3)

	if value, ok := cache.Get("b"); ok {
		fmt.Println("b =", value)
	}

	fmt.Println("size:", cache.Len())
}
```

Switching to LFU is a one-line change:

```go
cache := cavy.NewCache[string, int](cavy.LFU, 2)
```

## API

```go
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

func NewCache[K comparable, V any](policy Policy, capacity int) Cache[K, V]
```

| Method | Behavior |
|---|---|
| `Get(key)` | Returns the value and `true` if present; touches the item for recency/frequency tracking |
| `Put(key, value)` | Inserts or updates; may trigger eviction if the cache is at capacity |
| `Delete(key)` | Removes the key if present; returns whether it was present |
| `Contains(key)` | Checks presence without affecting recency/frequency |
| `GetAll()` | Returns a snapshot copy of all entries |
| `Len()` | Current number of items |
| `GetMaxCapacity()` | Configured capacity |
| `SetCapacity(n)` | Resizes capacity; returns `false` (no-op) if `n` would require evicting existing items, or if `n <= 0` |

## Policies

| Policy | Eviction rule |
|---|---|
| `cavy.LRU` | Evicts the least recently used item |
| `cavy.LFU` | Evicts the least frequently used item |

## Testing

```bash
go test ./...
```

The test suite covers shared cache semantics (insert/update/delete/contains), capacity-bound eviction, policy-specific ordering (LRU recency, LFU frequency), and `SetCapacity`'s data-loss-prevention contract.

## Roadmap

- [ ] Concurrency-safe variant (`sync.RWMutex` or lock-free)
- [ ] Benchmarks (LRU vs LFU vs naive map)
- [ ] Additional policies (FIFO, ARC)
- [ ] CI (GitHub Actions: `go test`, `go vet`, lint)

## License

See [LICENSE](./LICENSE).
