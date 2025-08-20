package kea

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	keaconfig "isc.org/stork/appcfg/kea"
	keactrl "isc.org/stork/appctrl/kea"
	"isc.org/stork/server/agentcomm"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	"isc.org/stork/server/eventcenter"
	storkutil "isc.org/stork/util"
)

// Get list of hooks for the given Kea daemon.
func GetDaemonHooks(dbDaemon *dbmodel.Daemon) (hooks []string) {
	if dbDaemon.KeaDaemon == nil || dbDaemon.KeaDaemon.Config == nil {
		return
	}
	libraries := dbDaemon.KeaDaemon.Config.GetHookLibraries()
	for _, library := range libraries {
		hooks = append(hooks, library.Library)
	}
	return
}

// Get list of log targets for the given Kea daemon.
func GetDaemonLogTargets(dbDaemon *dbmodel.Daemon) (logTargets []dbmodel.LogTarget) {
	if dbDaemon.KeaDaemon == nil || dbDaemon.KeaDaemon.Config == nil {
		return
	}
	for _, logTarget := range dbDaemon.LogTargets {
		logTargets = append(logTargets, *logTarget)
	}
	return
}

// The arguments of the version-get command response.
type VersionGetRespArgs struct {
	Extended string
}

// The response of the version-get command.
type VersionGetResponse struct {
	keactrl.ResponseHeader
	Arguments *VersionGetRespArgs `json:"arguments,omitempty"`
}

// Struct returned by GetAppState() function.
type AppStateMeta struct {
	Events            []*dbmodel.Event
	SameConfigDaemons map[string]bool
}

// Convenience function called from getStateFromCA and getStateFromDaemons which searches
// for the existing daemon within a machine. If the daemon does not exist a new instance is
// created. Otherwise, the function returns a shallow copy of the Daemon and KeaDaemon
// and sets the active flag to true.
func copyOrCreateActiveKeaDaemon(machine *dbmodel.Machine, daemonName dbmodel.DaemonName) *dbmodel.Daemon {
	daemon := machine.GetDaemonByName(daemonName)
	if daemon != nil {
		daemonCopy := dbmodel.ShallowCopyKeaDaemon(daemon)
		daemonCopy.Active = true
		return daemonCopy
	}
	return dbmodel.NewKeaDaemon(daemonName, true)
}

