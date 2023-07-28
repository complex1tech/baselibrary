package refmap

import (
	"sort"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/compare"
	"github.com/basecomplextech/baselibrary/ref"
)

var _ node[any, ref.Ref] = (*leafNode[any, ref.Ref])(nil)

type leafNode[K any, V ref.Ref] struct {
	items []leafItem[K, V]

	mut  bool
	refs int64
}

type leafItem[K any, V ref.Ref] struct {
	key   K
	value V
}

// newLeafNode returns a new mutable node.
func newLeafNode[K any, V ref.Ref](items ...Item[K, V]) *leafNode[K, V] {
	// Make node
	n := &leafNode[K, V]{
		items: make([]leafItem[K, V], 0, len(items)),

		mut:  true,
		refs: 1,
	}

	// Copy items
	for _, item := range items {
		n.items = append(n.items, leafItem[K, V]{
			key:   item.Key,
			value: item.Value,
		})
	}

	// Retain items
	for _, item := range n.items {
		item.value.Retain()
	}
	return n
}

// cloneLeafNode returns a mutable node clone.
func cloneLeafNode[K any, V ref.Ref](n *leafNode[K, V]) *leafNode[K, V] {
	// Copy node
	n1 := &leafNode[K, V]{
		items: make([]leafItem[K, V], len(n.items)),

		mut:  true,
		refs: 1,
	}
	copy(n1.items, n.items)

	// Retain items
	for _, item := range n1.items {
		item.value.Retain()
	}
	return n1
}

// nextLeafNode returns a new mutable node on a split, moves items to it.
func nextLeafNode[K any, V ref.Ref](items []leafItem[K, V]) *leafNode[K, V] {
	n := &leafNode[K, V]{
		refs:  1,
		mut:   true,
		items: make([]leafItem[K, V], len(items), cap(items)),
	}
	copy(n.items, items)
	return n
}

// retain/release

func (n *leafNode[K, V]) retain() {
	v := atomic.AddInt64(&n.refs, 1)
	if v == 1 {
		panic("retained already released node")
	}
}

func (n *leafNode[K, V]) release() {
	v := atomic.AddInt64(&n.refs, -1)
	switch {
	case v < 0:
		panic("released already released node")
	case v > 0:
		return
	}

	// Release and clear items
	for i, item := range n.items {
		item.value.Release()
		n.items[i] = leafItem[K, V]{}
	}
	n.items = n.items[:0]
}

func (n *leafNode[K, V]) refcount() int64 {
	return n.refs
}

// attrs

func (n *leafNode[K, V]) length() int {
	return len(n.items)
}

func (n *leafNode[K, V]) minKey() K {
	return n.items[0].key
}

func (n *leafNode[K, V]) maxKey() K {
	return n.items[len(n.items)-1].key
}

// mutate

func (n *leafNode[K, V]) clone() node[K, V] {
	return cloneLeafNode(n)
}

func (n *leafNode[K, V]) freeze() {
	n.mut = false
}

func (n *leafNode[K, V]) mutable() bool {
	return n.mut
}

// methods

// indexOf returns an index of an item with key >= key.
func (n *leafNode[K, V]) indexOf(key K, compare compare.Func[K]) int {
	return sort.Search(len(n.items), func(i int) bool {
		key0 := n.items[i].key
		cmp := compare(key0, key)
		return cmp >= 0
	})
}

func (n *leafNode[K, V]) get(key K, compare compare.Func[K]) (v V, ok bool) {
	index := n.indexOf(key, compare)

	// Return if not found
	if index >= len(n.items) {
		return
	}
	if compare(n.items[index].key, key) != 0 {
		return
	}

	item := n.items[index]
	return item.value, true
}

func (n *leafNode[K, V]) put(key K, value V, compare compare.Func[K]) bool {
	if !n.mut {
		panic("operation on immutable node")
	}
	if len(n.items) == maxItems {
		panic("cannot insert into full node")
	}

	// Find item by key
	index := n.indexOf(key, compare)

	// Replace existing if found
	if index < len(n.items) {
		item := &n.items[index]

		// Swap item
		if compare(item.key, key) == 0 {
			item.value = ref.Swap(item.value, value)
			return false
		}
	}

	// Grow capacity
	if cap(n.items) < len(n.items)+1 {
		new := 2*len(n.items) + 1
		items := make([]leafItem[K, V], len(n.items), new)

		copy(items, n.items)
		n.items = items
	}

	// Shift greater items right
	n.items = n.items[:len(n.items)+1]
	copy(n.items[index+1:], n.items[index:])

	// Insert new item at index
	n.items[index] = leafItem[K, V]{
		key:   key,
		value: value,
	}

	// Retain value
	value.Retain()
	return true
}

func (n *leafNode[K, V]) delete(key K, compare compare.Func[K]) bool {
	if !n.mut {
		panic("operation on immutable node")
	}

	// Find item by key
	index := n.indexOf(key, compare)

	// Return if not found
	if index >= len(n.items) {
		return false
	}
	if compare(n.items[index].key, key) != 0 {
		return false
	}

	// Release item
	item := n.items[index]
	item.value.Release()

	// Shift greater items left
	copy(n.items[index:], n.items[index+1:])
	n.items[len(n.items)-1] = leafItem[K, V]{}

	// Truncate items
	n.items = n.items[:len(n.items)-1]
	return true
}

func (n *leafNode[K, V]) contains(key K, compare compare.Func[K]) bool {
	index := n.indexOf(key, compare)
	if index >= len(n.items) {
		return false
	}

	cmp := compare(n.items[index].key, key)
	return cmp == 0
}

func (n *leafNode[K, V]) split() (node[K, V], bool) {
	if !n.mut {
		panic("operation on immutable node")
	}

	if len(n.items) < maxItems {
		return nil, false
	}

	// Calc middle index
	middle := len(n.items) / 2

	// Move items to next node
	next := nextLeafNode(n.items[middle:len(n.items)])

	// Clear and truncate items
	for i := middle; i < len(n.items); i++ {
		n.items[i] = leafItem[K, V]{}
	}
	n.items = n.items[:middle]
	return next, true
}
