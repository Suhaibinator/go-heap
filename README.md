# go-heap

[![CI](https://github.com/Suhaibinator/go-heap/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Suhaibinator/go-heap/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Suhaibinator/go-heap.svg)](https://pkg.go.dev/github.com/Suhaibinator/go-heap)
[![Go Report Card](https://goreportcard.com/badge/github.com/Suhaibinator/go-heap)](https://goreportcard.com/report/github.com/Suhaibinator/go-heap)
[![License: MIT](https://img.shields.io/github/license/Suhaibinator/go-heap)](LICENSE)

A generic binary heap (priority queue) for Go. Any type that exposes a
`Compare(other T) int` method can be stored in the heap — the same signature
used by `time.Time.Compare`, `netip.Addr.Compare`, `cmp.Compare`, and
`slices.SortFunc`, so types you already use for sorting work without changes.

```go
import heap "github.com/Suhaibinator/go-heap"
```

Requires Go 1.26.2+ (uses `iter.Seq`).

## The interface

```go
type Ordered[T any] interface {
    Compare(other T) int
}
```

Return value follows the standard convention: negative if the receiver is less
than `other`, zero if equal, positive if greater.

## Usage

### Primitive values via a wrapper type

```go
type Priority int
func (a Priority) Compare(b Priority) int { return cmp.Compare(int(a), int(b)) }

h := heap.New[Priority]()
h.Push(5); h.Push(1); h.Push(3)
v, _ := h.Pop() // 1
```

### Structs

```go
type Job struct {
    id, priority int
    name         string
}
func (a Job) Compare(b Job) int {
    if c := cmp.Compare(a.priority, b.priority); c != 0 { return c }
    return cmp.Compare(a.id, b.id)
}

h := heap.New[Job]()
h.Push(Job{id: 1, name: "deploy", priority: 5})
h.Push(Job{id: 2, name: "lint",   priority: 1})
```

### Max-heap

```go
h := heap.NewMax[Priority]()
```

`New` and `NewMax` share the same implementation; only the comparator differs.

### Bulk construction (O(n))

```go
h := heap.From([]Priority{5, 3, 8, 1, 9, 2})  // min
h := heap.MaxFrom([]Priority{5, 3, 8, 1, 9, 2}) // max
```

### Update and Remove via handles

`Push` returns an `*Item[T]` handle. Keep it if you'll need to change or
remove that element later (e.g., Dijkstra's algorithm, event timers,
cancellable jobs); otherwise discard it.

```go
job := h.Push(Job{id: 2, name: "lint", priority: 9})
// later, lint became urgent:
h.Update(job, Job{id: 2, name: "lint", priority: 0})
// or cancel it:
h.Remove(job)
```

After `Pop` or `Remove`, the handle is detached; subsequent `Update`/`Remove`
panics.

### Iteration

```go
for v := range h.Drain() { ... }  // pop in sorted order, consumes the heap
for v := range h.All()   { ... }  // non-destructive snapshot, unsorted
```

## API

| Method                     | Description                                       |
|----------------------------|---------------------------------------------------|
| `New[T]()`                 | Empty min-heap                                    |
| `NewMax[T]()`              | Empty max-heap                                    |
| `From(s)` / `MaxFrom(s)`   | O(n) heapify from a slice                         |
| `Len()`                    | Number of elements                                |
| `Peek()`                   | Root value, `ok=false` if empty                   |
| `Push(v)`                  | Insert; returns a handle                          |
| `Pop()`                    | Remove and return root, `ok=false` if empty       |
| `Update(handle, v)`        | Change a held value and re-heapify                |
| `Remove(handle)`           | Detach an element by handle and return its value  |
| `Drain()` / `All()`        | `iter.Seq[T]` iterators                           |

## Notes

- Not safe for concurrent use. Wrap externally with `sync.Mutex` if needed —
  matches the convention of `container/heap`, `slices`, and Go maps.
- Stdlib only; no third-party dependencies.

## License

MIT — see [LICENSE](LICENSE).