// Get state of Kea Control Agent using ForwardToKeaOverHTTP function.
// The state, that is stored into dbMachine, includes: version and config of CA.
// It also returns a list of all Kea daemons
func getStateFromCA(ctx context.Context, agents agentcomm.ConnectedAgents, dbMachine *dbmodel.Machine, daemonsMap map[string]*dbmodel.Daemon, daemonsErrors map[string]string) ([]dbmodel.DaemonName, error) {
	// prepare the command to get config and version from CA
	cmds := []keactrl.SerializableCommand{
		keactrl.NewCommandBase(keactrl.VersionGet, keactrl.CA),
		keactrl.NewCommandBase(keactrl.ConfigGet, keactrl.CA),
	}

	// get version and config from CA
	versionGetResp := []VersionGetResponse{}
	caConfigGetResp := []keactrl.HashedResponse{}

	dbDaemon := dbMachine.GetDaemonByName(dbmodel.DaemonNameCA)
	if dbDaemon == nil {
		return nil, errors.Errorf("machine %d has no Kea Control Agent daemon", dbMachine.ID)
	}

	cmdsResult, err := agents.ForwardToKeaOverHTTP(ctx, dbDaemon, cmds, &versionGetResp, &caConfigGetResp)
	if err != nil {
		return nil, err
	}
	if cmdsResult.Error != nil {
		return nil, cmdsResult.Error
	}

	daemonsMap[dbmodel.DaemonNameCA] = copyOrCreateActiveKeaDaemon(dbMachine, dbmodel.DaemonNameCA)

	// if no error in the version-get response then copy retrieved info about CA to its record
	dmn := daemonsMap[dbmodel.DaemonNameCA]
	err = cmdsResult.CmdsErrors[0]

	switch {
	case err != nil:
		// Use the error as-is.
	case len(versionGetResp) == 0:
		err = errors.Errorf("empty response")
	default:
		err = versionGetResp[0].GetError()
	}

	if err != nil {
		dmn.Active = false
		err = errors.WithMessage(err, "problem with version-get response from CA")
		log.WithError(err).Warn("Problem with version-get response from CA")
		daemonsErrors[dbmodel.DaemonNameCA] = err.Error()
		return nil, err
	}

	dmn.Version = versionGetResp[0].Text
	if versionGetResp[0].Arguments != nil {
		dmn.ExtendedVersion = versionGetResp[0].Arguments.Extended
	}

	// if no error in the config-get response then copy retrieved info about available daemons
	if len(caConfigGetResp) == 0 {
		err = errors.Errorf("empty response")
	} else if err = caConfigGetResp[0].GetError(); err != nil {
		// Use the response error.
	} else if caConfigGetResp[0].Arguments == nil {
		err = errors.Errorf("empty arguments")
	}

	if err != nil {
		dmn.Active = false
		err = errors.WithMessage(err, "problem with config-get response from CA")
		log.WithError(err).Warn("Problem with config-get response from CA")
		daemonsErrors[dbmodel.DaemonNameCA] = err.Error()
		return nil, err
	}

	// prepare a set of available daemons
	allDaemons := []string{}

	// Only set the new configuration if the configuration is added for the first
	// time or the hash values aren't matching.
	if (dmn.KeaDaemon.Config == nil) || (dmn.KeaDaemon.ConfigHash != caConfigGetResp[0].ArgumentsHash) {
		// Set the configuration for the daemon and populate selected configuration
		// information to the respective structures, e.g. logging information.
		err = dmn.SetConfigWithHash(dbmodel.NewKeaConfig(caConfigGetResp[0].Arguments),
			caConfigGetResp[0].ArgumentsHash)
		if err != nil {
			return nil, err
		}
	}

	sockets := dmn.KeaDaemon.Config.GetControlSockets()
	if sockets == nil {
		return allDaemons, nil
	}

	if sockets.Dhcp4 != nil {
		allDaemons = append(allDaemons, dbmodel.DaemonNameDHCPv4)
	}
	if sockets.Dhcp6 != nil {
		allDaemons = append(allDaemons, dbmodel.DaemonNameDHCPv6)
	}
	if sockets.D2 != nil {
		allDaemons = append(allDaemons, dbmodel.DaemonNameD2)
	}

	return allDaemons, nil
}

