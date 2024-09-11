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
