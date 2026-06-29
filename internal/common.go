package internal

import "github.com/Atharv3221/cavy/utils"

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type general_structure[K comparable, V any] struct {
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

func (g *general_structure[K, V]) GetMaxCapacity() int {
	return g.capacity
}

func (g *general_structure[K, V]) Contains(key K) bool {
	_, conatins := g.items[key]
	return conatins
}

func (g *general_structure[K, V]) GetAll() map[K]V {
	result := make(map[K]V, len(g.items))
	for key, n := range g.items {
		result[key] = n.value
	}
	return result
}

func (g *general_structure[K, V]) SetCapacity(capacity int) bool {
	var result bool
	g.capacity, result = utils.SetMaxCapacity(g.capacity, capacity, g.Len())
	return result
}

func (g *general_structure[K, V]) Len() int {
	return len(g.items)
}
