package internal

type lru[K comparable, V any] struct {
	general_structure[K, V]
}

func NewLRU[K comparable, V any](capacity int) *lru[K, V] {
	if capacity <= 0 {
		capacity = DefaultMaxCapacity
	}
	return &lru[K, V]{
		general_structure: general_structure[K, V]{
			capacity: capacity,
			items:    make(map[K]*node[K, V]),
		},
	}
}

func (l *lru[K, V]) Put(key K, value V) {
	if n, exists := l.items[key]; exists {
		n.value = value
		l.moveToFront(n)
		return
	}

	n := &node[K, V]{key: key, value: value}
	l.items[key] = n
	l.insertAtFront(n)

	if len(l.items) > l.capacity {
		l.evictTail()
	}
}

func (l *lru[K, V]) Get(key K) (V, bool) {
	n, ok := l.items[key]
	if !ok {
		var value V
		return value, false
	}
	l.moveToFront(n)
	return n.value, true
}

func (l *lru[K, V]) Delete(key K) bool {
	if !l.Contains(key) {
		return false
	}

	n := l.items[key]

	if n != l.head {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}

	if n != l.tail {
		n.next.prev = n.prev

	} else {
		l.tail = n.prev
	}

	delete(l.items, key)
	return true
}

// Removes last accessed element
func (l *lru[K, V]) evictTail() {
	if l.tail == nil {
		return
	}
	delete(l.items, l.tail.key)

	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		return
	}
	l.tail = l.tail.prev
	l.tail.next = nil
}

// Adds new node at Head
func (l *lru[K, V]) insertAtFront(n *node[K, V]) {
	if l.head == nil {
		l.head = n
		l.tail = n
		return
	}
	n.next = l.head
	l.head.prev = n
	l.head = n

}

// Moves passed node to Head
func (l *lru[K, V]) moveToFront(n *node[K, V]) {
	if n == l.head {
		return
	}

	n.prev.next = n.next

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}

	n.prev = nil
	n.next = l.head
	l.head.prev = n
	l.head = n
}
