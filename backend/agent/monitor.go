package agent

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	bind9config "isc.org/stork/daemoncfg/bind9"
	pdnsconfig "isc.org/stork/daemoncfg/pdns"
	"isc.org/stork/daemonctrl/daemonname"
	storkutil "isc.org/stork/util"
)

// Operations provided by the Stork agent to set up daemon-related configuration.
type AgentManager interface {
	AllowLog(path string)
}

// Supported protocol types.
type ProtocolType string

const (
	ProtocolTypeHTTP   ProtocolType = "http"
	ProtocolTypeHTTPS  ProtocolType = "https"
	ProtocolTypeSocket ProtocolType = "unix"
	ProtocolTypeRNDC   ProtocolType = "rndc"
)

// An access point for an application to retrieve information such
// as status or metrics.
type AccessPoint struct {
	Type     string
	Address  string
	Port     int64
	Protocol string
	Key      string
}

// Checks if two access points are equal.
func (ap *AccessPoint) IsEqual(other AccessPoint) bool {
	return ap.Type == other.Type &&
		ap.Address == other.Address &&
		ap.Port == other.Port &&
		ap.Protocol == other.Protocol &&
		ap.Key == other.Key
}

// String representation of an access point.
func (ap *AccessPoint) String() string {
	var b strings.Builder
	b.WriteString(ap.Type)
	b.WriteString(": ")
	b.WriteString(storkutil.HostWithPortURL(ap.Address, ap.Port, ap.Protocol))
	if ap.Type == AccessPointControl {
		b.WriteString(" (auth key: ")
		if ap.Key != "" {
			b.WriteString("found")
		} else {
			b.WriteString("not found")
		}
		b.WriteString(")")
	}
	return b.String()
}

// Currently supported types are: "control" and "statistics".
const (
	AccessPointControl    = "control"
	AccessPointStatistics = "statistics"
)

type Daemon interface {
	GetName() daemonname.Name
	GetAccessPoint(apType string) *AccessPoint
	GetAccessPoints() []AccessPoint
	// Checks if data of two daemons are equal.
	IsEqual(other Daemon) bool
	// Called when the monitor newly detects the daemon.
	// It allows the daemon to perform initialization tasks
	// such as starting a background goroutine.
	Bootstrap() error
	// Called when the monitor no longer detects the daemon.
	// It allows the daemon to perform cleanup tasks such as
	// stopping a background goroutine.
	Cleanup() error
	// Performs periodic processing of the daemon, e.g., detect logs or
	// refresh zone inventory.
	Evaluate(AgentManager) error
	String() string
}

// Daemon information. This structure is embedded
// in app specific structures like KeaApp and Bind9App.
type daemon struct {
	Name         daemonname.Name
	AccessPoints []AccessPoint
}

// Return the name of the daemon process.
func (d *daemon) GetName() daemonname.Name {
	return d.Name
}

// Returns all access points of the daemon.
func (d *daemon) GetAccessPoints() []AccessPoint {
	return d.AccessPoints
}

// Returns an access point of a given type. If the access point is not found,
// it returns nil.
func (d *daemon) GetAccessPoint(accessPointType string) *AccessPoint {
	for _, ap := range d.AccessPoints {
		if ap.Type == accessPointType {
			return &ap
		}
	}
	return nil
}

// String representation of a daemon.
func (d *daemon) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: ", d.Name))

	for i := 0; i < len(d.AccessPoints)-1; i++ {
		b.WriteString(d.AccessPoints[i].String())
		b.WriteString(", ")
	}
	if len(d.AccessPoints) > 0 {
		b.WriteString(d.AccessPoints[len(d.AccessPoints)-1].String())
	}

	return b.String()
}

// Checks if two applications are the same. It checks the name and access
// points including their configuration.
func (d *daemon) IsEqual(other Daemon) bool {
	if d.Name != other.GetName() {
		return false
	}

	otherAccessPoints := other.GetAccessPoints()
	if len(d.AccessPoints) != len(otherAccessPoints) {
		return false
	}

	for _, otherAccessPoint := range otherAccessPoints {
		thisAccessPoint := d.GetAccessPoint(otherAccessPoint.Type)
		if thisAccessPoint == nil {
			return false
		}
		if !thisAccessPoint.IsEqual(otherAccessPoint) {
			return false
		}
	}
	return true
}

// An interface representing the DNS daemon.
type DNSDaemon interface {
	Daemon
	GetZoneInventory() *zoneInventory
}

// Converts a process name to a daemon name. If the process name
// is not recognized, it returns an empty string.
func convertProcessNameToDaemonName(procName string) daemonname.Name {
	switch procName {
	case "kea-dhcp4":
		return daemonname.DHCPv4
	case "kea-dhcp6":
		return daemonname.DHCPv6
	case "kea-d2":
		return daemonname.D2
	case "kea-ctrl-agent":
		return daemonname.CA
	case "named":
		return daemonname.Bind9
	case "pdns_server":
		return daemonname.PDNS
	default:
		return ""
	}
}