// Get state of Kea machine daemons (beside Control Agent) using ForwardToKeaOverHTTP function.
// The state, that is stored into dbMachine, includes: version, config and runtime state of indicated Kea daemons.
func getStateFromDaemons(ctx context.Context, agents agentcomm.ConnectedAgents, dbMachine *dbmodel.Machine, daemonsMap map[string]*dbmodel.Daemon, allDaemons []dbmodel.DaemonName, daemonsErrors map[string]string) error {
	now := storkutil.UTCNow()

	for _, daemonName := range allDaemons {
		isDHCPDaemon := daemonName == dbmodel.DaemonNameDHCPv4 || daemonName == dbmodel.DaemonNameDHCPv6

		versionGetResponseItems := []VersionGetResponse{}
		configGetResponseItems := []keactrl.HashedResponse{}
		var statusGetResponseItems []StatusGetResponse

		cmds := []keactrl.SerializableCommand{
			keactrl.NewCommandBase(keactrl.VersionGet, daemonName),
			keactrl.NewCommandBase(keactrl.ConfigGet, daemonName),
		}
		responses := []any{&versionGetResponseItems, &configGetResponseItems}

		if isDHCPDaemon {
			cmds = append(cmds, keactrl.NewCommandBase(keactrl.StatusGet, daemonName))
			statusGetResponseItems = []StatusGetResponse{}
			responses = append(responses, &statusGetResponseItems)
		}

		daemon := dbMachine.GetDaemonByName(daemonName)
		if daemon == nil {
			return errors.Errorf("machine %d has no Kea daemon %s", dbMachine.ID, daemonName)
		}
		daemon = dbmodel.ShallowCopyKeaDaemon(daemon)

		cmdsResult, err := agents.ForwardToKeaOverHTTP(ctx, daemon, cmds, responses...)
		if err != nil {
			return err
		}
		if cmdsResult.Error != nil {
			return cmdsResult.Error
		}

		// first find old records of daemons in old daemons assigned to the app
		daemonsMap[daemonName] = daemon

		// process version-get responses
		err = cmdsResult.CmdsErrors[0]
		if err != nil {
			return errors.WithMessage(err, "problem with version-get response")
		}

		if len(versionGetResponseItems) != 0 {
			return errors.Errorf("unexpected number of version-get response items: %d", len(versionGetResponseItems))
		}
		versionGetResponse := versionGetResponseItems[0]
		if err := versionGetResponse.GetError(); err != nil {
			daemon.Active = false
			err = errors.WithMessage(err, "problem with version-get response")
			log.WithError(err).Warn("Problem with version-get response")
			daemonsErrors[daemon.Name] = err.Error()
			continue
		}
		daemon.Version = versionGetResponse.Text
		if versionGetResponse.Arguments != nil {
			daemon.ExtendedVersion = versionGetResponse.Arguments.Extended
		}

		// process config-get responses
		err = cmdsResult.CmdsErrors[1]
		if err != nil {
			return errors.WithMessage(err, "problem with config-get response")
		}

		if len(configGetResponseItems) != 1 {
			return errors.Errorf("unexpected number of config-get response items: %d", len(configGetResponseItems))
		}
		configGetResponse := configGetResponseItems[0]

		if err := configGetResponse.GetError(); err != nil {
			daemon.Active = false
			err = errors.WithMessage(err, "problem with config-get and kea daemon")
			log.WithError(err).Warn("Problem with config-get and kea daemon")
			daemonsErrors[daemon.Name] = err.Error()
			continue
		}

		if (daemon.KeaDaemon.Config == nil) || (daemon.KeaDaemon.ConfigHash != configGetResponse.ArgumentsHash) {
			// Set the configuration for the daemon and populate selected configuration
			// information to the respective structures, e.g. logging information.
			err = daemon.SetConfigWithHash(dbmodel.NewKeaConfig(configGetResponse.Arguments), configGetResponse.ArgumentsHash)
			if err != nil {
				errStr := fmt.Sprintf("%s", err)
				log.Warn(errStr)
				daemonsErrors[daemon.Name] = errStr
				continue
			}
		}

		if isDHCPDaemon {
			// process status-get responses
			err = cmdsResult.CmdsErrors[2]
			if err != nil {
				return errors.WithMessage(err, "problem with status-get response")
			}

			if len(statusGetResponseItems) != 1 {
				return errors.Errorf("unexpected number of status-get response items: %d", len(statusGetResponseItems))
			}
			statusGetResponse := statusGetResponseItems[0]

			if err := statusGetResponse.GetError(); err != nil {
				daemon.Active = false
				err = errors.WithMessage(err, "problem with status-get and kea daemon")
				log.WithError(err).Warn("Problem with status-get and kea daemon")
				daemonsErrors[daemon.Name] = err.Error()
				continue
			}

			if statusGetResponse.Arguments != nil {
				daemon.Uptime = statusGetResponse.Arguments.Uptime
				daemon.ReloadedAt = now.Add(time.Second * time.Duration(-statusGetResponse.Arguments.Reload))
			}
		}
	}

	return nil
}

