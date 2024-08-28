package dbmodel

import (
	keaconfig "isc.org/stork/appcfg/kea"
	dhcpmodel "isc.org/stork/datamodel/dhcp"
)

// Groups DHCP option definitions by space and code.
type definitionStore map[string]map[uint16]keaconfig.DHCPOptionDefinition

func (s definitionStore) find(space string, code uint16) *keaconfig.DHCPOptionDefinition {
	if definitions, ok := s[space]; ok {
		if definition, ok := definitions[code]; ok {
			return &definition
		}
	}
	return nil
}

// DHCP option definition lookup mechanism.
type DHCPOptionDefinitionLookup struct {
	standardDefinitions definitionStore
	customDefinitions   map[int64]definitionStore
}

// Creates new lookup instance.
func NewDHCPOptionDefinitionLookup() *DHCPOptionDefinitionLookup {
	dhcp4Definitions := keaconfig.GetStandardDHCPv4OptionDefinitions()
	dhcp6Definitions := keaconfig.GetStandardDHCPv6OptionDefinitions()
	var allDefinitions []keaconfig.DHCPOptionDefinition
	allDefinitions = append(allDefinitions, dhcp4Definitions...)
	allDefinitions = append(allDefinitions, dhcp6Definitions...)

	standardStore := make(definitionStore)
	for _, definition := range allDefinitions {
		if _, ok := standardStore[definition.Space]; !ok {
			standardStore[definition.Space] = make(map[uint16]keaconfig.DHCPOptionDefinition)
		}
		standardStore[definition.Space][definition.Code] = definition
	}

	return &DHCPOptionDefinitionLookup{
		standardDefinitions: standardStore,
		customDefinitions:   make(map[int64]definitionStore),
	}
}

// Checks if a definition of the specified option exists for the
// given daemon.
func (lookup DHCPOptionDefinitionLookup) DefinitionExists(daemonID int64, option dhcpmodel.DHCPOptionAccessor) bool {
	return lookup.Find(daemonID, option) != nil
}

// Finds option definition for the specified option. Internally, it queries standard
// Kea option definitions defined in the keaconfig package. In the future it will also
// be able to search for the runtime definitions in the database.
func (lookup DHCPOptionDefinitionLookup) Find(daemonID int64, option dhcpmodel.DHCPOptionAccessor) keaconfig.DHCPOptionDefinitionAccessor {
	// Check if the option is a standard one.
	if definition := lookup.standardDefinitions.find(option.GetSpace(), option.GetCode()); definition != nil {
		return definition
	}

	// Check the custom definitions.
	if definitions, ok := lookup.customDefinitions[daemonID]; ok {
		return definitions.find(option.GetSpace(), option.GetCode())
	}

	return nil
}