// The daemon monitor is responsible for detecting the daemons
// running in the operating system and periodically refreshing their states.
// They are available through assessors.
type Monitor interface {
	GetDaemons() []Daemon
	GetDaemonByAccessPoint(apType, address string, port int64) Daemon
	Start(AgentManager)
	Shutdown()
}

type monitor struct {
	requests                chan chan []Daemon // input to monitor, ie. channel for receiving requests
	quit                    chan bool          // channel for stopping app monitor
	running                 bool
	wg                      *sync.WaitGroup
	commander               storkutil.CommandExecutor
	processManager          *ProcessManager
	bind9FileParser         bind9FileParser
	explicitBind9ConfigPath string
	pdnsConfigParser        pdnsConfigParser
	keaHTTPClientConfig     HTTPClientConfig

	// List of detected daemons on the host.
	// Nil if the monitor has no perform detection yet.
	daemons []Daemon
}

// Creates an Monitor instance. It used to start it as well, but this is now done
// by a dedicated method Start(). Make sure you call Start() before using daemon
// monitor.
func NewMonitor(explicitBind9ConfigPath string, keaHTTPClientConfig HTTPClientConfig) Monitor {
	sm := &monitor{
		requests:                make(chan chan []Daemon),
		quit:                    make(chan bool),
		wg:                      &sync.WaitGroup{},
		commander:               storkutil.NewSystemCommandExecutor(),
		processManager:          NewProcessManager(),
		bind9FileParser:         bind9config.NewParser(),
		pdnsConfigParser:        pdnsconfig.NewParser(),
		explicitBind9ConfigPath: explicitBind9ConfigPath,
		keaHTTPClientConfig:     keaHTTPClientConfig,
		running:                 false,
		daemons:                 nil,
	}
	return sm
}

// This function starts the actual monitor. This start is delayed in case we want to only
// do command line parameters parsing, e.g. to print version or help and quit.
func (sm *monitor) Start(storkAgent AgentManager) {
	sm.wg.Add(1)
	go sm.run(storkAgent)
}

// Run the main loop of the monitor. It continually detects the daemons
// running on the host and refreshes their states.
func (sm *monitor) run(storkAgent AgentManager) {
	log.Printf("Started daemon monitor")

	sm.running = true
	defer sm.wg.Done()

	// Run app detection one time immediately at startup.
	sm.detectDaemons()

	// Evaluate all detected daemons.
	sm.evaluateDaemons(storkAgent)

	// Prepare ticker.
	const detectionInterval = 10 * time.Second
	ticker := time.NewTicker(detectionInterval)
	defer ticker.Stop()

	for {
		select {
		case ret := <-sm.requests:
			// Process user request.
			ret <- sm.daemons

		case <-ticker.C:
			// Periodic detection.
			ticker.Stop()

			sm.detectDaemons()
			sm.evaluateDaemons(storkAgent)

			// Reset ticker.
			ticker.Reset(detectionInterval)

		case <-sm.quit:
			// exit run
			log.Printf("Stopped app monitor")
			sm.running = false
			return
		}
	}
}

// Splits the daemons into newly started, untouched (already existed), untouched (duplicated) and stopped ones.
func splitDaemonsByTransition(previous, next []Daemon) (started, unchanged, unchangedDuplicated, stopped []Daemon) {
	// Daemons no longer running.
	stoppedMap := make(map[int]bool)
	for i := 0; i < len(previous); i++ {
		stoppedMap[i] = true
	}

	// Daemons newly started.
	startedMap := make(map[int]bool)
	for i := 0; i < len(next); i++ {
		startedMap[i] = true
	}

	// Daemons unchanged.
	unchangedMap := make(map[int]bool)
	unchangedDuplicatedMap := make(map[int]bool)

	for ip, p := range previous {
		for in, n := range next {
			if p.IsEqual(n) {
				// Daemon is still running.
				stoppedMap[ip] = false
				startedMap[in] = false
				unchangedMap[ip] = true
				unchangedDuplicatedMap[in] = true
				break
			}
		}
	}

	for ip, isStopped := range stoppedMap {
		if isStopped {
			stopped = append(stopped, previous[ip])
			log.Infof("Daemon stopped: %s", previous[ip].String())
		}
	}

	for in, isStarted := range startedMap {
		if isStarted {
			started = append(started, next[in])
			log.Infof("Daemon started: %s", next[in].String())
		}
	}

	for in, isUnchanged := range unchangedMap {
		if isUnchanged {
			unchanged = append(unchanged, previous[in])
		}
	}

	for in, isUnchangedDuplicated := range unchangedDuplicatedMap {
		if isUnchangedDuplicated {
			unchangedDuplicated = append(unchangedDuplicated, next[in])
		}
	}

	return
}

