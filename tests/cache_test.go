package tests

import (
	"testing"

	"github.com/Atharv3221/cavy"
)

// runCacheTests exercises the basic Cache contract: Get/Put/Update/Delete/
// Contains/GetAll/Len behavior. Fixed signature to actually accept key2/value2.
func runCacheTests[K comparable, V comparable](
	t *testing.T,
	cache cavy.Cache[K, V],
	key1 K, value1 V,
	key2 K, value2 V,
) {
	t.Helper()

	// Cache should start empty.
	if cache.Len() != 0 {
		t.Fatalf("expected empty cache, got %d items", cache.Len())
	}

	// Missing key.
	if _, ok := cache.Get(key1); ok {
		t.Fatal("expected missing key")
	}
	if cache.Contains(key1) {
		t.Fatal("expected Contains to return false for missing key")
	}

	// Insert first item.
	cache.Put(key1, value1)

	if cache.Len() != 1 {
		t.Fatalf("expected len 1, got %d", cache.Len())
	}
	if !cache.Contains(key1) {
		t.Fatal("expected Contains to return true after insert")
	}

	value, ok := cache.Get(key1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != value1 {
		t.Fatalf("expected %v, got %v", value1, value)
	}

	// Update existing key. Note: value2 is reused as the "updated" value
	// for key1, matching the original test's intent.
	cache.Put(key1, value2)

	if cache.Len() != 1 {
		t.Fatalf("expected len to remain 1 after update, got %d", cache.Len())
	}

	value, ok = cache.Get(key1)
	if !ok {
		t.Fatal("expected updated key to exist")
	}
	if value != value2 {
		t.Fatalf("expected updated value %v, got %v", value2, value)
	}

	// Insert another key.
	cache.Put(key2, value1)

	if cache.Len() != 2 {
		t.Fatalf("expected len 2, got %d", cache.Len())
	}

	// Delete.
	if !cache.Delete(key1) {
		t.Fatal("expected Delete to return true for existing key")
	}

	if cache.Len() != 1 {
		t.Fatalf("expected len 1 after delete, got %d", cache.Len())
	}

	if _, ok := cache.Get(key1); ok {
		t.Fatal("expected deleted key to be absent")
	}
	if cache.Contains(key1) {
		t.Fatal("expected Contains to return false after delete")
	}

	// Delete non-existent key should be a safe no-op.
	if cache.Delete(key1) {
		t.Fatal("expected Delete to return false for non-existent key")
	}

	// GetAll.
	all := cache.GetAll()

	if len(all) != 1 {
		t.Fatalf("expected GetAll to return 1 item, got %d", len(all))
	}

	if all[key2] != value1 {
		t.Fatalf("expected GetAll to contain %v", key2)
	}
}

// runCapacityTests verifies that a freshly created cache respects its
// configured capacity, evicts something on overflow, and that the most
// recently inserted item survives the eviction.
func runCapacityTests[K comparable, V comparable](
	t *testing.T,
	cache cavy.Cache[K, V],
	capacity int,
	keys []K, values []V,
) {
	t.Helper()

	if len(keys) < capacity+1 || len(values) < capacity+1 {
		t.Fatalf("need at least %d keys/values for capacity tests", capacity+1)
	}

	if cache.GetMaxCapacity() != capacity {
		t.Fatalf("expected max capacity %d, got %d", capacity, cache.GetMaxCapacity())
	}

	// Fill to capacity.
	for i := 0; i < capacity; i++ {
		cache.Put(keys[i], values[i])
	}

	if cache.Len() != capacity {
		t.Fatalf("expected len %d, got %d", capacity, cache.Len())
	}

	// Inserting beyond capacity should evict something and keep len bounded.
	cache.Put(keys[capacity], values[capacity])

	if cache.Len() != capacity {
		t.Fatalf("expected len to remain at capacity %d after eviction, got %d", capacity, cache.Len())
	}

	// The newest key must be present.
	if !cache.Contains(keys[capacity]) {
		t.Fatalf("expected most recently inserted key %v to be present", keys[capacity])
	}

	// Exactly one of the original keys should have been evicted.
	evictedCount := 0
	for i := 0; i < capacity; i++ {
		if !cache.Contains(keys[i]) {
			evictedCount++
		}
	}
	if evictedCount != 1 {
		t.Fatalf("expected exactly 1 eviction, got %d", evictedCount)
	}
}

func TestLRUCommon(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)
	runCacheTests(t, cache, 1, "one", 2, "two")
}

func TestLFUCommon(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LFU, 2)
	runCacheTests(t, cache, 1, "one", 2, "two")
}

func TestLRUCapacityEviction(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)
	runCapacityTests(t, cache, 2,
		[]int{1, 2, 3},
		[]string{"one", "two", "three"},
	)
}

func TestLFUCapacityEviction(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LFU, 2)
	runCapacityTests(t, cache, 2,
		[]int{1, 2, 3},
		[]string{"one", "two", "three"},
	)
}

