package dbmodel

import (
	keaconfig "isc.org/stork/appcfg/kea"
	dbops "isc.org/stork/server/database"
)

// A factory for creating DHCPOptionDefinitionLookup instances for daemons.
// It caches the created objects. The cache is not thread-safe.
// It allows to significantly reduce the number of queries to the database if
// multiple entities referenced the same daemon are processed in a loop.
type DHCPOptionDefinitionLookups struct {
	db    dbops.DBI
	cache map[int64]keaconfig.DHCPOptionDefinitionLookup
}

// Creates new lookup instance.
func NewDHCPOptionDefinitionLookups(db dbops.DBI) *DHCPOptionDefinitionLookups {
	return &DHCPOptionDefinitionLookups{
		db:    db,
		cache: make(map[int64]keaconfig.DHCPOptionDefinitionLookup),
	}
}

// Returns lookup for the specified daemon.
func (c *DHCPOptionDefinitionLookups) GetLookup(daemonID int64) (keaconfig.DHCPOptionDefinitionLookup, error) {
	if lookup, ok := c.cache[daemonID]; ok {
		return lookup, nil
	}

	// ToDo: Read only the definitions, not the whole config.
	daemon, err := GetDaemonByID(c.db, daemonID)
	if err != nil {
		return nil, err
	}

	if daemon.KeaDaemon.Config == nil {
		// The config is not fetched yet.
		return nil, nil
	}

	lookup := keaconfig.NewDHCPOptionDefinitionLookup(
		daemon.KeaDaemon.Config.GetDHCPOptionDefinitions(),
	)
	c.cache[daemonID] = lookup
	return lookup, nil
}
