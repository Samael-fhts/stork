package dbmodel

import (
	keaconfig "isc.org/stork/appcfg/kea"
	dbops "isc.org/stork/server/database"
)

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
