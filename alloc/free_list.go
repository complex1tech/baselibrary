package alloc

import (
	"sync/atomic"
	"unsafe"
)

const (
	freeListAllocAttempts   = 3
	freeListDeallocAttempts = 3
)

// FreeList keeps a linked list of free objects in the arena.
type FreeList[T any] struct {
	arena *Arena
	size  uintptr
	free  uintptr // last free item
}

type freeListItem struct {
	next uintptr // next free item
}

func newFreeList[T any](arena *Arena) *FreeList[T] {
	var zero T
	size := unsafe.Sizeof(zero)

	itemSize := unsafe.Sizeof(freeListItem{})
	if size < itemSize {
		size = itemSize
	}

	return &FreeList[T]{
		arena: arena,
		size:  size,
	}
}

// Alloc allocates a zero object from the free list or from an arena.
func (l *FreeList[T]) Alloc() *T {
	var zero T

	for i := 0; i < freeListAllocAttempts; i++ {
		// load current item
		free := atomic.LoadUintptr(&l.free)
		if free == 0 {
			break
		}
		uptr := *(*unsafe.Pointer)(unsafe.Pointer(&free))

		// swap it with previous
		item := (*freeListItem)(uptr)
		if !atomic.CompareAndSwapUintptr(&l.free, free, item.next) {
			continue
		}

		// reset and return object
		result := (*T)(uptr)
		*result = zero
		return result
	}

	// allocate new object
	return Alloc[T](l.arena)
}

// Dealloc tries to put an object back into the free list.
func (l *FreeList[T]) Dealloc(ptr *T) {
	// reset object
	var zero T
	*ptr = zero

	// cast it into item
	item := (*freeListItem)(unsafe.Pointer(ptr))

	for i := 0; i < freeListDeallocAttempts; i++ {
		// load current item
		free := atomic.LoadUintptr(&l.free)
		item.next = free

		// swap it with next
		next := (uintptr)(unsafe.Pointer(ptr))
		if atomic.CompareAndSwapUintptr(&l.free, free, next) {
			return
		}
	}
}