// Get state of Kea application daemons using ForwardToKeaOverHTTP function.
// The state that is stored into dbApp includes: version, config and runtime state of indicated Kea daemons.
func GetAppState(ctx context.Context, agents agentcomm.ConnectedAgents, dbApp *dbmodel.App, eventCenter eventcenter.EventCenter) *AppStateMeta {
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// get state from CA
	daemonsMap := map[string]*dbmodel.Daemon{}
	daemonsErrors := map[string]string{}
	allDaemons, dhcpDaemons, err := getStateFromCA(ctx2, agents, dbApp, daemonsMap, daemonsErrors)
	if err != nil {
		log.Warnf("Problem getting state from Kea CA: %s", err)
	}

	// if no problems then now get state from the rest of Kea daemons
	err = getStateFromDaemons(ctx2, agents, dbApp, daemonsMap, allDaemons, dhcpDaemons, daemonsErrors)
	if err != nil {
		log.Warnf("Problem getting state from Kea daemons: %s", err)
	}

	// If this is new app let's set its active/inactive state based on the
	// active/inactive state of its daemons. Also, convert the map to the
	// list of daemons.
	if dbApp.ID == 0 {
		dbApp.Active, dbApp.Daemons = createNewAppState(daemonsMap)
		return nil
	}

	newActive, overrideDaemons, newDaemons, events, sameConfigDaemons := findChangesAndRaiseEvents(dbApp, daemonsMap, daemonsErrors)

	// update app state
	dbApp.Active = newActive
	if overrideDaemons {
		dbApp.Daemons = newDaemons
	}

	// Return supplementary information about the state returned.
	state := &AppStateMeta{
		Events:            events,
		SameConfigDaemons: sameConfigDaemons,
	}

	return state
}

// Determines whether the new app is active or inactive based on the
// active/inactive state of its daemons. It returns a boolean flag
// indicating whether the app is active or not and the list of
// daemons to be assigned to the app. This function is called by the
// GetAppState function when adding a new app.
func createNewAppState(daemonsMap map[string]*dbmodel.Daemon) (active bool, daemons []*dbmodel.Daemon) {
	for name := range daemonsMap {
		daemon := daemonsMap[name]
		// If all daemons are active then whole app is active.
		active = active && daemon.Active

		// If this is new daemon and it is not active then disable its monitoring.
		if !daemon.Active {
			daemon.Monitored = false
		}
		daemons = append(daemons, daemon)
	}

	return active, daemons
}

