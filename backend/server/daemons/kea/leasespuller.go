package kea

import (
	"context"
	"iter"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	// "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	// keaconfig "isc.org/stork/daemoncfg/kea"
	// keactrl "isc.org/stork/daemonctrl/kea"
	// "isc.org/stork/datamodel/daemonname"
	"isc.org/stork/server/agentcomm"
	storkutil "isc.org/stork/util"

	// dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	// storkutil "isc.org/stork/util"
)

// Leases puller is responsible for fetching lease data from Kea via the agents.
type LeasesPuller struct {
	*agentcomm.PeriodicPuller
	db *pg.DB
}

// Create a LeasesPuller object that, in the background, pulls lease records
// from Kea. The retreived records are added to the database.
func NewLeasesPuller(db *pg.DB, agents agentcomm.ConnectedAgents) (*LeasesPuller, error) {
	puller := LeasesPuller{
		nil,
		db,
	}
	return &puller, nil
}

// Shut down the LeasesPuller. It ends the goroutine that pulls lease records.
func (puller *LeasesPuller) Shutdown() {
	puller.PeriodicPuller.Shutdown()
}

// A unique key for identifying Keas that are talking to the same database or writing to the same leasefile.  In order to use this effectively, EITHER:
//   - set `machine` and `leasefilePath`, leaving all other fields at the zero
//     value
//   - set `dbHost`, `dbPort`, and `dbName`, leaving all other fields at the zero
//     value.
//   - set `unique` to a non-zero unique value
//
// If two daemons are looking at the same leasefilePath on the same machine,
// they are using the same lease database. If two daemons are looking at the
// same database host, port, and database name, then they are using the same
// lease database. Changing any one of those values (very likely) means that
// they are pointed at different databases. `unique` is provided as an escape
// hatch to deal gracefully with uncommon Kea configurations (persist=false, any
// future enhancement to add a new lease database type).
//
// Known edge cases where this fails:
//   - Using nftables/iptables to redirect two external ports to the same RDBMS
//     (incorrectly sees them as different)
//   - Using CNAMEs to point two hosts at the same IP (incorrectly sees them as
//     different)
//   - Using multiple NICs or multicast tricks to point two IPs at the same
//     computer (incorrectly sees them as different)
//   - Mounting a filesystem on multiple machines using SMB or NFS (incorrectly
//     sees them as different)
//   - Running an agent on a host and having it monitor multiple Keas writing
//     leasefiles in containers (incorrectly sees them as the same)
type leaseDBUniqueKey struct {
	machine       int64
	leasefilePath string
	dbHost        string
	dbPort        int
	dbName        string
	unique        int
}

// Pull leases from all monitored Kea daemons and store them in the database.
func (puller *LeasesPuller) pullLeases() error {
	// Get a list of all of the DHCP daemons from the database. This function
	// must continue to return the daemons in a deterministic order, otherwise
	// the following code might break.
	daemons, err := dbmodel.GetDHCPDaemons(puller.db)
	if err != nil {
		return err
	}

	// Take only the first daemon which uses each unique database.
	uniqueCtr := 1
	selectedDaemons := map[leaseDBUniqueKey]*dbmodel.Daemon{}
	for _, daemon := range daemons {
		if daemon.KeaDaemon == nil || daemon.KeaDaemon.Config == nil {
			log.WithField("daemon_id", daemon.ID).Debug("Ignoring non-Kea or configless daemon")
			continue
		}
		databases := daemon.KeaDaemon.Config.GetAllDatabases()
		if databases.Lease == nil {
			log.WithField("daemon_id", daemon.ID).Debug("Ignoring daemon with no lease database")
			continue
		}
		key := leaseDBUniqueKey{}
		if databases.Lease.Type == "memfile" && databases.Lease.Persist != nil && *databases.Lease.Persist {
			// TODO: get full memfile path from status-get response
			key.leasefilePath = databases.Lease.Path
			key.machine = daemon.MachineID
		} else if databases.Lease.Type == "mysql" || databases.Lease.Type == "postgresql" {
			key.dbHost = databases.Lease.Host
			key.dbName = databases.Lease.Name
			// TODO: add this to the parsed part of the config.
			// key.dbPort = databases.Lease.Port
		} else {
			key.unique = uniqueCtr
			uniqueCtr++
		}
		_, ok := selectedDaemons[key]
		if ok {
			// There's already a daemon matching this key in the map, so don't
			// overwrite it.
			continue
		}
		selectedDaemons[key] = &daemon
	}

	// Get lease records from each daemon.
	var errors []error
	daemonsOkCnt := 0
	for _, daemon := range selectedDaemons {
		err := puller.getLeasesFromDaemon(daemon)
		if err != nil {
			errors = append(errors, err)
			log.WithError(err).Warnf("Could not retreive leases from daemon %d", daemon.ID)
		} else {
			daemonsOkCnt++
		}
	}
	return storkutil.CombineErrors("errors occurred while trying to pull leases from one or more daemons", errors)
}

func (puller *LeasesPuller) getLeasesFromDaemon(daemon *dbmodel.Daemon) error {
	if daemon.KeaDaemon == nil || !daemon.Active {
		log.WithField("daemon_id", daemon.ID).Debug("Skipping daemon because it is not Kea or it is inactive")
		return nil
	}

	if !daemon.Name.IsDHCP() {
		log.WithField("daemon_id", daemon.ID).Debug("Skipping daemon because it is not a DHCPD")
		return nil
	}

	ctx := context.Background()
	// TODO: retreive last seen CLTT
	for response, err := range puller.Agents.ReceiveKeaLeases(ctx, daemon, 0) {
		switch {
		case err != nil:
			return err
		case response == nil:
			return errors.New("unexpected nil in response stream of Kea leases")
		// Everything worked; happy path.
		default:
			// TODO: collect leases
		}
	}
	return nil
}
