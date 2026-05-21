package telemetry

import (
	"encoding/binary"
)

// EventType defines the lifecycle stage of a request.
type EventType uint8

const (
	EventStart       EventType = 0
	EventBackendSent EventType = 1
	EventEnd         EventType = 2
	EventError       EventType = 3
)

// 16-byte structure sent to the frontend.
// It corresponds to the layout defined in the Architecture contract.
// Total Size: 4 + 1 + 1 + 2 + 4 + 4 = 16 bytes.
type PackedEvent struct {
	ReqID      uint32    // 0-4
	Type       EventType // 4-5
	BackendID  uint8     // 5-6
	Metric     uint16    // 6-8 (Latency in ms or Payload Size)
	SourceHash uint32    // 8-12 (IP Hash)
	Padding    uint32    // 12-16 (Reserved/Flags)
}

// Writes the binary representation to a byte slice.
// This is the "hot path" - it must be fast.
func (e *PackedEvent) Encode(buf []byte) {
	// Boundary check (assuming buf is at least 16 bytes)
	_ = buf[15]

	binary.LittleEndian.PutUint32(buf[0:4], e.ReqID)
	buf[4] = uint8(e.Type)
	buf[5] = e.BackendID
	binary.LittleEndian.PutUint16(buf[6:8], e.Metric)
	binary.LittleEndian.PutUint32(buf[8:12], e.SourceHash)
	binary.LittleEndian.PutUint32(buf[12:16], e.Padding)
}

// Interface for the Telemetry System
type Aggregator interface {
	// Push adds an event to the ring buffer.
	// If the buffer is full, it should drop the event and return false.
	Push(event PackedEvent) bool
}
