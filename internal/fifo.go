package internal

type fifo[K comparable, V any] struct {
	general_structure[K, V]
}

func NewFIFO[K comparable, V any](capacity int) *fifo[K, V] {
	if capacity <= 0 {
		capacity = DefaultMaxCapacity
	}
	return &fifo[K, V]{
		general_structure: general_structure[K, V]{
			capacity: capacity,
			items:    make(map[K]*node[K, V]),
		},
	}
}

func (f *fifo[K, V]) Get(key K) (V, bool) {
	n, ok := f.items[key]
	var value V
	if ok {
		value = n.value
	}
	return value, ok
}

func (f *fifo[K, V]) Put(key K, value V) {
	if f.Contains(key) {
		// avoid duplicate
		n, _ := f.items[key]
		f.deleteNode(n)
	}
	f.addFront(&node[K, V]{key: key, value: value})
	if f.Len() > f.capacity {
		f.deleteNode(f.tail)
	}
}

func (f *fifo[K, V]) Delete(key K) bool {
	result := false
	if f.Contains(key) {
		f.deleteNode(f.items[key])
		result = true
	}
	return result
}

func (f *fifo[K, V]) addFront(n *node[K, V]) {
	f.items[n.key] = n
	if f.head == nil {
		f.head = n
		f.tail = n
	} else {
		f.head.next = n
		n.prev = f.head
		f.head = n
	}
}

func (f *fifo[K, V]) deleteNode(n *node[K, V]) {
	if n == nil {
		return
	}
	if n.next != nil {
		n.next.prev = n.prev
	}

	if n.prev != nil {
		n.prev.next = n.next
	}

	if n == f.head {
		f.head = n.prev
	}

	if n == f.tail {
		f.tail = n.next
	}
	delete(f.items, n.key)
}
