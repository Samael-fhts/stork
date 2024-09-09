package keaconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

// Embeds the standard DHCPv4 option definitions.
//
//go:embed stdoptiondef4.json
var stdDHCPv4OptionDefsJSON []byte

// Embeds the standard DHCPv6 option definitions.
//
//go:embed stdoptiondef6.json
var stdDHCPv6OptionDefsJSON []byte

// Parses the embedded JSON data with standard DHCP option definitions.
func parseDHCPOptionDefinitions(jsonData []byte) []DHCPOptionDefinition {
	var definitions []DHCPOptionDefinition
	err := json.Unmarshal(jsonData, &definitions)
	if err != nil {
		// The embedded JSON data should be correct, so this is a programming error.
		panic(fmt.Sprintf("failed to parse standard DHCP option definitions: %v", err))
	}
	return definitions
}

// Returns the standard DHCPv4 option definitions.
//
//nolint:gochecknoglobals
var GetStandardDHCPv4OptionDefinitions = sync.OnceValue(func() []DHCPOptionDefinition {
	return parseDHCPOptionDefinitions(stdDHCPv4OptionDefsJSON)
})

// Returns the standard DHCPv6 option definitions.
//
//nolint:gochecknoglobals
var GetStandardDHCPv6OptionDefinitions = sync.OnceValue(func() []DHCPOptionDefinition {
	return parseDHCPOptionDefinitions(stdDHCPv6OptionDefsJSON)
})
