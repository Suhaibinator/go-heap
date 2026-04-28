package heap_test

import (
	"cmp"
	"fmt"

	heap "github.com/Suhaibinator/go-heap/heap"
)

// Priority is a primitive wrapper that satisfies heap.Ordered[Priority].
type Priority int

func (a Priority) Compare(b Priority) int { return cmp.Compare(int(a), int(b)) }

func ExampleNew() {
	h := heap.New[Priority]()
	for _, v := range []Priority{5, 3, 8, 1, 9, 2} {
		h.Push(v)
	}
	for v := range h.Drain() {
		fmt.Print(v, " ")
	}
	// Output: 1 2 3 5 8 9
}

func ExampleNewMax() {
	h := heap.NewMax[Priority]()
	for _, v := range []Priority{5, 3, 8, 1, 9, 2} {
		h.Push(v)
	}
	for v := range h.Drain() {
		fmt.Print(v, " ")
	}
	// Output: 9 8 5 3 2 1
}

// Job is a struct ordered by priority, with id breaking ties for determinism.
type Job struct {
	id       int
	name     string
	priority int
}

func (a Job) Compare(b Job) int {
	if c := cmp.Compare(a.priority, b.priority); c != 0 {
		return c
	}
	return cmp.Compare(a.id, b.id)
}

func ExampleHeap_struct() {
	h := heap.New[Job]()
	h.Push(Job{id: 1, name: "deploy", priority: 5})
	h.Push(Job{id: 2, name: "lint", priority: 1})
	h.Push(Job{id: 3, name: "test", priority: 3})

	for j := range h.Drain() {
		fmt.Printf("p=%d %s\n", j.priority, j.name)
	}
	// Output:
	// p=1 lint
	// p=3 test
	// p=5 deploy
}

// Example using Update to lower a job's priority after it has been queued —
// the pattern Dijkstra's algorithm uses on its priority queue.
func ExampleHeap_Update() {
	h := heap.New[Job]()
	h.Push(Job{id: 1, name: "deploy", priority: 5})
	lint := h.Push(Job{id: 2, name: "lint", priority: 9})
	h.Push(Job{id: 3, name: "test", priority: 3})

	// "lint" turned out to be urgent — re-prioritize it ahead of everything.
	h.Update(lint, Job{id: 2, name: "lint", priority: 0})

	for j := range h.Drain() {
		fmt.Printf("p=%d %s\n", j.priority, j.name)
	}
	// Output:
	// p=0 lint
	// p=3 test
	// p=5 deploy
}

func ExampleFrom() {
	// Bulk construction from an existing slice is O(n).
	h := heap.From([]Priority{5, 3, 8, 1, 9, 2})
	v, _ := h.Peek()
	fmt.Println(v)
	// Output: 1
}
