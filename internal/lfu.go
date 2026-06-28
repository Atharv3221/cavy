package internal

import (
	"container/list"

	"github.com/Atharv3221/cavy/utils"
)

// lfu_node is a single cache entry tracked within a frequency bucket.
type lfu_node[K comparable, V any] struct {
	key   K
	value V
	freq  int
}

// lfu implements an O(1) Least-Frequently-Used eviction policy.
type lfu[K comparable, V any] struct {
	capacity int
	minFreq  int
	items    map[K]*list.Element
	freqs    map[int]*list.List
}

// NewLFU creates an LFU cache with the given capacity.
func NewLFU[K comparable, V any](capacity int) *lfu[K, V] {
	if capacity <= 0 {
		capacity = DefaultMaxCapacity
	}
	return &lfu[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		freqs:    make(map[int]*list.List),
	}
}

// Len returns the number of items currently stored.
func (l *lfu[K, V]) Len() int {
	return len(l.items)
}

// Get returns the value for key and bumps its frequency by one.
func (l *lfu[K, V]) Get(key K) (V, bool) {
	element, ok := l.items[key]
	if !ok {
		var value V
		return value, false
	}
	node := element.Value.(*lfu_node[K, V])
	l.touch(element, node)
	return node.value, true
}

// Put inserts or updates key.
func (l *lfu[K, V]) Put(key K, value V) {
	if l.Contains(key) {
		element := l.items[key]
		node := element.Value.(*lfu_node[K, V])
		node.value = value
		l.touch(element, node)
		return
	}

	if len(l.items) >= l.capacity {
		l.evict()
	}

	entry := &lfu_node[K, V]{key: key, value: value, freq: 1}
	bucket := l.bucket(1)
	l.items[key] = bucket.PushFront(entry)
	l.minFreq = 1
}

// Checks if key is present in the items
func (l *lfu[K, V]) Contains(key K) bool {
	if _, exits := l.items[key]; exits {
		return true
	}
	return false
}

// Delete removes key from the cache, if present.
func (l *lfu[K, V]) Delete(key K) bool {
	if !l.Contains(key) {
		return false
	}
	element, _ := l.items[key]
	entry := element.Value.(*lfu_node[K, V])
	l.removeFromBucket(element, entry.freq)
	delete(l.items, key)
	l.updateMinFrequency(l.minFreq)
	return true
}

func (l *lfu[K, V]) SetCapacity(capacity int) bool {
	var result bool
	l.capacity, result = utils.SetMaxCapacity(l.capacity, capacity, l.Len())
	return result
}

// Returns max capacity of the lfu
func (l *lfu[K, V]) GetMaxCapacity() int {
	return l.capacity
}

// Return all key, value pair in map
func (l *lfu[K, V]) GetAll() map[K]V {
	result := make(map[K]V, l.Len())

	for key, element := range l.items {
		entry := element.Value.(*lfu_node[K, V])
		result[key] = entry.value
	}

	return result
}

// Moves entry from its current frequency bucket to freq + 1,
// and advances minFreq if the old bucket just became empty.
func (l *lfu[K, V]) touch(element *list.Element, entry *lfu_node[K, V]) {
	oldFreq := entry.freq
	l.removeFromBucket(element, oldFreq)

	entry.freq++
	bucket := l.bucket(entry.freq)
	l.items[entry.key] = bucket.PushFront(entry)

	l.updateMinFrequency(oldFreq)
}

// evict drops the least-recently-used entry from the lowest
// frequency bucket
func (l *lfu[K, V]) evict() {
	bucket, _ := l.freqs[l.minFreq]
	back := bucket.Back()
	entry := back.Value.(*lfu_node[K, V])
	bucket.Remove(back)
	if bucket.Len() == 0 {
		delete(l.freqs, l.minFreq)
	}
	delete(l.items, entry.key)
	l.updateMinFrequency(entry.freq)
}

// Return list corresponding to the frequeny
func (l *lfu[K, V]) bucket(freq int) *list.List {
	bucket, exists := l.freqs[freq]
	if !exists {
		bucket = list.New()
		l.freqs[freq] = bucket
	}
	return bucket
}

func (l *lfu[K, V]) removeFromBucket(element *list.Element, freq int) {
	bucket := l.freqs[freq]

	bucket.Remove(element)

	if bucket.Len() == 0 {
		delete(l.freqs, freq)
	}
}

func (l *lfu[K, V]) updateMinFrequency(oldFreq int) {
	if oldFreq != l.minFreq {
		return
	}

	if len(l.items) == 0 {
		l.minFreq = 0
		return
	}

	for {
		if _, ok := l.freqs[l.minFreq]; ok {
			return
		}
		l.minFreq++
	}
}
