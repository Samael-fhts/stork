package storkutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test that the atomic time is properly created.
func TestNewAtomicTime(t *testing.T) {
	// Arrange & Act
	atomicTime := NewAtomicTime()

	// Assert
	require.NotNil(t, atomicTime)
	require.NotNil(t, atomicTime.value)
}

// Test that the empty atomic time returns the zero time when loaded.
func TestAtomicTimeLoadEmpty(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()

	// Act
	loadedTime, ok := atomicTime.Load()

	// Assert
	require.False(t, ok)
	require.Equal(t, time.Time{}, loadedTime)
}

// Test that the atomic time is properly stored and loaded.
func TestAtomicTimeStoreAndLoad(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()
	expectedTime := time.Date(2023, time.February, 16, 17, 30, 0, 0, time.UTC)

	// Act
	atomicTime.Store(expectedTime)
	loadedTime, ok := atomicTime.Load()

	// Assert
	require.True(t, ok)
	require.Equal(t, expectedTime, loadedTime)
}

// Test that the atomic time is properly swapped if the atomic time is empty.
func TestAtomicTimeSwapEmpty(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()
	expectedTime := time.Date(2023, time.February, 16, 17, 30, 0, 0, time.UTC)

	// Act
	previousTime := atomicTime.Swap(expectedTime)

	// Assert
	require.Equal(t, time.Time{}, previousTime)
	loadedTime, ok := atomicTime.Load()
	require.True(t, ok)
	require.Equal(t, expectedTime, loadedTime)
}

// Test that the atomic time is properly swapped if the atomic time is not empty.
func TestAtomicTimeSwapNotEmpty(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()
	expectedTime := time.Date(2023, time.February, 16, 17, 30, 0, 0, time.UTC)
	previousTime := time.Date(2023, time.February, 16, 17, 29, 0, 0, time.UTC)
	atomicTime.Store(previousTime)

	// Act
	previousTime = atomicTime.Swap(expectedTime)

	// Assert
	require.Equal(t, previousTime, previousTime)
	loadedTime, ok := atomicTime.Load()
	require.True(t, ok)
	require.Equal(t, expectedTime, loadedTime)
}

// Test that the atomic time is properly compared and swapped if the atomic time is empty.
func TestAtomicTimeCompareAndSwapEmpty(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()
	expectedTime := time.Date(2023, time.February, 16, 17, 30, 0, 0, time.UTC)

	// Act
	swapped := atomicTime.CompareAndSwap(time.Time{}, expectedTime)

	// Assert
	require.False(t, swapped)
	loadedTime, ok := atomicTime.Load()
	require.False(t, ok)
	require.Equal(t, time.Time{}, loadedTime)
}

// Test that the atomic time is properly compared and swapped if the atomic time is not empty.
func TestAtomicTimeCompareAndSwapNotEmpty(t *testing.T) {
	// Arrange
	atomicTime := NewAtomicTime()
	expectedTime := time.Date(2023, time.February, 16, 17, 30, 0, 0, time.UTC)
	previousTime := time.Date(2023, time.February, 16, 17, 29, 0, 0, time.UTC)
	atomicTime.Store(previousTime)

	// Act
	swapped := atomicTime.CompareAndSwap(previousTime, expectedTime)

	// Assert
	require.True(t, swapped)
	loadedTime, ok := atomicTime.Load()
	require.True(t, ok)
	require.Equal(t, expectedTime, loadedTime)
}
