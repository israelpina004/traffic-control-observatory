package telemetry

import (
	"sync"
)

// Define our ring buffer
type RingBuffer struct {
	data []PackedEvent
	head uint64
	tail uint64
	size uint64
	mu   sync.Mutex
}

// Creates a new ring buffer with the given size
func NewRingBuffer(size uint64) *RingBuffer {
	return &RingBuffer{
		data: make([]PackedEvent, size), // Fixed size storage
		head: 0,                         // Next index to write to
		tail: 0,                         // Next index to read from
		size: size,                      // Fixed size of the buffer
		mu:   sync.Mutex{},              // Mutex for thread safety
	}
}

func (rb *RingBuffer) Head() uint64 {
	return rb.head
}

func (rb *RingBuffer) Tail() uint64 {
	return rb.tail
}

// Adds an event to the ring buffer
func (rb *RingBuffer) Push(event PackedEvent) bool {
	// Lock the mutex to ensure thread safety
	// (If two threads try to push at the same time, one will block)
	rb.mu.Lock()
	defer rb.mu.Unlock()

	indexForWrite := rb.head % rb.size

	if rb.head-rb.tail >= rb.size {
		return false
	}

	rb.data[indexForWrite] = event
	rb.head++

	return true
}

// Removes all events from the ring buffer and returns them
func (rb *RingBuffer) PopAll() []PackedEvent {
	// Lock the mutex to ensure thread safety
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// If there are no events, return nil
	count := rb.head - rb.tail
	if count == 0 {
		return nil
	}

	// If there are events, copy them to the aggregator
	// Our final array is only as big as the number of events.
	aggregator := make([]PackedEvent, count)

	// Determine the starting index
	startIdx := rb.tail % rb.size

	// If the events are in a single chunk, copy them directly. copy(dst, src) is a heavily optimized function.
	if startIdx+count <= rb.size {
		copy(aggregator, rb.data[startIdx:startIdx+count])
	} else {
		// If the events wrap around the end of the array, copy them in two parts
		firstPartSize := rb.size - startIdx
		copy(aggregator[:firstPartSize], rb.data[startIdx:])
		copy(aggregator[firstPartSize:], rb.data[:count-firstPartSize])
	}

	// Update the tail
	rb.tail += count
	return aggregator
}
