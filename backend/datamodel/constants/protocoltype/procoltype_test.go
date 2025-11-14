package protocoltype

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that the secure protocol types are indicated properly.
func TestProtocolTypeIsSecure(t *testing.T) {
	require.True(t, HTTPS.IsSecure())
	require.True(t, RNDC.IsSecure())

	require.False(t, HTTP.IsSecure())
	require.False(t, Socket.IsSecure())
}

// Test that parsing protocol types from strings works properly.
func TestParseProtocolType(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		pt, err := Parse("http")
		require.NoError(t, err)
		require.Equal(t, HTTP, pt)
	})

	t.Run("HTTPS", func(t *testing.T) {
		pt, err := Parse("https")
		require.NoError(t, err)
		require.Equal(t, HTTPS, pt)
	})

	t.Run("Socket", func(t *testing.T) {
		pt, err := Parse("unix")
		require.NoError(t, err)
		require.Equal(t, Socket, pt)
	})

	t.Run("RNDC", func(t *testing.T) {
		pt, err := Parse("rndc")
		require.NoError(t, err)
		require.Equal(t, RNDC, pt)
	})

	t.Run("Unknown protocol type", func(t *testing.T) {
		pt, err := Parse("unknown")
		require.Error(t, err)
		require.Equal(t, Unspecified, pt)
	})

	t.Run("HTTP uppercase", func(t *testing.T) {
		pt, err := Parse("HTTP")
		require.Error(t, err)
		require.Equal(t, Unspecified, pt)
	})
}
