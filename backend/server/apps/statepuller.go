package apps

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	keaconfig "isc.org/stork/appcfg/kea"
	"isc.org/stork/server/agentcomm"
	"isc.org/stork/server/apps/bind9"
	"isc.org/stork/server/apps/kea"
	"isc.org/stork/server/apps/pdns"
	"isc.org/stork/server/configreview"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	"isc.org/stork/server/eventcenter"
	storkutil "isc.org/stork/util"
)

// Instance of the puller which periodically checks the status of the Kea apps.
// Besides basic status information the High Availability status is fetched.
type StatePuller struct {
	*agentcomm.PeriodicPuller
	EventCenter                eventcenter.EventCenter
	ReviewDispatcher           configreview.Dispatcher
	DHCPOptionDefinitionLookup keaconfig.DHCPOptionDefinitionLookup
}

// Create an instance of the puller which periodically checks the status of
// the Kea daemons.
func NewStatePuller(db *dbops.PgDB, agents agentcomm.ConnectedAgents, eventCenter eventcenter.EventCenter, reviewDispatcher configreview.Dispatcher, lookup keaconfig.DHCPOptionDefinitionLookup) (*StatePuller, error) {
	puller := &StatePuller{
		EventCenter:                eventCenter,
		ReviewDispatcher:           reviewDispatcher,
		DHCPOptionDefinitionLookup: lookup,
	}
	periodicPuller, err := agentcomm.NewPeriodicPuller(db, agents, "State Puller",
		"state_puller_interval", puller.pullData)
	if err != nil {
		return nil, err
	}
	puller.PeriodicPuller = periodicPuller
	return puller, nil
}

// Stops the timer triggering status checks.
func (puller *StatePuller) Shutdown() {
	puller.PeriodicPuller.Shutdown()
}

// Gets the status of machines and their daemons and stores useful information in the database.
func (puller *StatePuller) pullData() error {
	// get list of all authorized machines from database
	authorized := true
	dbMachines, err := dbmodel.GetAllMachines(puller.DB, &authorized)
	if err != nil {
		return err
	}

	// get state from machines and their daemons
	var lastErr error
	okCnt := 0
	for _, dbM := range dbMachines {
		dbM2 := dbM
		ctx := context.Background()
		errStr := UpdateMachineAndDaemonsState(ctx, puller.DB, &dbM2, puller.Agents, puller.EventCenter, puller.ReviewDispatcher, puller.DHCPOptionDefinitionLookup)
		if errStr != "" {
			lastErr = errors.New(errStr)
			log.Errorf("Error occurred while getting info from machine %d: %s", dbM2.ID, errStr)
		} else {
			okCnt++
		}
	}
	log.Printf("Completed pulling information from machines: %d/%d succeeded", okCnt, len(dbMachines))
	return lastErr
}

// Store updated machine fields in to database.
func updateMachineFields(db *dbops.PgDB, dbMachine *dbmodel.Machine, m *agentcomm.State) error {
	// update state fields in machine
	dbMachine.State.AgentVersion = m.AgentVersion
	dbMachine.State.Cpus = m.Cpus
	dbMachine.State.CpusLoad = m.CpusLoad
	dbMachine.State.Memory = m.Memory
	dbMachine.State.Hostname = m.Hostname
	dbMachine.State.Uptime = m.Uptime
	dbMachine.State.UsedMemory = m.UsedMemory
	dbMachine.State.Os = m.Os
	dbMachine.State.Platform = m.Platform
	dbMachine.State.PlatformFamily = m.PlatformFamily
	dbMachine.State.PlatformVersion = m.PlatformVersion
	dbMachine.State.KernelVersion = m.KernelVersion
	dbMachine.State.KernelArch = m.KernelArch
	dbMachine.State.VirtualizationSystem = m.VirtualizationSystem
	dbMachine.State.VirtualizationRole = m.VirtualizationRole
	dbMachine.State.HostID = m.HostID
	dbMachine.LastVisitedAt = m.LastVisitedAt
	dbMachine.Error = m.Error
	err := dbmodel.UpdateMachine(db, dbMachine)
	if err != nil {
		return errors.Wrapf(err, "problem updating machine %+v", dbMachine)
	}
	return nil
}