// Detects changes in the returned app state comparing to the state recorded in the
// database. It raises events when a daemon changes its state between active and
// inactive state. It also raises events about detected daemon restarts and when
// configuration change was detected. This function should only be called from
// the GetAppState function. The following values are returned: boolean value
// indicating whether the app is considered active or inactive after update;
// a boolean flag indicating whether daemons in the app should be replaced with
// daemons returned in 3rd argument; list of events to be passed to the event
// center; map of names of daemons for which configuration remains the same.
func findChangesAndRaiseEvents(dbApp *dbmodel.App, daemonsMap map[string]*dbmodel.Daemon, daemonsErrors map[string]string) (bool, bool, []*dbmodel.Daemon, []*dbmodel.Event, map[string]bool) {
	var (
		newDaemons []*dbmodel.Daemon
		events     []*dbmodel.Event
	)

	newCADaemon, ok := daemonsMap["ca"]
	if !ok || !newCADaemon.Active {
		// Kea Control Agent was not found in the response or it is inactive.
		for _, oldDaemon := range dbApp.Daemons {
			// For all active daemons we need to mark them as inactive and raise events
			// about the daemons being unreachable.
			if oldDaemon.Active {
				oldDaemon.Active = false

				// Add a pointer to the app in the daemon because it will be needed
				// when creating the event below.
				oldDaemon.App = dbApp
				errStr := daemonsErrors[oldDaemon.Name]
				ev := eventcenter.CreateEvent(dbmodel.EvError, "{daemon} is unreachable", errStr, dbApp.Machine, dbApp, oldDaemon)
				events = append(events, ev)
			}
		}
		// In addition, raise an event indicating that the whole app is unreachable.
		if dbApp.Active {
			ev := eventcenter.CreateEvent(dbmodel.EvError, "{app} is unreachable", dbApp.Machine, dbApp)
			events = append(events, ev)
		}
		// First three values indicate that there is nothing to do in the database.
		// The events variable carries the list of generated events. The last value
		// indicates that we have detected no daemons with no configuration change.
		// In fact, we didn't go that far to check that.
		return false, false, nil, events, nil
	}

	newActive := true
	sameConfigDaemons := make(map[string]bool)

	// Let's make sure that all daemons have a back pointer to the app because
	// it will be needed by event center to generate events.
	for name := range daemonsMap {
		daemonsMap[name].App = dbApp
	}

	// Go over the new daemons (received from config-get) and detect any changes
	// to the currently known state of these daemons.
	for name := range daemonsMap {
		daemon := daemonsMap[name]
		// If all daemons are active then whole app is active.
		newActive = newActive && daemon.Active

		// Add this daemon to the list of detected daemons.
		newDaemons = append(newDaemons, daemon)

		// Determine changes in app daemons state and store them as events.
		// Later this events will be passed to EventCenter when all the changes
		// are stored in database.
		oldDaemon := dbApp.GetDaemonByName(daemon.Name)
		if oldDaemon == nil {
			continue
		}

		// Add a pointer to the app in the daemon because it will be used by the
		// event center when new events are created.
		oldDaemon.App = dbApp

		// Check whether the daemon has transitioned between active and inactive states.
		if daemon.Active != oldDaemon.Active {
			lvl := dbmodel.EvWarning
			text := "{daemon} is "
			if daemon.Active && !oldDaemon.Active {
				// Daemon was inactive and now it is active again.
				text += "reachable now"
			} else if !daemon.Active && oldDaemon.Active {
				// Daemon was active and now it is inactive. This has higher
				// severity.
				text += "unreachable"
				lvl = dbmodel.EvError
			}
			errStr := daemonsErrors[oldDaemon.Name]
			ev := eventcenter.CreateEvent(lvl, text, errStr, dbApp.Machine, dbApp, oldDaemon)
			events = append(events, ev)

			// Check if daemon has been restarted.
		} else if daemon.Uptime < oldDaemon.Uptime {
			text := "{daemon} has been restarted"
			ev := eventcenter.CreateEvent(dbmodel.EvWarning, text, dbApp.Machine, dbApp, oldDaemon)
			events = append(events, ev)
		}

		// Check if daemon version has changed.
		if daemon.Version != oldDaemon.Version {
			text := fmt.Sprintf("{daemon} version changed from %s to %s",
				oldDaemon.Version, daemon.Version)
			ev := eventcenter.CreateEvent(dbmodel.EvWarning, text, dbApp.Machine, dbApp, oldDaemon)
			events = append(events, ev)
		}

		// Check if the daemon's configuration remains the same.
		if same := handleConfigEvent(daemon, oldDaemon, &events); same {
			// Daemons configuration seems to be the same since previous update. Let's
			// make a note of it so we don't unnecessarily process its configuration.
			sameConfigDaemons[daemon.Name] = true
			log.Infof("Configuration of Kea: id %d, daemon: %s has not changed since last fetch; skipping database update for that daemon", dbApp.ID, daemon.Name)
		}
	}

	return newActive, true, newDaemons, events, sameConfigDaemons
}

// Detects a situation that the daemon configuration remains the same after update
// or raises events about config change otherwise.
func handleConfigEvent(daemon, oldDaemon *dbmodel.Daemon, events *[]*dbmodel.Event) bool {
	if daemon.KeaDaemon != nil && oldDaemon.KeaDaemon != nil {
		if daemon.KeaDaemon.ConfigHash == oldDaemon.KeaDaemon.ConfigHash {
			return true
		}
		// Raise this event only if we're certain that the configuration has
		// changed based on the comparison of the hash values.
		text := "Configuration change detected for {daemon}"
		ev := eventcenter.CreateEvent(dbmodel.EvInfo, text, daemon)
		*events = append(*events, ev)
	}
	return false
}

