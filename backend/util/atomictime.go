package storkutil

import (
	"sync/atomic"
	"time"
)

// AtomicTime is a wrapper around atomic.Value that allows to store and load
// time.Time values atomically.
type AtomicTime struct {
	// From atomic docs: A Value must not be copied after first use.
	value *atomic.Value
}

// Constructs a new AtomicTime instance.
func NewAtomicTime() AtomicTime {
	return AtomicTime{
		value: &atomic.Value{},
	}
}

// Loads the value from the atomic storage.
func (at AtomicTime) Load() (time.Time, bool) {
	v := at.value.Load()
	if v == nil {
		return time.Time{}, false
	}
	return v.(time.Time), true
}

// Stores the value in the atomic storage.
func (at AtomicTime) Store(t time.Time) {
	at.value.Store(t)
}

// Compares the value in the atomic storage with the expected value and if they
// are the same, stores the new value.
func (at AtomicTime) CompareAndSwap(old, new time.Time) bool {
	return at.value.CompareAndSwap(old, new)
}

// Swaps the value in the atomic storage with the new value and returns the old
// value.
func (at AtomicTime) Swap(new time.Time) time.Time {
	timeOrNil := at.value.Swap(new)
	if timeOrNil == nil {
		return time.Time{}
	}
	return timeOrNil.(time.Time)
}
