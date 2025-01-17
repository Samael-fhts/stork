package keaconfig

import (
	dhcpmodel "isc.org/stork/datamodel/dhcp"
)

// An interface to a structure providing option definition lookup capabilities
// for a given daemon.
type DHCPOptionDefinitionLookup interface {
	// Checks if a definition of the specified option exists for the
	// given daemon.
	DefinitionExists(dhcpmodel.DHCPOptionAccessor) bool
	// Searches for an option definition for the specified daemon ID and option value.
	Find(dhcpmodel.DHCPOptionAccessor) DHCPOptionDefinitionAccessor
}

// Represents a DHCP option definition lookup mechanism.
type definitionLookup map[string]map[uint16]DHCPOptionDefinition

// Creates a new instance of the DHCPOptionDefinitionLookup structure.
// It accepts a list of custom definitions that will be merged with the
// standard definitions.
func NewDHCPOptionDefinitionLookup(customDefinitions []DHCPOptionDefinition) DHCPOptionDefinitionLookup {
	var allDefinitions []DHCPOptionDefinition

	dhcp4Definitions := GetStandardDHCPv4OptionDefinitions()
	allDefinitions = append(allDefinitions, dhcp4Definitions...)

	dhcp6Definitions := GetStandardDHCPv6OptionDefinitions()
	allDefinitions = append(allDefinitions, dhcp6Definitions...)

	allDefinitions = append(allDefinitions, customDefinitions...)

	lookup := make(definitionLookup)
	for _, definition := range allDefinitions {
		if _, ok := lookup[definition.Space]; !ok {
			lookup[definition.Space] = make(map[uint16]DHCPOptionDefinition)
		}
		lookup[definition.Space][definition.Code] = definition
	}

	return lookup
}

// Checks if a definition of the specified option exists.
func (lookup definitionLookup) DefinitionExists(option dhcpmodel.DHCPOptionAccessor) bool {
	if definitions, ok := lookup[option.GetSpace()]; ok {
		_, ok := definitions[option.GetCode()]
		return ok
	}
	return false
}

// Searches for an option definition for the specified daemon ID and option value.
func (lookup definitionLookup) Find(option dhcpmodel.DHCPOptionAccessor) DHCPOptionDefinitionAccessor {
	if definitions, ok := lookup[option.GetSpace()]; ok {
		if definition, ok := definitions[option.GetCode()]; ok {
			return definition
		}
	}
	return nil
}
