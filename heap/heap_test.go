package heap

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"testing"
)

// Int is a comparable wrapper around int used in tests.
type Int int

func (a Int) Compare(b Int) int { return cmp.Compare(int(a), int(b)) }

// Str is a comparable wrapper around string used in tests.
type Str string

func (a Str) Compare(b Str) int { return cmp.Compare(string(a), string(b)) }

// Task is a struct payload ordered by priority then id (for deterministic ties).
type Task struct {
	id       int
	priority int
}

func (a Task) Compare(b Task) int {
	if c := cmp.Compare(a.priority, b.priority); c != 0 {
		return c
	}
	return cmp.Compare(a.id, b.id)
}

func drain[T Ordered[T]](h *Heap[T]) []T {
	out := make([]T, 0, h.Len())
	for v := range h.Drain() {
		out = append(out, v)
	}
	return out
}

func TestPushPopMin(t *testing.T) {
	h := New[Int]()
	for _, v := range []Int{5, 3, 8, 1, 9, 2, 7, 4, 6} {
		h.Push(v)
	}
	got := drain(h)
	want := []Int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !slices.Equal(got, want) {
		t.Fatalf("min Drain = %v, want %v", got, want)
	}
}

func TestPushPopMax(t *testing.T) {
	h := NewMax[Int]()
	for _, v := range []Int{5, 3, 8, 1, 9, 2, 7, 4, 6} {
		h.Push(v)
	}
	got := drain(h)
	want := []Int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	if !slices.Equal(got, want) {
		t.Fatalf("max Drain = %v, want %v", got, want)
	}
}

func TestStrings(t *testing.T) {
	h := New[Str]()
	for _, s := range []Str{"banana", "apple", "cherry", "date"} {
		h.Push(s)
	}
	got := drain(h)
	want := []Str{"apple", "banana", "cherry", "date"}
	if !slices.Equal(got, want) {
		t.Fatalf("string Drain = %v, want %v", got, want)
	}
}

func TestStruct(t *testing.T) {
	h := New[Task]()
	h.Push(Task{id: 1, priority: 5})
	h.Push(Task{id: 2, priority: 1})
	h.Push(Task{id: 3, priority: 3})
	h.Push(Task{id: 4, priority: 1})
	got := drain(h)
	want := []Task{{2, 1}, {4, 1}, {3, 3}, {1, 5}}
	if !slices.Equal(got, want) {
		t.Fatalf("task Drain = %v, want %v", got, want)
	}
}

func TestPeekAndLen(t *testing.T) {
	h := New[Int]()
	if _, ok := h.Peek(); ok {
		t.Fatal("Peek on empty heap should return ok=false")
	}
	if _, ok := h.Pop(); ok {
		t.Fatal("Pop on empty heap should return ok=false")
	}
	if h.Len() != 0 {
		t.Fatalf("Len = %d, want 0", h.Len())
	}
	h.Push(7)
	h.Push(3)
	h.Push(5)
	if h.Len() != 3 {
		t.Fatalf("Len = %d, want 3", h.Len())
	}
	if v, ok := h.Peek(); !ok || v != 3 {
		t.Fatalf("Peek = (%v, %v), want (3, true)", v, ok)
	}
	if h.Len() != 3 {
		t.Fatalf("Peek mutated Len: got %d, want 3", h.Len())
	}
}

func TestFromHeapifyMatchesPushes(t *testing.T) {
	values := []Int{42, 7, 13, 99, 1, 23, 88, 4, 56, 31, 67, 2}

	pushed := New[Int]()
	for _, v := range values {
		pushed.Push(v)
	}
	heapified := From(values)

	gotPushed := drain(pushed)
	gotHeapified := drain(heapified)
	if !slices.Equal(gotPushed, gotHeapified) {
		t.Fatalf("From vs Push order differ:\n  From:  %v\n  Push:  %v", gotHeapified, gotPushed)
	}

	wantSorted := slices.Clone(values)
	slices.Sort(wantSorted)
	if !slices.Equal(gotHeapified, wantSorted) {
		t.Fatalf("From Drain = %v, want %v", gotHeapified, wantSorted)
	}
}

func TestMaxFromHeapify(t *testing.T) {
	values := []Int{42, 7, 13, 99, 1, 23, 88, 4, 56, 31, 67, 2}
	h := MaxFrom(values)
	got := drain(h)
	want := slices.Clone(values)
	slices.SortFunc(want, func(a, b Int) int { return cmp.Compare(b, a) })
	if !slices.Equal(got, want) {
		t.Fatalf("MaxFrom Drain = %v, want %v", got, want)
	}
}

func TestUpdateRaisesAndLowers(t *testing.T) {
	h := New[Int]()
	a := h.Push(10)
	b := h.Push(20)
	c := h.Push(30)

	// Lower b's priority below the current root.
	h.Update(b, 5)
	if v, _ := h.Peek(); v != 5 {
		t.Fatalf("after lowering b, Peek = %v, want 5", v)
	}

	// Raise a above c.
	h.Update(a, 100)

	got := drain(h)
	want := []Int{5, 30, 100}
	if !slices.Equal(got, want) {
		t.Fatalf("Drain = %v, want %v", got, want)
	}
	_ = c
}

