package loadbalancer

import (
	"container/heap"
	"math/rand"
	"time"
)

type Random struct {
}

func InitRandom() *Random {
	return &Random{}
}

type RoundRobin struct {
}

func InitRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

type LeastConnections struct {
}

func InitLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

type PeakEWMA struct {
}

func InitPeakEWMA() *PeakEWMA {
	return &PeakEWMA{}
}

type P2C struct {
}

func InitP2C() *P2C {
	return &P2C{}
}

func (r *Random) NextEndpoint(backends *BackendHeap, reqId uint32) *Backend {
	index := rand.Intn(len(backends.backends))
	best := backends.backends[index]
	IncrementActiveConns(best)
	return best
}

func (r *Random) Completed(backends *BackendHeap, backend *Backend, latency time.Duration) {
	DecrementActiveConns(backend)
	UpdateLatencyStats(backend, latency)
}

func (r *Random) Name() string {
	return "Random"
}

func (rr *RoundRobin) NextEndpoint(backends *BackendHeap, reqId uint32) *Backend {
	index := reqId % uint32(len(backends.backends))
	best := backends.backends[index]
	IncrementActiveConns(best)
	return best
}

func (rr *RoundRobin) Completed(backends *BackendHeap, backend *Backend, latency time.Duration) {
	DecrementActiveConns(backend)
	UpdateLatencyStats(backend, latency)
}

func (rr *RoundRobin) Name() string {
	return "Round Robin"
}

func (lc *LeastConnections) NextEndpoint(backends *BackendHeap, reqId uint32) *Backend {
	if backends.Len() > 0 {
		best := backends.backends[0]
		IncrementActiveConns(best)
		heap.Fix(backends, best.Index)
		return best
	}

	return nil
}

func (lc *LeastConnections) Completed(backends *BackendHeap, backend *Backend, latency time.Duration) {
	DecrementActiveConns(backend)
	UpdateLatencyStats(backend, latency)
	heap.Fix(backends, backend.Index)
}

func (lc *LeastConnections) Name() string {
	return "Least Connections"
}

func (peakEWMA *PeakEWMA) NextEndpoint(backends *BackendHeap, reqId uint32) *Backend {
	if backends.Len() > 0 {
		best := backends.backends[0]
		IncrementActiveConns(best)
		heap.Fix(backends, best.Index)
		return best
	}

	return nil
}

func (peakEWMA *PeakEWMA) Completed(backends *BackendHeap, backend *Backend, latency time.Duration) {
	DecrementActiveConns(backend)
	UpdateLatencyStats(backend, latency)
	heap.Fix(backends, backend.Index)
}

func (peakEWMA *PeakEWMA) Name() string {
	return "Peak EWMA"
}

func (p2c *P2C) NextEndpoint(backends *BackendHeap, reqId uint32) *Backend {
	idx1 := rand.Intn(len(backends.backends))
	idx2 := rand.Intn(len(backends.backends))

	backend1 := backends.backends[idx1]
	backend2 := backends.backends[idx2]

	iVal := backend1.LatencyStats.Load()
	jVal := backend2.LatencyStats.Load()

	if iVal == 0 {
		return backend1
	}

	if jVal == 0 {
		return backend2
	}

	iLatency := time.Duration(iVal)
	jLatency := time.Duration(jVal)

	iCost := float64(iLatency) * float64(backend1.ActiveConns+1)
	jCost := float64(jLatency) * float64(backend2.ActiveConns+1)

	if iCost < jCost {
		IncrementActiveConns(backend1)
		return backend1
	}

	IncrementActiveConns(backend2)
	return backend2
}

func (p2c *P2C) Completed(backends *BackendHeap, backend *Backend, latency time.Duration) {
	DecrementActiveConns(backend)
	UpdateLatencyStats(backend, latency)
}

func (p2c *P2C) Name() string {
	return "Power of Two Choices"
}
