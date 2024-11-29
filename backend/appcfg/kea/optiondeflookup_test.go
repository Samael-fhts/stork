package keaconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that the option definition lookup is created correctly for a nil
// custom definitions list.
func TestNewDHCPOptionDefinitionLookupForNilCustomDefinitions(t *testing.T) {
	// Arrange & Act
	lookup := NewDHCPOptionDefinitionLookup(nil)

	// Assert
	require.NotNil(t, lookup)
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "dhcp4", Code: 1}))
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "dhcp6", Code: 1}))
	require.False(t, lookup.DefinitionExists(&DHCPOption{Space: "dhcp4", Code: 1000}))
}

// Test that the option definition lookup is created correctly for a non-nil
// custom definitions list.
func TestNewDHCPOptionDefinitionLookupWithCustomDefinitions(t *testing.T) {
	// Arrange
	customDefinitions := []DHCPOptionDefinition{
		{Space: "foo", Code: 1000},
		{Space: "bar", Code: 1000},
	}

	// Act
	lookup := NewDHCPOptionDefinitionLookup(customDefinitions)

	// Assert
	require.NotNil(t, lookup)
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "dhcp4", Code: 1}))
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "dhcp6", Code: 1}))
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "foo", Code: 1000}))
	require.True(t, lookup.DefinitionExists(&DHCPOption{Space: "bar", Code: 1000}))
}

// Test that the option definition lookup returns the correct definition.
func TestDHCPOptionDefinitionLookupFind(t *testing.T) {
	// Arrange
	customDefinitions := []DHCPOptionDefinition{
		{Space: "foo", Code: 1000},
		{Space: "bar", Code: 1000},
	}
	lookup := NewDHCPOptionDefinitionLookup(customDefinitions)

	t.Run("Standard definition", func(t *testing.T) {
		// Act
		definition := lookup.Find(&DHCPOption{Space: "dhcp4", Code: 1})

		// Assert
		require.NotNil(t, definition)
		require.Equal(t, "dhcp4", definition.GetSpace())
		require.Equal(t, uint16(1), definition.GetCode())
		require.False(t, definition.GetArray())
		require.Empty(t, definition.GetEncapsulate())
		require.Equal(t, "subnet-mask", definition.GetName())
		require.Empty(t, definition.GetRecordTypes())
		require.Equal(t, "ipv4-address", definition.GetType())
	})

	t.Run("Standard definition with record types", func(t *testing.T) {
		// Act
		definition := lookup.Find(&DHCPOption{Space: "dhcp4", Code: 78})

		// Assert
		require.NotNil(t, definition)
		require.Equal(t, "dhcp4", definition.GetSpace())
		require.Equal(t, uint16(78), definition.GetCode())
		require.True(t, definition.GetArray())
		require.Empty(t, definition.GetEncapsulate())
		require.Equal(t, "slp-directory-agent", definition.GetName())
		require.Equal(t, []string{"bool", "ipv4-address"}, definition.GetRecordTypes())
		require.Equal(t, "record", definition.GetType())
	})

	t.Run("Custom definition", func(t *testing.T) {
		// Act
		definition := lookup.Find(&DHCPOption{Space: "foo", Code: 1000})

		// Assert
		require.NotNil(t, definition)
		require.Equal(t, "foo", definition.GetSpace())
		require.Equal(t, uint16(1000), definition.GetCode())
		require.False(t, definition.GetArray())
		require.Empty(t, definition.GetEncapsulate())
		require.Empty(t, definition.GetName())
		require.Empty(t, definition.GetRecordTypes())
		require.Empty(t, definition.GetType())
	})
}
