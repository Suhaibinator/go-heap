// Package heap provides a generic binary heap (priority queue) over any type
// that knows how to compare itself.
//
// To use the heap, your type must implement Ordered by exposing a Compare
// method whose signature matches time.Time.Compare, netip.Addr.Compare, and
// the contract of cmp.Compare:
//
//	Compare(other) < 0  receiver is "less than" other
//	Compare(other) == 0 equal
//	Compare(other) > 0  receiver is "greater than" other
//
// New returns a min-heap; NewMax returns a max-heap. Both share the same
// implementation and differ only in the comparator they install.
//
// The heap is not safe for concurrent use. Wrap externally with a sync.Mutex
// if you need to share it across goroutines.
package heap

import "iter"

// Ordered is satisfied by any type that can compare itself to another value
// of the same type.
type Ordered[T any] interface {
	Compare(other T) int
}

// Item is a handle to an element stored in a Heap. The handle stays valid
// across Push/Pop/Update/Remove operations and lets callers update or remove
// a specific element without searching the heap. After Pop or Remove, the
// handle becomes stale and further use panics.
type Item[T Ordered[T]] struct {
	value T
	index int // -1 once detached from a heap
}

// Value returns the element currently held by the handle.
func (it *Item[T]) Value() T { return it.value }

// Heap is a generic binary heap. The zero value is not usable; construct one
// with New, NewMax, From, or MaxFrom.
type Heap[T Ordered[T]] struct {
	data []*Item[T]
	less func(a, b T) bool
}

// New returns an empty min-heap.
func New[T Ordered[T]]() *Heap[T] {
	return &Heap[T]{less: minLess[T]}
}

// NewMax returns an empty max-heap.
func NewMax[T Ordered[T]]() *Heap[T] {
	return &Heap[T]{less: maxLess[T]}
}

// From returns a min-heap containing the given items, built in O(n) via
// Floyd's heapify. The input slice is not retained.
func From[T Ordered[T]](items []T) *Heap[T] {
	return fromSlice(items, minLess[T])
}

// MaxFrom returns a max-heap containing the given items, built in O(n) via
// Floyd's heapify. The input slice is not retained.
func MaxFrom[T Ordered[T]](items []T) *Heap[T] {
	return fromSlice(items, maxLess[T])
}

func minLess[T Ordered[T]](a, b T) bool { return a.Compare(b) < 0 }
func maxLess[T Ordered[T]](a, b T) bool { return a.Compare(b) > 0 }

func fromSlice[T Ordered[T]](items []T, less func(a, b T) bool) *Heap[T] {
	h := &Heap[T]{
		data: make([]*Item[T], len(items)),
		less: less,
	}
	for i, v := range items {
		h.data[i] = &Item[T]{value: v, index: i}
	}
	for i := len(h.data)/2 - 1; i >= 0; i-- {
		h.siftDown(i)
	}
	return h
}

// Len returns the number of elements in the heap.
func (h *Heap[T]) Len() int { return len(h.data) }

// Peek returns the root element without removing it. The second return value
// is false if the heap is empty.
func (h *Heap[T]) Peek() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.data[0].value, true
}

// Push inserts v into the heap and returns a handle to it. Callers that don't
// need Update or Remove can ignore the return value.
func (h *Heap[T]) Push(v T) *Item[T] {
	it := &Item[T]{value: v, index: len(h.data)}
	h.data = append(h.data, it)
	h.siftUp(it.index)
	return it
}

// Pop removes and returns the root element. The second return value is false
// if the heap is empty.
func (h *Heap[T]) Pop() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.removeAt(0), true
}

// Update changes the value held by it and re-heapifies. It panics if it has
// already been detached from the heap (e.g., via Pop or Remove).
func (h *Heap[T]) Update(it *Item[T], v T) {
	h.checkAttached(it, "Update")
	it.value = v
	if !h.siftDown(it.index) {
		h.siftUp(it.index)
	}
}

// Remove detaches it from the heap and returns its value. It panics if it has
// already been detached.
func (h *Heap[T]) Remove(it *Item[T]) T {
	h.checkAttached(it, "Remove")
	return h.removeAt(it.index)
}

// Drain returns an iterator that pops elements in sorted order, consuming the
// heap. Stopping iteration early leaves the remaining elements in the heap.
func (h *Heap[T]) Drain() iter.Seq[T] {
	return func(yield func(T) bool) {
		for len(h.data) > 0 {
			v := h.removeAt(0)
			if !yield(v) {
				return
			}
		}
	}
}

// All returns an iterator over a snapshot of the heap's elements in
// heap-array order (not sorted). The heap is not modified, and concurrent
// mutation during iteration is undefined.
func (h *Heap[T]) All() iter.Seq[T] {
	snap := make([]T, len(h.data))
	for i, it := range h.data {
		snap[i] = it.value
	}
	return func(yield func(T) bool) {
		for _, v := range snap {
			if !yield(v) {
				return
			}
		}
	}
}

func (h *Heap[T]) checkAttached(it *Item[T], op string) {
	if it == nil {
		panic("heap: " + op + " called with nil item")
	}
	if it.index < 0 || it.index >= len(h.data) || h.data[it.index] != it {
		panic("heap: " + op + " called on detached item")
	}
}

func (h *Heap[T]) removeAt(i int) T {
	last := len(h.data) - 1
	out := h.data[i]
	if i != last {
		h.swap(i, last)
	}
	h.data[last] = nil
	h.data = h.data[:last]
	out.index = -1
	if i < last {
		if !h.siftDown(i) {
			h.siftUp(i)
		}
	}
	return out.value
}

func (h *Heap[T]) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
	h.data[i].index = i
	h.data[j].index = j
}

func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(h.data[i].value, h.data[parent].value) {
			return
		}
		h.swap(i, parent)
		i = parent
	}
}

// siftDown moves data[i] down until the heap property holds. Returns true if
// any swap happened, so Update can fall back to siftUp when no descent occurred.
func (h *Heap[T]) siftDown(i int) bool {
	n := len(h.data)
	start := i
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		smallest := left
		if right := left + 1; right < n && h.less(h.data[right].value, h.data[left].value) {
			smallest = right
		}
		if !h.less(h.data[smallest].value, h.data[i].value) {
			break
		}
		h.swap(i, smallest)
		i = smallest
	}
	return i > start
}
