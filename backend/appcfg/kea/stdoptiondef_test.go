package keaconfig

import (
	"testing"

	require "github.com/stretchr/testify/require"
)

// Finds an option definition by code and space.
func findOptionDefinition(definitions []DHCPOptionDefinition, code uint16, space string) *DHCPOptionDefinition {
	for _, def := range definitions {
		if def.Code == code && def.Space == space {
			return &def
		}
	}
	return nil
}

// Test that a DHCPv4 option definition can be found by code and space.
func TestFindDHCPv4OptionDefinition(t *testing.T) {
	definitions := GetStandardDHCPv4OptionDefinitions()
	def := findOptionDefinition(definitions, 72, "dhcp4")
	require.NotNil(t, def)
	//  Validate the option definition.
	require.True(t, def.GetArray())
	require.EqualValues(t, 72, def.GetCode())
	require.Empty(t, "", def.GetEncapsulate())
	require.Equal(t, "www-server", def.GetName())
	require.Zero(t, def.GetRecordTypes())
	require.Equal(t, "dhcp4", def.GetSpace())
	require.Equal(t, IPv4AddressOption, def.GetType())
}

// Test that nil is returned when searched option definition is not found.
func TestFindDHCPv4OptionDefinitionNotExists(t *testing.T) {
	definitions := GetStandardDHCPv4OptionDefinitions()
	def := findOptionDefinition(definitions, 150, "dhcp4")
	require.Nil(t, def)
}

// Test that a DHCPv6 option definition can be found by code and space.
func TestFindDHCPv6OptionDefinition(t *testing.T) {
	definitions := GetStandardDHCPv6OptionDefinitions()
	def := findOptionDefinition(definitions, 89, "s46-cont-mape-options")
	require.NotNil(t, def)
	//  Validate the option definition.
	require.False(t, def.GetArray())
	require.EqualValues(t, 89, def.GetCode())
	require.Equal(t, "s46-rule-options", def.GetEncapsulate())
	require.Equal(t, "s46-rule", def.GetName())
	require.Len(t, def.GetRecordTypes(), 5)
	require.Equal(t, "s46-cont-mape-options", def.GetSpace())
	require.Equal(t, RecordOption, def.GetType())
}

// Test that nil is returned when searched option definition is not found.
func TestFindDHCPv6OptionDefinitionNotExists(t *testing.T) {
	definitions := GetStandardDHCPv4OptionDefinitions()
	def := findOptionDefinition(definitions, 11, "foo")
	require.Nil(t, def)
}