// daemonCompare compares two daemons for equality. Two daemons are considered
// equal if their type matches and if they have the same control port. Return
// true if equal, false otherwise.
func daemonCompare(dbDaemon dbmodel.Daemon, grpcDaemon *agentcomm.Daemon) bool {
	if dbDaemon.Name != dbmodel.DaemonName(grpcDaemon.Name) {
		return false
	}
	if len(dbDaemon.AccessPoints) != len(grpcDaemon.AccessPoints) {
		return false
	}
	accessPointIndex := map[string]*dbmodel.AccessPoint{}
	for _, pt := range dbDaemon.AccessPoints {
		accessPointIndex[pt.Type] = pt
	}

	for _, grpcPt := range grpcDaemon.AccessPoints {
		dbPt, ok := accessPointIndex[grpcPt.Type]
		if !ok {
			return false
		}

		if dbPt.Port != grpcPt.Port || dbPt.Address != grpcPt.Address || dbPt.Key != grpcPt.Key || dbPt.UseSecureProtocol != grpcPt.UseSecureProtocol {
			return false
		}
	}

	return true
}

// For each provided discovered daemon, try to find a matching daemon in the
// database. If it is found, use it, otherwise create a new daemon.
func mergeNewAndOldDaemons(db *dbops.PgDB, dbMachine *dbmodel.Machine, discoveredDaemons []*agentcomm.Daemon) (existing, deleted []dbmodel.Daemon, err error) {
	oldDaemons, err := dbmodel.GetDaemonsByMachine(db, dbMachine.ID)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "cannot get machine's daemons from db")
	}

	matchedDaemons := make([]dbmodel.Daemon, 0, len(discoveredDaemons))

DISCOVERED_LOOP:
	for _, discoveredDaemon := range discoveredDaemons {
		for _, oldDaemon := range oldDaemons {
			if daemonCompare(oldDaemon, discoveredDaemon) {
				matchedDaemons = append(matchedDaemons, oldDaemon)
				oldDaemons = append(oldDaemons[:0], oldDaemons[1:]...) // remove matched daemon
				continue DISCOVERED_LOOP
			}
		}

		// The daemon was not found in the old daemons, so create a new one.
		accessPoints := make([]*dbmodel.AccessPoint, len(discoveredDaemon.AccessPoints))
		for i, point := range discoveredDaemon.AccessPoints {
			accessPoints[i] = &dbmodel.AccessPoint{
				Type:              point.Type,
				Address:           point.Address,
				Port:              point.Port,
				Key:               point.Key,
				UseSecureProtocol: point.UseSecureProtocol,
			}
		}

		newDaemon := dbmodel.Daemon{
			MachineID:    dbMachine.ID,
			Machine:      dbMachine,
			Name:         dbmodel.DaemonName(discoveredDaemon.Name),
			AccessPoints: accessPoints,
		}
		matchedDaemons = append(matchedDaemons, newDaemon)
	}

	return matchedDaemons, oldDaemons, nil
}

