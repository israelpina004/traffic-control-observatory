package loadbalancer

import (
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          string
	ActiveConns  int64
	LatencyStats atomic.Int64
	Index        int
}

func IncrementActiveConns(b *Backend) {
	atomic.AddInt64(&b.ActiveConns, 1)
}

func DecrementActiveConns(b *Backend) {
	atomic.AddInt64(&b.ActiveConns, -1)
}

const ewmaAlpha = 0.2

func UpdateLatencyStats(b *Backend, latency time.Duration) {
	currLatency := b.LatencyStats.Load()

	if currLatency == 0 {
		b.LatencyStats.Store(int64(latency))
		return
	}

	oldEWMA := time.Duration(currLatency)
	newEWMA := time.Duration(float64(oldEWMA)*ewmaAlpha + float64(latency)*(1-ewmaAlpha))
	b.LatencyStats.Store(int64(newEWMA))
}

type Strategy interface {
	NextEndpoint(backends *BackendHeap, reqId uint32) *Backend
	Completed(backends *BackendHeap, backend *Backend, latency time.Duration)
	Name() string
}