// Removes associations between the daemon, shared networks, subnets and hosts.
func deleteDaemonAssociations(tx *pg.Tx, daemon *dbmodel.Daemon) error {
	// Remove associations between the daemon and the existing hosts.
	// We will recreate the associations using new configuration.
	_, err := dbmodel.DeleteDaemonFromHosts(tx, daemon.ID, dbmodel.HostDataSourceConfig)
	if err != nil {
		return err
	}

	// Remove associations between the daemon and the subnets. We will
	// recreate the associations using new configuration.
	_, err = dbmodel.DeleteDaemonFromSubnets(tx, daemon.ID)
	if err != nil {
		return err
	}

	// Remove associations between the daemon and the subnets. We will
	// recreate the associations using new configuration.
	_, err = dbmodel.DeleteDaemonFromSharedNetworks(tx, daemon.ID)
	if err != nil {
		return err
	}

	// Remove associations between the daemon and the services. We will
	// recreate the associations using new configuration.
	_, err = dbmodel.DeleteDaemonFromServices(tx, daemon.ID)
	if err != nil {
		return err
	}

	return nil
}

// Deletes empty shared networks and orphaned subnets and hosts.
func deleteEmptyAndOrphanedObjects(tx *pg.Tx) error {
	// Removed the hosts that no longer belong to any app.
	_, err := dbmodel.DeleteOrphanedHosts(tx)
	if err != nil {
		return err
	}

	// Remove the subnets that no longer belong to any daemon.
	_, err = dbmodel.DeleteOrphanedSubnets(tx)
	if err != nil {
		return err
	}

	// Delete the shared networks that no longer belong to any daemon.
	_, err = dbmodel.DeleteOrphanedSharedNetworks(tx)
	if err != nil {
		return err
	}
	return nil
}

// Detects and commits the discovered services into the database for each
// daemon belonging to the machine.
func detectAndCommitServices(tx *pg.Tx, machine *dbmodel.Machine) error {
	for _, daemon := range machine.Daemons {
		// Check what HA services the daemon belongs to.
		services, err := DetectHAServices(tx, daemon)
		if err != nil {
			return err
		}

		// For the given daemon, iterate over the services and add/update them in the
		// database.
		err = dbmodel.CommitServicesIntoDB(tx, services, daemon)
		if err != nil {
			return err
		}
	}
	return nil
}

// Adds events specific to the recent machine updates.
func addOnCommitMachineEvents(machine *dbmodel.Machine, addedDaemons, deletedDaemons []*dbmodel.Daemon, state *AppStateMeta, eventCenter eventcenter.EventCenter) {
	if machine.ID == 0 {
		eventCenter.AddInfoEvent("added {machine}", machine)
	}

	for _, daemon := range deletedDaemons {
		daemon.Machine = machine
		eventCenter.AddInfoEvent("removed {daemon} from {machine}", machine, daemon)
	}
	for _, daemon := range addedDaemons {
		daemon.Machine = machine
		eventCenter.AddInfoEvent("added {daemon} to {machine}", machine, daemon)
	}
	if state != nil {
		for _, ev := range state.Events {
			eventCenter.AddEvent(ev)
		}
	}
}

// Adds events specific to the recent app/daemon subnets updates.
func addOnCommitSubnetEvents(app *dbmodel.App, daemon *dbmodel.Daemon, addedSubnets []*dbmodel.Subnet, eventCenter eventcenter.EventCenter) {
	if len(addedSubnets) > 0 {
		// add event per subnet only if there is not more than 10 subnets
		if len(addedSubnets) < 10 {
			for _, sn := range addedSubnets {
				eventCenter.AddInfoEvent("added {subnet} to {daemon} in {app}", sn, daemon, app)
			}
		}
		t := fmt.Sprintf("added %d subnets to {daemon} in {app}", len(addedSubnets))
		eventCenter.AddInfoEvent(t, daemon, app)
	}
}

