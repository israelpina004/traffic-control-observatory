package main

import (
	"fmt"
	"time"

	// "math/rand"
	"net"
	"sync"
	"sync/atomic"
)

const (
	TargetAddr  = "127.0.0.1:8080"
	Concurrency = 50
	// Duration    = 10 * time.Second
)

func main() {
	var (
		ops  uint64 // Success operations
		errs uint64 // Failed operations
	)

	// fmt.Printf("Starting benchmark with target %s and %d workers for %v seconds...\n", TargetAddr, Concurrency, Duration)

	// start := time.Now()
	// done := make(chan bool)

	// // Timer to stop benchmark
	// go func() {
	// 	time.Sleep(Duration)
	// 	close(done)
	// }()

	var wg sync.WaitGroup
	wg.Add(Concurrency)

	for i := 0; i < Concurrency; i++ {
		go func() {
			defer wg.Done()
			// isSlow := rand.Intn(2) == 0
			for {
				select {
				// case <-done:
				// 	return
				default:
					if sendRequest() { // Pass isSlow to simulate slow clients
						atomic.AddUint64(&ops, 1)
					} else {
						atomic.AddUint64(&errs, 1)
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()
	// elapsed := time.Since(start)

	fmt.Println("\n--- Results ---")
	// fmt.Printf("Time: %v\n", elapsed)
	fmt.Printf("Total Requests: %d\n", ops)
	fmt.Printf("Failed Requests: %d\n", errs)
	// fmt.Printf("RPS: %.2f\n", float64(ops)/elapsed.Seconds())
}

// Sends a request to the load balancer. If isSlow is true, it will sleep for 200ms before sending the request, simulating a slow client.
func sendRequest() bool { // isSlow bool
	// Dial the load balancer

	// Simulate slow client
	// if isSlow {
	// 	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	// }

	conn, err := net.Dial("tcp", TargetAddr)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Write some data to the load balancer
	_, err = conn.Write([]byte("Hello"))
	if err != nil {
		return false
	}

	// Read the response from the load balancer
	_, err = conn.Read(make([]byte, 1024))
	if err != nil {
		return false
	}

	return true
}
