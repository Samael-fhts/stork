package dbmodel

import (
	keadata "isc.org/stork/daemondata/kea"
)

// Extends basic Lease information with database specific information.
type Lease struct {
	ID int64

	keadata.Lease

	DaemonId int64
	Daemon   *Daemon
}
