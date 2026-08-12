package history

import (
	"reflect"
	"testing"
)

func TestNewRingRejectsInvalidCapacity(t *testing.T) {
	t.Parallel()

	for _, capacity := range []int{0, -1} {
		if _, err := NewRing[int](capacity); err == nil {
			t.Fatalf("NewRing(%d) returned no error", capacity)
		}
	}
}

func TestRingRetainsOrderedValuesAndEvictsOldest(t *testing.T) {
	t.Parallel()

	ring, err := NewRing[int](3)
	if err != nil {
		t.Fatal(err)
	}
	if ring.Len() != 0 || ring.Capacity() != 3 || len(ring.Values()) != 0 {
		t.Fatalf("empty ring len=%d capacity=%d values=%v", ring.Len(), ring.Capacity(), ring.Values())
	}

	for _, value := range []int{1, 2, 3, 4, 5} {
		ring.Append(value)
	}
	if ring.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", ring.Len())
	}
	if got := ring.Values(); !reflect.DeepEqual(got, []int{3, 4, 5}) {
		t.Fatalf("Values() = %v, want [3 4 5]", got)
	}
}

func TestCapacityOneAndReturnedCopy(t *testing.T) {
	t.Parallel()

	ring, err := NewRing[int](1)
	if err != nil {
		t.Fatal(err)
	}
	ring.Append(1)
	ring.Append(2)
	values := ring.Values()
	values[0] = 99
	if got := ring.Values(); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("Values() after caller mutation = %v, want [2]", got)
	}
}
