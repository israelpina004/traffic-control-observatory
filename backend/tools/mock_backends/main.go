package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	startPort := 9001
	endPort := 9010

	for p := startPort; p <= endPort; p++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			startServer(port)
		}(p)
	}

	fmt.Printf("Started mock backends on ports %d-%d\n", startPort, endPort)
	wg.Wait()
}

func startServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to listen on %s: %v\n", addr, err)
		return
	}
	// Don't close listener, keep running

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Accept error on %d: %v\n", port, err)
			continue
		}
		go handle(conn, port)
	}
}

func handle(conn net.Conn, port int) {
	defer conn.Close()

	// Simulate latency in the backend servers

	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	// Simple TCP Echo/Response
	response := fmt.Sprintf("Backend %d OK", port)
	conn.Write([]byte(response))
}