// Analyzes the processes running on the host and detects supported daemons.
func (sm *monitor) detectDaemons() {
	var daemons []Daemon

	// Lists processes running on the host and detectable by the monitor.
	processes, _ := sm.processManager.ListProcesses()

	for _, p := range processes {
		procName, _ := p.getName()
		daemonName := convertProcessNameToDaemonName(procName)
		if daemonName == "" {
			// Process is not a supported daemon.
			continue
		}

		switch daemonName {
		case daemonname.DHCPv4, daemonname.DHCPv6, daemonname.D2, daemonname.CA:
			// Kea DHCP server.
			detectedDaemons, err := detectKeaDaemons(p, sm.keaHTTPClientConfig, sm.commander)
			if err != nil {
				log.WithField("daemon", daemonName).WithError(err).Warn("Failed to detect Kea daemon(s)")
				continue
			}
			daemons = append(daemons, detectedDaemons...)

		case daemonname.Bind9:
			// BIND 9 DNS server.
			detectedDaemon, err := detectBind9Daemon(
				p,
				sm.commander,
				sm.explicitBind9ConfigPath,
				sm.bind9FileParser,
			)
			if err != nil {
				log.WithError(err).Warnf("Failed to detect BIND 9 DNS server daemon")
				continue
			}
			daemons = append(daemons, detectedDaemon)
		case daemonname.PDNS:
			// PowerDNS server.
			detectedDaemon, err := detectPowerDNSDaemon(p, sm.pdnsConfigParser)
			if err != nil {
				log.WithError(err).Warn("Failed to detect PowerDNS server app")
				continue
			}
			daemons = append(daemons, detectedDaemon)
		default:
			// This should never be the case given that we list only supported processes.
			log.Warnf("Unsupported daemon name %s", daemonName)
			continue
		}
	}

	if len(daemons) == 0 && (sm.daemons == nil || len(sm.daemons) != 0) {
		// It is a first detection when no daemon is detected.
		// Agent is starting up but no daemon to monitor has been detected.
		// Usually, the agent is installed with at least one monitored daemon.
		// The below message is printed for easier troubleshooting.
		log.Warn("No daemon detected for monitoring; please check if they are running, and Stork can communicate with them.")
		sm.daemons = []Daemon{}
	}

	startedDaemons, runningDaemons, duplicatedDaemons, stoppedDaemons := splitDaemonsByTransition(sm.daemons, daemons)
	newMonitorDaemons := []Daemon{} // Non-nil slice.
	newMonitorDaemons = append(newMonitorDaemons, runningDaemons...)

	for _, d := range stoppedDaemons {
		if err := d.Cleanup(); err != nil {
			log.WithError(err).WithField("daemon", d.String()).Warn("Failed to cleanup daemon")
		}
	}

	for _, d := range duplicatedDaemons {
		if err := d.Cleanup(); err != nil {
			log.WithError(err).WithField("daemon", d.String()).Warn("Failed to cleanup duplicated daemon")
		}
	}

	for _, d := range startedDaemons {
		if err := d.Bootstrap(); err != nil {
			log.WithError(err).WithField("daemon", d.String()).Warn("Failed to bootstrap daemon")
		}
		newMonitorDaemons = append(newMonitorDaemons, d)
	}

	sm.daemons = newMonitorDaemons
}

// Evaluates the detected daemons.
func (sm *monitor) evaluateDaemons(storkAgent AgentManager) {
	for _, d := range sm.daemons {
		if err := d.Evaluate(storkAgent); err != nil {
			log.WithError(err).WithField("daemon", d.String()).Warn("Failed to evaluate daemon")
		}
	}
}

// Get a list of detected daemons by a monitor.
func (sm *monitor) GetDaemons() []Daemon {
	ret := make(chan []Daemon)
	sm.requests <- ret
	daemons := <-ret
	return daemons
}

// Get a daemon from a monitor that matches provided params.
func (sm *monitor) GetDaemonByAccessPoint(apType, address string, port int64) Daemon {
	for _, d := range sm.GetDaemons() {
		for _, ap := range d.GetAccessPoints() {
			if ap.Type == apType && ap.Address == address && ap.Port == port {
				return d
			}
		}
	}
	return nil
}

// Shut down monitor. Stop background goroutines.
func (sm *monitor) Shutdown() {
	for _, d := range sm.GetDaemons() {
		err := d.Cleanup()
		if err != nil {
			log.WithError(err).Warnf("Failed to cleanup daemon %s", d.String())
		}
	}
	sm.quit <- true
	sm.wg.Wait()
}
