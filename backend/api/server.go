package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/user/traffic-control-observatory/loadbalancer"
	"github.com/user/traffic-control-observatory/telemetry"
	"github.com/user/traffic-control-observatory/workers"
)

// A connected client. We set up a channel to send messages to the client.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type BackendInfo struct {
	Routed    int64
	Completed int64
	Errored   int64
	Latency   int64
}

// Writes messages to the client.
func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		msg, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Printf("Failed to write to WebSocket: %v\n", err)
			return
		}
	}
}

// Handles the REST API and WebSocket connections for the frontend
type Server struct {
	port                string
	clients             map[*Client]bool
	clientsMu           sync.Mutex
	upgrader            websocket.Upgrader
	backendStats        map[uint8]*BackendInfo
	StartSimulationChan chan string
	runner              *workers.Runner
}

// Creates new API server
func NewServer(port string, runner *workers.Runner) *Server {
	return &Server{
		port:    port,
		clients: make(map[*Client]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		backendStats:        make(map[uint8]*BackendInfo),
		StartSimulationChan: make(chan string),
		runner:              runner,
	}
}

// Begins HTTP and WebSocket servers.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// WebSocket endpoints
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/stats", s.handleGetBackendStats)
	mux.HandleFunc("/api/start", s.startSimulationHandler)
	mux.HandleFunc("/api/workers/toggle", s.toggleWorkersHandler)

	fmt.Printf("Visualization API & WebSockets listening on %s\n", s.port)

	return http.ListenAndServe(s.port, mux)
}

type toggleReq struct {
	Running bool `json:"running"`
}

func (s *Server) toggleWorkersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req toggleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode toggle request", http.StatusBadRequest)
		return
	}

	if req.Running {
		s.runner.Start()
	} else {
		s.runner.Stop()
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

type StrategyReq struct {
	Strategy string `json:"strategy"`
}

func (s *Server) startSimulationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StrategyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode strategy request", http.StatusBadRequest)
		return
	}

	s.StartSimulationChan <- req.Strategy

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// Handles the REST API endpoint for getting backend stats
func (s *Server) handleGetBackendStats(w http.ResponseWriter, r *http.Request) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Write the backend stats to the response
	if err := json.NewEncoder(w).Encode(s.backendStats); err != nil {
		http.Error(w, "Failed to encode backend stats", http.StatusInternalServerError)
		return
	}
}

// Handles WebSocket connections for real-time telemetry events.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade WebSocket connection: %v\n", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()

	fmt.Println("Client connected.")

	go client.writePump()

	// Listen for disconnects so we can clean up the connection.
	go func() {
		defer func() {
			s.clientsMu.Lock()
			delete(s.clients, client)
			s.clientsMu.Unlock()
			close(client.send)
			fmt.Println("Client disconnected.")
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// Sends an array of telemetry events to all connected WebSocket clients.
// Also updates the backend stats.
func (s *Server) BroadcastEvents(events []telemetry.PackedEvent, backends []*loadbalancer.Backend) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	// Update backend stats
	for _, e := range events {
		if _, ok := s.backendStats[e.BackendID]; !ok {
			s.backendStats[e.BackendID] = &BackendInfo{
				Routed:    0,
				Completed: 0,
				Errored:   0,
			}
		}

		switch e.Type {
		case telemetry.EventBackendSent:
			s.backendStats[e.BackendID].Routed++
		case telemetry.EventEnd:
			s.backendStats[e.BackendID].Completed++
		case telemetry.EventError:
			s.backendStats[e.BackendID].Errored++
		}
	}

	for i := range backends {
		if _, ok := s.backendStats[uint8(i)]; !ok {
			s.backendStats[uint8(i)] = &BackendInfo{
				Latency: backends[i].LatencyStats.Load(),
			}
		} else {
			s.backendStats[uint8(i)].Latency = backends[i].LatencyStats.Load()
		}
	}

	if len(s.clients) == 0 {
		return
	}

	// Convert events to JSON
	data, err := json.Marshal(events)
	if err != nil {
		fmt.Printf("Failed to marshal telemetry events: %v\n", err)
		return
	}

	// Broadcast to all connected clients
	for conn := range s.clients {
		select {
		case conn.send <- data:
			continue
		default:
			conn.conn.Close()
		}
	}
}