// TestLRUEvictsLeastRecentlyUsed checks LRU-specific ordering semantics:
// touching a key via Get should protect it from eviction.
func TestLRUEvictsLeastRecentlyUsed(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)

	cache.Put(1, "one")
	cache.Put(2, "two")

	// Access key 1 so it becomes the most recently used.
	if _, ok := cache.Get(1); !ok {
		t.Fatal("expected key 1 to exist")
	}

	// Insert key 3, which should evict key 2 (least recently used).
	cache.Put(3, "three")

	if cache.Contains(2) {
		t.Fatal("expected key 2 to be evicted as least recently used")
	}
	if !cache.Contains(1) {
		t.Fatal("expected key 1 to still be present")
	}
	if !cache.Contains(3) {
		t.Fatal("expected key 3 to be present")
	}
}

// TestLRUPutUpdateRefreshesRecency checks that re-Put-ing an existing key
// also counts as "using" it for LRU purposes.
func TestLRUPutUpdateRefreshesRecency(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)

	cache.Put(1, "one")
	cache.Put(2, "two")

	// Updating key 1 via Put should also count as a use.
	cache.Put(1, "ONE")

	cache.Put(3, "three")

	if cache.Contains(2) {
		t.Fatal("expected key 2 to be evicted as least recently used")
	}
	if !cache.Contains(1) {
		t.Fatal("expected key 1 to still be present after refresh via Put")
	}
}

// TestLFUEvictsLeastFrequentlyUsed checks LFU-specific ordering semantics:
// a key accessed more often should be protected from eviction.
func TestLFUEvictsLeastFrequentlyUsed(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LFU, 2)

	cache.Put(1, "one")
	cache.Put(2, "two")

	// Access key 1 multiple times to raise its frequency above key 2's.
	cache.Get(1)
	cache.Get(1)

	// Insert key 3, which should evict key 2 (least frequently used).
	cache.Put(3, "three")

	if cache.Contains(2) {
		t.Fatal("expected key 2 to be evicted as least frequently used")
	}
	if !cache.Contains(1) {
		t.Fatal("expected key 1 to still be present")
	}
	if !cache.Contains(3) {
		t.Fatal("expected key 3 to be present")
	}
}

func TestSetCapacityGrow(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)

	cache.Put(1, "one")
	cache.Put(2, "two")

	if !cache.SetCapacity(5) {
		t.Fatal("expected SetCapacity to succeed when growing")
	}

	if cache.GetMaxCapacity() != 5 {
		t.Fatalf("expected max capacity 5, got %d", cache.GetMaxCapacity())
	}

	// Existing items should not be evicted by growing.
	if cache.Len() != 2 {
		t.Fatalf("expected len to remain 2 after growing capacity, got %d", cache.Len())
	}

	cache.Put(3, "three")
	cache.Put(4, "four")
	cache.Put(5, "five")

	if cache.Len() != 5 {
		t.Fatalf("expected len 5 after filling new capacity, got %d", cache.Len())
	}
}

func TestSetCapacityShrinkToExactFit(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 5)

	cache.Put(1, "one")
	cache.Put(2, "two")
	cache.Put(3, "three")

	if !cache.SetCapacity(3) {
		t.Fatal("expected SetCapacity to succeed when new capacity exactly fits current item count")
	}

	if cache.GetMaxCapacity() != 3 {
		t.Fatalf("expected max capacity 3, got %d", cache.GetMaxCapacity())
	}
	if cache.Len() != 3 {
		t.Fatalf("expected len to remain 3, got %d", cache.Len())
	}
}
func TestSetCapacityInvalid(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)

	if cache.SetCapacity(0) {
		t.Fatal("expected SetCapacity(0) to fail")
	}
	if cache.SetCapacity(-1) {
		t.Fatal("expected SetCapacity(-1) to fail")
	}

	if cache.GetMaxCapacity() != 2 {
		t.Fatalf("expected max capacity to remain 2 after invalid SetCapacity calls, got %d", cache.GetMaxCapacity())
	}
}

func TestNewCacheInvalidPolicyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid policy")
		}
	}()

	cavy.NewCache[int, string](cavy.Policy(999), 2)
}

func TestDefaultCapacity(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 0)

	cache.Put(1, "one")

	// Depending on implementation, a zero-capacity cache may either
	// reject all inserts or immediately evict whatever is added.
	// Either way it should never hold an item.
	if cache.Len() != 1 {
		t.Fatalf("expected zero-capacity cache to set to default items, got %d", cache.Len())
	}
}

func TestGetAllReturnsIndependentCopy(t *testing.T) {
	cache := cavy.NewCache[int, string](cavy.LRU, 2)

	cache.Put(1, "one")

	all := cache.GetAll()
	all[1] = "mutated"
	all[2] = "added"

	value, ok := cache.Get(1)
	if !ok || value != "one" {
		t.Fatal("expected GetAll's returned map mutation to not affect the cache")
	}
	if cache.Len() != 1 {
		t.Fatalf("expected len to remain 1, got %d", cache.Len())
	}
}