// Retrieve remotely machine and its daemons state, and store it in the database.
func UpdateMachineAndDaemonsState(ctx context.Context, db *dbops.PgDB, dbMachine *dbmodel.Machine, agents agentcomm.ConnectedAgents, eventCenter eventcenter.EventCenter, reviewDispatcher configreview.Dispatcher, lookup keaconfig.DHCPOptionDefinitionLookup) string {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// get state of machine from agent
	state, err := agents.GetState(ctx2, dbMachine)
	if err != nil {
		log.WithError(err).Warn("Cannot get state of machine")
		dbMachine.Error = "Cannot get state of machine"
		err = dbmodel.UpdateMachine(db, dbMachine)
		if err != nil {
			log.Error(err)
			return "Problem updating record in database"
		}
		return ""
	}

	agentVersion, err := storkutil.ParseSemanticVersion(state.AgentVersion)
	if err != nil {
		log.WithError(err).Error("Cannot parse agent version: %s", state.AgentVersion)
		return "Cannot parse agent version"
	}

	if agentVersion.LessThanOrEqual(storkutil.NewSemanticVersion(2, 3, 0)) {
		// The agent communicates through the Kea CA. It cannot detect the
		// other daemons.
		var additionalDaemons []*agentcomm.Daemon

		for _, daemon := range state.Daemons {
			// The old agent used app type instead of daemon name. The app type
			// for Kea was "kea" that means that the Kea CA daemon has been
			// detected.
			if daemon.Name != agentcomm.DaemonNameKea {
				continue
			}
			// Convert the daemon name to the proper one.
			daemon.Name = agentcomm.DaemonNameCA

			// Fetch the Kea CA configuration to retrieve a list of running
			// daemons.
			config, _, err := kea.GetConfig(ctx, agents, daemon)
			if err != nil {
				return "Cannot get Kea CA configuration: " + err.Error()
			}

			daemonNames := config.ControlSockets.GetConfiguredDaemonNames()
			for _, name := range daemonNames {
				if name == string(agentcomm.DaemonNameCA) {
					continue
				}

				additionalDaemons = append(additionalDaemons, &agentcomm.Daemon{
					Name: agentcomm.DaemonName(name),
					// Communication with this daemon is done through the Kea CA.
					AccessPoints: daemon.AccessPoints,
					Machine:      daemon.Machine,
				})
			}
		}

		// Append the additional daemons to the list of daemons.
		state.Daemons = append(state.Daemons, additionalDaemons...)
	}

	// The Stork server doesn't gather the Stork agent configuration, so we cannot
	// detect its change. It used to compare the current agent state and the database
	// entry to merely recognize the HTTP credentials state change but this
	// parameter has been removed from the agent state. The following variable is
	// a placeholder for the possible future implementation of the Stork agent
	// configuration change detection.
	isStorkAgentChanged := false

	// store machine's state in db
	err = updateMachineFields(db, dbMachine, state)
	if err != nil {
		msg := "Cannot update machine in db"
		log.WithError(err).Error(msg)
		return msg
	}

	// take old daemons from db and new daemons fetched from the machine
	// and match them and prepare a list of all daemons
	existingDaemons, deletedDaemons, err := mergeNewAndOldDaemons(db, dbMachine, state.Daemons)
	if err != nil {
		return "Cannot merge new and old daemons: " + err.Error()
	}

	// remove daemons that no longer exist on the machine
	for _, dbDaemon := range deletedDaemons {
		log.Infof("Removing daemon %s from machine %s", dbDaemon.Name, dbMachine.Name)
		err = dbmodel.DeleteDaemon(db, &dbDaemon)
		if err != nil {
			log.WithError(err).Errorf("Cannot delete daemon %s from machine %s", dbDaemon.Name, dbMachine.Name)
			return "Cannot delete daemon from database"
		}
	}

	// go through all daemons and store their changes in database
	for _, dbDaemon := range allDaemons {
		// get daemon state from the machine
		switch dbDaemon.Name {
		case dbmodel.DaemonNameDHCPv4, dbmodel.DaemonNameDHCPv6, dbmodel.DaemonNameCA, dbmodel.DaemonNameD2:
			state := kea.GetDaemonState(ctx2, agents, dbDaemon, eventCenter)
			err = kea.CommitDaemonIntoDB(db, dbDaemon, eventCenter, state, lookup)
			if err == nil {
				// Let's now identify new daemons or the daemons with updated
				// configurations and schedule configuration reviews for them
				conditionallyBeginKeaConfigReviews(dbDaemon, state, reviewDispatcher, isStorkAgentChanged)
			}
		case dbmodel.DaemonNameBind9:
			bind9.GetDaemonState(ctx2, agents, dbDaemon, eventCenter)
			err = bind9.CommitDaemonIntoDB(db, dbDaemon, eventCenter)
		case dbmodel.DaemonNamePDNS:
			pdns.GetDaemonState(ctx2, agents, dbDaemon, eventCenter)
			err = pdns.CommitDaemonIntoDB(db, dbDaemon, eventCenter)
		default:
			err = nil
		}

		if err != nil {
			log.WithError(err).Errorf("Cannot store daemon state")
			return "Problem storing daemon state in the database"
		}
	}

	// add all daemons to machine's daemons list - it will be used in ReST API functions
	// to return state of machine and its daemons
	dbMachine.Daemons = allDaemons

	return ""
}

// This function iterates over the daemon and checks if a new config
// review should be performed. It is performed when daemon's configuration
// or dispatcher's signature has changed.
func conditionallyBeginKeaConfigReviews(daemon *dbmodel.Daemon, state *kea.AppStateMeta, reviewDispatcher configreview.Dispatcher, storkAgentConfigChanged bool) {
	// Let's make sure that the config pointer is set. It can be nil
	// when the daemon is inactive.
	if daemon.KeaDaemon == nil || daemon.KeaDaemon.Config == nil {
		return
	}

	var triggers configreview.Triggers
	if storkAgentConfigChanged {
		triggers = append(triggers, configreview.StorkAgentConfigModified)
	}

	isConfigModified := true
	if state != nil && state.SameConfigDaemons != nil {
		if isSame, ok := state.SameConfigDaemons[daemon.Name]; ok && isSame {
			if daemon.ConfigReview != nil &&
				daemon.ConfigReview.Signature == reviewDispatcher.GetSignature() {
				// Configuration of this daemon hasn't changed and the dispatcher has
				// no checkers modified since the last review. Skip the config modified trigger.
				isConfigModified = false
			}
		}
	}
	if isConfigModified {
		triggers = append(triggers, configreview.ConfigModified)
	}

	if len(triggers) != 0 {
		_ = reviewDispatcher.BeginReview(daemon, triggers, nil)
	}
}