func TestRemoveRootMiddleLeaf(t *testing.T) {
	for _, removeIdx := range []int{0, 3, 6} { // root, middle, leaf
		h := New[Int]()
		items := make([]*Item[Int], 0, 7)
		for _, v := range []Int{8, 4, 6, 2, 5, 1, 3} {
			items = append(items, h.Push(v))
		}
		removed := h.Remove(items[removeIdx])
		got := drain(h)
		if !slices.IsSorted(got) {
			t.Fatalf("after Remove(%d=%v), Drain not sorted: %v", removeIdx, removed, got)
		}
		if len(got) != 6 {
			t.Fatalf("after Remove, len = %d, want 6", len(got))
		}
	}
}

func TestStaleHandlePanicsAfterPop(t *testing.T) {
	h := New[Int]()
	it := h.Push(1)
	h.Pop()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Update on popped handle should panic")
		}
	}()
	h.Update(it, 2)
}

func TestStaleHandlePanicsAfterRemove(t *testing.T) {
	h := New[Int]()
	a := h.Push(1)
	h.Push(2)
	h.Remove(a)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Remove on detached handle should panic")
		}
	}()
	h.Remove(a)
}

func TestNilHandlePanics(t *testing.T) {
	h := New[Int]()
	h.Push(1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Update with nil handle should panic")
		}
	}()
	h.Update(nil, 2)
}

func TestAllSnapshotIsNonDestructive(t *testing.T) {
	h := New[Int]()
	for _, v := range []Int{3, 1, 4, 1, 5, 9, 2, 6} {
		h.Push(v)
	}
	startLen := h.Len()
	count := 0
	seen := []Int{}
	for v := range h.All() {
		seen = append(seen, v)
		count++
	}
	if count != startLen {
		t.Fatalf("All yielded %d, want %d", count, startLen)
	}
	if h.Len() != startLen {
		t.Fatalf("All mutated heap: len %d -> %d", startLen, h.Len())
	}

	gotSorted := slices.Clone(seen)
	slices.Sort(gotSorted)
	wantSorted := []Int{1, 1, 2, 3, 4, 5, 6, 9}
	if !slices.Equal(gotSorted, wantSorted) {
		t.Fatalf("All multiset = %v, want %v", gotSorted, wantSorted)
	}
}

func TestDrainEarlyStopLeavesRest(t *testing.T) {
	h := New[Int]()
	for _, v := range []Int{5, 3, 8, 1, 9, 2, 7} {
		h.Push(v)
	}
	taken := []Int{}
	for v := range h.Drain() {
		taken = append(taken, v)
		if len(taken) == 3 {
			break
		}
	}
	if !slices.Equal(taken, []Int{1, 2, 3}) {
		t.Fatalf("first 3 = %v, want [1 2 3]", taken)
	}
	if h.Len() != 4 {
		t.Fatalf("after early stop, Len = %d, want 4", h.Len())
	}
	rest := drain(h)
	if !slices.Equal(rest, []Int{5, 7, 8, 9}) {
		t.Fatalf("rest = %v, want [5 7 8 9]", rest)
	}
}

// Property test: random Push/Update/Remove sequence should still drain sorted,
// and the multiset should match what was pushed minus what was removed.
func TestProperty_RandomOps(t *testing.T) {
	const N = 10_000
	rng := rand.New(rand.NewPCG(0xC0FFEE, 0xBEEF))

	h := New[Int]()
	live := map[*Item[Int]]Int{}
	multiset := map[Int]int{}

	for range N {
		op := rng.IntN(10)
		switch {
		case op < 6 || len(live) == 0: // push
			v := Int(rng.IntN(1_000_000))
			it := h.Push(v)
			live[it] = v
			multiset[v]++
		case op < 8: // update
			it := pickKey(rng, live)
			oldV := live[it]
			newV := Int(rng.IntN(1_000_000))
			h.Update(it, newV)
			live[it] = newV
			multiset[oldV]--
			if multiset[oldV] == 0 {
				delete(multiset, oldV)
			}
			multiset[newV]++
		default: // remove
			it := pickKey(rng, live)
			v := live[it]
			got := h.Remove(it)
			if got != v {
				t.Fatalf("Remove returned %v, expected %v", got, v)
			}
			delete(live, it)
			multiset[v]--
			if multiset[v] == 0 {
				delete(multiset, v)
			}
		}
	}

	if h.Len() != sumMultiset(multiset) {
		t.Fatalf("Len %d, multiset count %d", h.Len(), sumMultiset(multiset))
	}

	out := drain(h)
	if !slices.IsSorted(out) {
		t.Fatal("drain output is not sorted")
	}
	gotMS := map[Int]int{}
	for _, v := range out {
		gotMS[v]++
	}
	if len(gotMS) != len(multiset) {
		t.Fatalf("multiset size mismatch: got %d, want %d", len(gotMS), len(multiset))
	}
	for k, v := range multiset {
		if gotMS[k] != v {
			t.Fatalf("multiset[%v] = %d, want %d", k, gotMS[k], v)
		}
	}
}

func pickKey[K comparable, V any](rng *rand.Rand, m map[K]V) K {
	target := rng.IntN(len(m))
	i := 0
	for k := range m {
		if i == target {
			return k
		}
		i++
	}
	panic("unreachable")
}

func sumMultiset(m map[Int]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