// Inserts or updates information about Kea app in the database. Next, it extracts
// Kea's configurations and uses to either update or create new shared networks,
// subnets and pools. Finally, the relations between the subnets and the Kea app
// are created. Note that multiple apps can be associated with the same subnet.
func CommitAppIntoDB(db *dbops.PgDB, app *dbmodel.App, eventCenter eventcenter.EventCenter, state *AppStateMeta, lookup keaconfig.DHCPOptionDefinitionLookup) (err error) {
	err = db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		// Let's first add or update the app in the database. It must be done
		// before detecting the subnets and shared networks because we need to
		// know daemon IDs to associate the subnets and shared networks with.
		// The daemon IDs are assigned by the database when the daemons are
		// first added.
		var addedDaemons, deletedDaemons []*dbmodel.Daemon
		if app.ID == 0 {
			// New app, insert it.
			addedDaemons, err = dbmodel.AddApp(tx, app)
		} else {
			// Existing app, update it if needed.
			addedDaemons, deletedDaemons, err = dbmodel.UpdateApp(tx, app)
		}

		if err != nil {
			return err
		}

		networks := make(map[string][]dbmodel.SharedNetwork)
		subnets := make(map[string][]dbmodel.Subnet)
		globalHosts := make(map[string][]dbmodel.Host)

		for _, daemon := range app.Daemons {
			if state != nil && state.SameConfigDaemons != nil {
				// There are quite frequent cases when the daemons' configurations haven't
				// changed since last update. If that's the case, this map contains the
				// names of these daemons. For such daemons we should safely skip processing
				// subnets and shared networks. This saves many CPU cycles.
				if ok := state.SameConfigDaemons[daemon.Name]; ok {
					continue
				}
			}

			// Remove daemon associations with hosts, subnets and shared networks.
			err = deleteDaemonAssociations(tx, daemon)
			if err != nil {
				return err
			}

			// Go over the shared networks and subnets stored in the Kea configuration
			// and match them with the existing entries in the database. If some of
			// the shared networks or subnets do not exist they are instantiated and
			// returned here.
			networks[daemon.Name], subnets[daemon.Name], err = detectDaemonNetworks(tx, daemon, lookup)
			if err != nil {
				err = errors.Wrapf(err, "unable to detect subnets and shared networks for Kea daemon %s belonging to app with ID %d", daemon.Name, app.ID)
				return err
			}

			if state == nil || state.SameConfigDaemons == nil || !state.SameConfigDaemons[daemon.Name] {
				// Go over the global reservations stored in the Kea configuration and
				// match them with the existing global hosts.
				globalHosts[daemon.Name], err = detectGlobalHostsFromConfig(tx, daemon, lookup)
				if err != nil {
					err = errors.Wrapf(err, "unable to detect global host reservations for Kea daemon %d", daemon.ID)
					return err
				}
			}
		}

		// Add events to the database.
		addOnCommitAppEvents(app, addedDaemons, deletedDaemons, state, eventCenter)

		for _, daemon := range app.Daemons {
			// For the given daemon, iterate over the networks and subnets and update their
			// global instances accordingly in the database.
			addedSubnets, err := dbmodel.CommitNetworksIntoDB(tx, networks[daemon.Name], subnets[daemon.Name])
			if err != nil {
				return err
			}

			// For the given app, iterate over the global hosts and update their instances
			// in the database or insert them into the database.
			if err = dbmodel.CommitGlobalHostsIntoDB(tx, globalHosts[daemon.Name]); err != nil {
				return err
			}

			// Add subnet related events to the database.
			addOnCommitSubnetEvents(app, daemon, addedSubnets, eventCenter)
		}

		// Detect and commit discovered services for each daemon.
		if err = detectAndCommitServices(tx, app); err != nil {
			return err
		}

		// Remove empty shared networks and orphaned subnets and hosts.
		err = deleteEmptyAndOrphanedObjects(tx)
		return err
	})
	return errors.Wrapf(err, "problem committing updates for app %d", app.ID)
}
