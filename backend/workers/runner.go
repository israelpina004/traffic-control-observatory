package workers

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Runner struct {
	targetAddr  string
	concurrency int
	cancelChan  chan struct{}
	mu          sync.Mutex
	wg          sync.WaitGroup
	ops         uint64
	errs        uint64
}

func NewRunner(targetAddr string, concurrency int) *Runner {
	return &Runner{
		targetAddr:  targetAddr,
		concurrency: concurrency,
	}
}

func (r *Runner) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancelChan = make(chan struct{})
	r.ops = 0
	r.errs = 0

	for i := 0; i < r.concurrency; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				select {
				case <-r.cancelChan:
					return
				default:
					if r.sendRequest() {
						atomic.AddUint64(&r.ops, 1)
					} else {
						atomic.AddUint64(&r.errs, 1)
					}

					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}

func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cancelChan == nil {
		return
	}

	close(r.cancelChan)
	r.wg.Wait()
	r.cancelChan = nil

	fmt.Printf("\n--- Results ---\n")
	fmt.Printf("Total Requests: %d\n", r.ops)
	fmt.Printf("Failed Requests: %d\n", r.errs)
}

func (r *Runner) sendRequest() bool {
	conn, err := net.Dial("tcp", r.targetAddr)
	if err != nil {
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Hello"))
	if err != nil {
		return false
	}

	_, err = conn.Read(make([]byte, 1024))
	if err != nil {
		return false
	}

	return true
}
