package dbmodel

import (
	"sync"

	keaconfig "isc.org/stork/appcfg/kea"
	dhcpmodel "isc.org/stork/datamodel/dhcp"
)

// Groups DHCP option definitions by space and code.
type definitionStore map[string]map[uint16]keaconfig.DHCPOptionDefinition

// Returns the definition of the specified option if it exists. Otherwise, it
// returns nil.
func (s definitionStore) find(space string, code uint16) *keaconfig.DHCPOptionDefinition {
	if definitions, ok := s[space]; ok {
		if definition, ok := definitions[code]; ok {
			return &definition
		}
	}
	return nil
}

// Checks if a definition of the specified option exists.
func (s definitionStore) has(space string, code uint16) bool {
	if definitions, ok := s[space]; ok {
		_, ok := definitions[code]
		return ok
	}
	return false
}

// DHCP option definition lookup mechanism.
type DHCPOptionDefinitionLookup struct {
	standardDefinitions definitionStore
	customDefinitions   map[int64]definitionStore
	// It must be stored as a pointer because the methods of this structure
	// use the reference receiver.
	customDefinitionMutex *sync.RWMutex
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
	if lookup.standardDefinitions.has(option.GetSpace(), option.GetCode()) {
		return true
	}

	lookup.customDefinitionMutex.RLock()
	definitions, ok := lookup.customDefinitions[daemonID]
	lookup.customDefinitionMutex.RUnlock()

	if ok {
		return definitions.has(option.GetSpace(), option.GetCode())
	}
	return false
}

// Finds option definition for the specified option. Internally, it queries standard
// Kea option definitions defined in the keaconfig package. In the future it will also
// be able to search for the runtime definitions in the database.
func (lookup DHCPOptionDefinitionLookup) Find(daemonID int64, option dhcpmodel.DHCPOptionAccessor) keaconfig.DHCPOptionDefinitionAccessor {
	// Check if the option is a standard one.
	if definition := lookup.standardDefinitions.find(option.GetSpace(), option.GetCode()); definition != nil {
		return definition
	}

	lookup.customDefinitionMutex.RLock()
	definitions, ok := lookup.customDefinitions[daemonID]
	lookup.customDefinitionMutex.RUnlock()

	// Check the custom definitions.
	if ok {
		return definitions.find(option.GetSpace(), option.GetCode())
	}

	return nil
}

// Sets the custom option definitions for the specified daemon.
func (lookup *DHCPOptionDefinitionLookup) SetDefinitions(daemonID int64, definitions []keaconfig.DHCPOptionDefinition) {
	// Lock the mutex to prevent concurrent access to the custom definitions.
	lookup.customDefinitionMutex.Lock()
	defer lookup.customDefinitionMutex.Unlock()

	// Clear the existing definitions.
	lookup.customDefinitions[daemonID] = make(definitionStore)

	// Add the new definitions.
	for _, definition := range definitions {
		if _, ok := lookup.customDefinitions[daemonID][definition.Space]; !ok {
			lookup.customDefinitions[daemonID][definition.Space] = make(map[uint16]keaconfig.DHCPOptionDefinition)
		}
		lookup.customDefinitions[daemonID][definition.Space][definition.Code] = definition
	}
}
