package fsm

import (
	"container/heap"
)

// Min-heap of container groups prioritized by the version
type MinHeap struct {
	data      []ManagedSet
	EqualFunc func(a, b interface{}) bool
	LessFunc  func(a, b interface{}) bool
	NewFunc   func(interface{}) ManagedSet
}

func (h *MinHeap) Len() int { return len(h.data) }
func (h *MinHeap) Less(i, j int) bool {
	return h.LessFunc(h.data[i].GetKey(), h.data[j].GetKey())
}
func (h *MinHeap) Swap(i, j int) { h.data[i], h.data[j] = h.data[j], h.data[i] }
func (h *MinHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	h.data = append(h.data, x.(ManagedSet))
}
func (h *MinHeap) Pop() interface{} {
	old := h.data
	n := len(old)
	x := old[n-1]
	h.data = old[0 : n-1]
	return x
}

func (h *MinHeap) Add(c Managed) *Fsm {
	for _, cg := range h.data {
		if h.EqualFunc(cg.GetKey(), c.GetKey()) {
			return cg.Add(c)
		}
	}
	// container's image doesn't match any known. create new
	cg := h.NewFunc(c.GetKey())
	heap.Push(h, cg)
	return cg.Add(c)
}

func (h *MinHeap) Remove(c Managed) {
	for i, cg := range h.data {
		if h.EqualFunc(cg.GetKey(), c.GetKey()) {
			cg.Remove(c)
			if cg.Empty() {
				heap.Remove(h, i)
				return // iteration now invalid
			}
			return
		}
	}
}

func (h *MinHeap) Visit(visit func(ManagedSet)) {
	if visit == nil {
		return
	}
	for _, cg := range h.data {
		visit(cg)
	}
}

func (h *MinHeap) Instances(groupBy interface{}) []*Fsm {
	for _, cg := range h.data {
		if h.EqualFunc(cg.GetKey(), groupBy) {
			return cg.Instances()
		}
	}
	return []*Fsm{}
}
