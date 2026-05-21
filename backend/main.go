package main

import (
	"bytes"
	"fmt"
	"os"

	"hash/crc32"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/user/traffic-control-observatory/api"
	"github.com/user/traffic-control-observatory/loadbalancer"
	"github.com/user/traffic-control-observatory/telemetry"
	"github.com/user/traffic-control-observatory/workers"
)

const (
	PORT        = ":8080"
	VM_IP       = "127.0.0.1" // The GCP Server IP
	CONCURRENCY = 25
)

var (
	backends = []*loadbalancer.Backend{
		{
			URL:   VM_IP + ":9001",
			Index: 0,
		},
		{
			URL:   VM_IP + ":9002",
			Index: 1,
		},
		{
			URL:   VM_IP + ":9003",
			Index: 2,
		},
		{
			URL:   VM_IP + ":9004",
			Index: 3,
		},
		{
			URL:   VM_IP + ":9005",
			Index: 4,
		},
		{
			URL:   VM_IP + ":9006",
			Index: 5,
		},
		{
			URL:   VM_IP + ":9007",
			Index: 6,
		},
		{
			URL:   VM_IP + ":9008",
			Index: 7,
		},
		{
			URL:   VM_IP + ":9009",
			Index: 8,
		},
		{
			URL:   VM_IP + ":9010",
			Index: 9,
		},
	}

	backendsMu = sync.Mutex{}
)

var rrCounter uint32

type LoadBalancer struct {
	Strategy    loadbalancer.Strategy
	Backends    []*loadbalancer.Backend
	BackendHeap *loadbalancer.BackendHeap
	Mu          sync.Mutex
}

func (lb *LoadBalancer) Init(strategy string) {
	lb.Mu.Lock()
	defer lb.Mu.Unlock()
	switch strategy {
	case "random":
		lb.Strategy = loadbalancer.InitRandom()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			return true
		})
	case "round_robin":
		lb.Strategy = loadbalancer.InitRoundRobin()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			return true
		})
	case "least_connections":
		lb.Strategy = loadbalancer.InitLeastConnections()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			return i.ActiveConns < j.ActiveConns
		})
	case "peak_ewma":
		lb.Strategy = loadbalancer.InitPeakEWMA()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			costI := i.LatencyStats.Load() * (i.ActiveConns + 1)
			costJ := j.LatencyStats.Load() * (j.ActiveConns + 1)

			return costI < costJ
		})
	case "p2c":
		lb.Strategy = loadbalancer.InitP2C()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			return true
		})
	default:
		lb.Strategy = loadbalancer.InitRandom()
		lb.BackendHeap = loadbalancer.InitBackendHeap(lb.Backends, func(i, j *loadbalancer.Backend) bool {
			return true
		})
	}
}

func main() {
	// Start workers
	runner := workers.NewRunner(VM_IP+PORT, CONCURRENCY)

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8081"
	}

	// Start API Server
	apiServer := api.NewServer(":"+apiPort, runner)
	go func() {
		if err := apiServer.Start(); err != nil {
			fmt.Printf("Failed to start API server: %v\n", err)
		}
	}()

	strategy := <-apiServer.StartSimulationChan

	lb := LoadBalancer{
		Backends: backends,
	}

	lb.Init(strategy)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("Failed to bind to %s: %v\n", PORT, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Traffic Control Observatory LB listening on %s using %s strategy\n", PORT, strategy)
	telemetrySystem := telemetry.NewRingBuffer(1024 * 16)

	go func() {
		for newStrategy := range apiServer.StartSimulationChan {
			fmt.Printf("Changing LB strategy to %s...\n", newStrategy)
			lb.Init(newStrategy)
		}
	}()

	// Main Accept Loop
	go telemetryDrainer(telemetrySystem, apiServer)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		// Spawn Worker (One Goroutine Per Connection)
		// We pass the telemetry system to the worker
		go handleConnection(conn, telemetrySystem, &lb)
	}
}

// Drains the telemetry ring buffer and sends events to the API server.
func telemetryDrainer(rb *telemetry.RingBuffer, apiServer *api.Server) {
	ticker := time.NewTicker(16 * time.Millisecond)

	defer ticker.Stop()

	for range ticker.C {
		events := rb.PopAll()

		if len(events) > 0 {
			apiServer.BroadcastEvents(events, backends)
			// fmt.Printf("Broadcasted %d events to UI. (Head: %d, Tail: %d)\n", len(events), rb.Head(), rb.Tail())
		}
	}
}

// Handles a single connection, from start to end.
func handleConnection(conn net.Conn, telemetrySystem *telemetry.RingBuffer, lb *LoadBalancer) {
	defer conn.Close()

	// Generate Request ID and Source Hash
	reqId := atomic.AddUint32(&rrCounter, 1)
	clientIP := conn.RemoteAddr().String()
	hashClientIP := crc32.ChecksumIEEE([]byte(clientIP))

	// Emit Telemetry: EventStart
	event := telemetry.PackedEvent{
		ReqID:      uint32(reqId),
		Type:       telemetry.EventStart,
		BackendID:  0,
		Metric:     0,
		SourceHash: hashClientIP,
		Padding:    0,
	}

	// Push to telemetry system (ring buffer)
	telemetrySystem.Push(event)

	// Read Initial Bytes (Header detection)
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil { // Emit Telemetry: EventError if reading data from client fails
		event = telemetry.PackedEvent{
			ReqID: uint32(reqId),
			Type:  telemetry.EventError,
		}
		telemetrySystem.Push(event)

		fmt.Printf("Request %d failed to read from connection: %v\n", reqId, err)
		return
	}

	// Select Backend
	backendsMu.Lock()
	backend := lb.Strategy.NextEndpoint(lb.BackendHeap, uint32(reqId))
	backendsMu.Unlock()

	// Record start time
	startTime := time.Now()
	requestSuccessful := false

	// Emit Telemetry: EventEnd / EventError
	// We use a defer to ensure that the event is sent even if the connection is closed unexpectedly.
	// We also record the time it took to proxy the data to the backend.
	defer func() {
		latency := time.Since(startTime)

		var eventType telemetry.EventType
		if !requestSuccessful {
			latency = 10 * time.Second
			eventType = telemetry.EventError
		} else {
			eventType = telemetry.EventEnd
		}

		event = telemetry.PackedEvent{
			ReqID:     uint32(reqId),
			Type:      eventType,
			BackendID: uint8(backend.Index),
		}

		telemetrySystem.Push(event)
		backendsMu.Lock()
		lb.Strategy.Completed(lb.BackendHeap, backend, latency)
		backendsMu.Unlock()
	}()

	// Dial Backend
	dest, err := net.Dial("tcp", backend.URL)
	if err != nil {
		fmt.Printf("Request %d failed to connect to backend: %v\n", reqId, err)
		return
	}
	defer dest.Close()

	// Emit Telemetry: EventBackendSent
	event = telemetry.PackedEvent{
		ReqID:     uint32(reqId),
		Type:      telemetry.EventBackendSent,
		BackendID: uint8(backend.Index),
	}

	telemetrySystem.Push(event)

	// Proxy Data (io.Copy or manual buffer loop)
	serverReader := io.MultiReader(bytes.NewReader(buffer[:n]), conn)

	go io.Copy(dest, serverReader)

	io.Copy(conn, dest)

	requestSuccessful = true
}
