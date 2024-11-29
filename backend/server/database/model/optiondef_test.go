package dbmodel_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keaconfig "isc.org/stork/appcfg/kea"
	dbmodel "isc.org/stork/server/database/model"
	dbmodeltest "isc.org/stork/server/database/model/test"
	dbtest "isc.org/stork/server/database/test"
	storkutil "isc.org/stork/util"
)

// Test that the option definition lookup can identify the standard options
// DHCPv4 options for which Kea should know their definitions.
func TestStandardDHCPv4OptionDefinitionExists(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	existingCodes := []uint16{99, 108, 155, 212, 213}
	for _, code := range existingCodes {
		option := dbmodel.DHCPOption{
			Code:     code,
			Space:    "dhcp4",
			Universe: storkutil.IPv4,
		}
		require.True(t, lookup.DefinitionExists(option))
	}
}

// Test that the option definition lookup indicates that the DHCPv4
// suboption does not have a definition.
func TestDHCPv4SuboptionDefinition(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	option := dbmodel.DHCPOption{
		Code:     15,
		Space:    "foo",
		Universe: storkutil.IPv4,
	}
	require.False(t, lookup.DefinitionExists(option))
}

// Test that the option definition lookup flags the standard options
// for which the definitions do not exist.
func TestStandardDHCPv4OptionDefinitionNotExists(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	nonExistingCodes := []uint16{0, 106, 165, 180, 215, 224}
	for _, code := range nonExistingCodes {
		option := dbmodel.DHCPOption{
			Code:     code,
			Space:    "dhcp4",
			Universe: storkutil.IPv4,
		}
		require.False(t, lookup.DefinitionExists(option))
	}
}

// Test that the option definition lookup can identify the standard options
// DHCPv6 options for which Kea should know their definitions.
func TestStandardDHCPv6OptionDefinitionExists(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	option := dbmodel.DHCPOption{
		Code:     103,
		Space:    "dhcp6",
		Universe: storkutil.IPv6,
	}
	require.True(t, lookup.DefinitionExists(option))
}

// Test that the option definition lookup flags the standard options
// for which the definitions do not exist.
func TestStandardDHCPv6OptionDefinitionNotExists(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	nonExistingCodes := []uint16{0, 145}
	for _, code := range nonExistingCodes {
		option := dbmodel.DHCPOption{
			Code:     code,
			Space:    "dhcp6",
			Universe: storkutil.IPv6,
		}
		require.False(t, lookup.DefinitionExists(option))
	}
}

// Test that the option definition lookup indicates that the DHCPv6
// suboption does not have a definition.
func TestDHCPv6SuboptionDefinition(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	option := dbmodel.DHCPOption{
		Code:     15,
		Space:    "foo",
		Universe: storkutil.IPv6,
	}
	require.False(t, lookup.DefinitionExists(option))
}

// Test that standard option definition exists for a non-top level
// option space.
func TestStandardDHCPv6OptionDefinitionInOtherSpace(t *testing.T) {
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)

	option := dbmodel.DHCPOption{
		Code:     89,
		Space:    "s46-cont-mape-options",
		Universe: storkutil.IPv6,
	}
	require.True(t, lookup.DefinitionExists(option))
}

// Test that option definition lookup can find a definition for a Kea
// standard option.
func TestFindStdDHCPOptionDefinition(t *testing.T) {
	option := dbmodel.DHCPOption{
		Code:     89,
		Space:    "s46-cont-mape-options",
		Universe: storkutil.IPv6,
	}
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)
	def := lookup.Find(option)
	require.NotNil(t, def)
}

// Test that nil value is returned if an option definition is not found.
func TestFindStdDHCPOptionDefinitionNotFound(t *testing.T) {
	option := dbmodel.DHCPOption{
		Code:     1,
		Space:    "foo",
		Universe: storkutil.IPv6,
	}
	lookup := keaconfig.NewDHCPOptionDefinitionLookup(nil)
	def := lookup.Find(option)
	require.Nil(t, def)
}

// Test that the lookup is properly returned.
func TestDHCPOptionDefinitionLookupsGetLookup(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	keaServer1, _ := dbmodeltest.NewKeaDHCPv4Server(db)
	require.NoError(t, keaServer1.Configure(`{ "Dhcp4": {
		"option-def": [
			{
				"name": "foo",
				"code": 1000,
				"space": "dhcp4",
				"type": "uint32"
			}
		]
	}}`))

	keaServer2, _ := dbmodeltest.NewKeaDHCPv4Server(db)
	require.NoError(t, keaServer2.Configure(`{ "Dhcp4": {
		"option-def": [
			{
				"name": "bar",
				"code": 1000,
				"space": "dhcp4",
				"type": "uint32"
			}
		]
	}}`))

	lookups := dbmodel.NewDHCPOptionDefinitionLookups(db)

	// Act
	lookup1, err1 := lookups.GetLookup(keaServer1.ID)
	lookup2, err2 := lookups.GetLookup(keaServer2.ID)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, lookup1)
	require.NotNil(t, lookup2)

	// Check the lookups return different definitions.
	definition1 := lookup1.Find(&dbmodel.DHCPOption{Code: 1000, Space: "dhcp4"})
	require.NotNil(t, definition1)
	require.Equal(t, "foo", definition1.GetName())
	require.EqualValues(t, 1000, definition1.GetCode())

	definition2 := lookup2.Find(&dbmodel.DHCPOption{Code: 1000, Space: "dhcp4"})
	require.NotNil(t, definition2)
	require.Equal(t, "bar", definition2.GetName())
	require.EqualValues(t, 1000, definition2.GetCode())
}
