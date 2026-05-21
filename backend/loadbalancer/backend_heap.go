package loadbalancer

import "container/heap"

type LessFunc func(i, j *Backend) bool

type BackendHeap struct {
	backends []*Backend
	less     LessFunc
}

func InitBackendHeap(backends []*Backend, less LessFunc) *BackendHeap {
	h := &BackendHeap{
		backends: backends,
		less:     less,
	}
	heap.Init(h)
	return h
}

func (h BackendHeap) Len() int {
	return len(h.backends)
}

func (h BackendHeap) Less(i, j int) bool {
	return h.less(h.backends[i], h.backends[j])
}

func (h BackendHeap) Swap(i, j int) {
	h.backends[i], h.backends[j] = h.backends[j], h.backends[i]
	h.backends[i].Index = i
	h.backends[j].Index = j
}

func (h *BackendHeap) Push(x any) {
	item := x.(*Backend)
	item.Index = len(h.backends)
	h.backends = append(h.backends, item)
}

func (h *BackendHeap) Pop() any {
	old := h.backends
	n := len(old)
	item := old[n-1]
	h.backends = old[0 : n-1]
	return item
}
