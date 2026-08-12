// Package history provides bounded in-memory sample storage.
package history

import "fmt"

// Ring is a fixed-capacity circular buffer. It allocates its backing storage
// once and overwrites the oldest value when full.
type Ring[T any] struct {
	values []T
	start  int
	length int
}

// NewRing constructs an empty ring with a positive fixed capacity.
func NewRing[T any](capacity int) (*Ring[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("history capacity must be positive")
	}
	return &Ring[T]{values: make([]T, capacity)}, nil
}

// Capacity returns the maximum number of retained values.
func (ring *Ring[T]) Capacity() int {
	return len(ring.values)
}

// Len returns the current number of retained values.
func (ring *Ring[T]) Len() int {
	return ring.length
}

// Append adds a value, evicting the oldest value when the ring is full.
func (ring *Ring[T]) Append(value T) {
	if ring.length < len(ring.values) {
		index := (ring.start + ring.length) % len(ring.values)
		ring.values[index] = value
		ring.length++
		return
	}

	ring.values[ring.start] = value
	ring.start = (ring.start + 1) % len(ring.values)
}

// Values returns an independent oldest-to-newest copy of retained values.
func (ring *Ring[T]) Values() []T {
	result := make([]T, ring.length)
	for index := range ring.length {
		result[index] = ring.values[(ring.start+index)%len(ring.values)]
	}
	return result
}
