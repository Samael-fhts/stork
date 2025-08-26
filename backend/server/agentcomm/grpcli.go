package agentcomm

import (
	"context"
	"io"
	"iter"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	agentapi "isc.org/stork/api"
	"isc.org/stork/appcfg/dnsconfig"
	keactrl "isc.org/stork/appctrl/kea"
	"isc.org/stork/appdata/bind9stats"
	pdnsdata "isc.org/stork/appdata/pdns"
	dbmodel "isc.org/stork/server/database/model"
	storkutil "isc.org/stork/util"
)

var _ ConnectedAgents = (*connectedAgentsImpl)(nil)

// An access point for a daemon to retrieve information such
// as status or metrics.
type AccessPoint struct {
	Type              string
	Address           string
	Port              int64
	Key               string
	UseSecureProtocol bool
}

// Currently supported types are: "control" and "statistics".
const (
	AccessPointControl    = "control"
	AccessPointStatistics = "statistics"
)

// Daemon names returned by the Stork Agent.
type DaemonName = string

const (
	// It is a deprecated name that Stork agents prior 2.3 use for the Kea
	// Control Agent.
	DaemonNameKea    DaemonName = "kea"
	DaemonNameDHCPv4 DaemonName = "dhcp4"
	DaemonNameDHCPv6 DaemonName = "dhcp6"
	DaemonNameD2     DaemonName = "d2"
	DaemonNameCA     DaemonName = "ca"
	DaemonNamePDNS   DaemonName = "pdns"
	DaemonNameBind9  DaemonName = "bind9"
)

// An interface to a daemon that can receive commands from Stork.
// Kea daemon receiving control commands is an example.
type ControlledDaemon interface {
	GetControlAccessPoint() (string, int64, string, bool, error)
	GetStatisticsAccessPoint() (string, int64, string, bool, error)
	GetMachineTag() dbmodel.MachineTag
	GetName() DaemonName
}

// An interface to a machine that can receive commands from Stork.
type ControlledMachine interface {
	GetAddress() string
	GetAgentPort() int64
}

// The daemon entry detected by an agent. It unambiguously indicates the
// daemon location.
type Daemon struct {
	Name         DaemonName
	AccessPoints []AccessPoint
	Machine      dbmodel.MachineTag
}

// Implements the agentcomm.ControlledDaemon interface.
var _ ControlledDaemon = (*Daemon)(nil)

// Return the name of the daemon.
func (d *Daemon) GetName() DaemonName {
	return d.Name
}

// Returns the control access point of the daemon. It returns an error if
// no control access point is found.
func (d *Daemon) GetControlAccessPoint() (string, int64, string, bool, error) {
	for _, ap := range d.AccessPoints {
		if ap.Type == AccessPointControl {
			return ap.Address, ap.Port, ap.Key, ap.UseSecureProtocol, nil
		}
	}
	return "", 0, "", false, errors.Errorf("no control access point for daemon %s", d.Name)
}

// Returns the statistics access point of the daemon. It returns an error if
// no statistics access point is found.
func (d *Daemon) GetStatisticsAccessPoint() (string, int64, string, bool, error) {
	for _, ap := range d.AccessPoints {
		if ap.Type == AccessPointStatistics {
			return ap.Address, ap.Port, ap.Key, ap.UseSecureProtocol, nil
		}
	}
	return "", 0, "", false, errors.Errorf("no statistics access point for daemon %s", d.Name)
}

// Returns the machine tag of the daemon.
func (d *Daemon) GetMachineTag() dbmodel.MachineTag {
	return d.Machine
}

// State of the machine. It describes multiple properties of the machine like number of CPUs
// or operating system name and version.
type State struct {
	Address              string
	AgentVersion         string
	Cpus                 int64
	CpusLoad             string
	Memory               int64
	Hostname             string
	Uptime               int64
	UsedMemory           int64
	Os                   string
	Platform             string
	PlatformFamily       string
	PlatformVersion      string
	KernelVersion        string
	KernelArch           string
	VirtualizationSystem string
	VirtualizationRole   string
	HostID               string
	LastVisitedAt        time.Time
	Error                string
	Daemons              []*Daemon
}

// MakeAccessPoint is an utility to make an array of one access point.
func MakeAccessPoint(tp, address, key string, port int64) []AccessPoint {
	return []AccessPoint{{
		Type:    tp,
		Address: address,
		Port:    port,
		Key:     key,
	}}
}

// An interface to the response from gRPC including a command status.
type agentResponse interface {
	GetStatus() *agentapi.Status
}

// Based on the gRPC request, response and an error it checks the state of
// the communication with an agent and returns this state. The status message
// is returned as a second parameter.
func (agents *connectedAgentsImpl) checkAgentCommState(stats *CommStats, reqData any, reqErr error) (CommErrorTransition, string) {
	commErrors := stats.GetAgentErrorCount(reqData)

	switch {
	case commErrors == 0 && reqErr != nil:
		// New communication issue.
		stats.IncreaseAgentErrorCount(reqData)
		return CommErrorNew, reqErr.Error()
	case commErrors > 0 && reqErr != nil:
		// Old communication issue
		stats.IncreaseAgentErrorCount(reqData)
		return CommErrorContinued, reqErr.Error()
	case commErrors == 0 && reqErr == nil:
		// Everything still ok.
		return CommErrorNone, ""
	case commErrors > 0 && reqErr == nil:
		// Communication resumed.
		stats.ResetAgentErrorCount(reqData)
		return CommErrorReset, ""
	}
	return CommErrorNone, ""
}

func (agents *connectedAgentsImpl) checkBind9CommState(stats *CommStatsBind9, accessPointType string, resp any) (CommErrorTransition, string) {
	var (
		status  *agentapi.Status
		details string
	)
	if statusResp, ok := resp.(agentResponse); ok {
		status = statusResp.GetStatus()
		details = statusResp.GetStatus().Message
	}
	if status == nil {
		return CommErrorNone, details
	}

	commErrors := stats.GetErrorCount(accessPointType)

	switch {
	case commErrors == 0 && status.GetCode() != agentapi.Status_OK:
		stats.IncreaseErrorCount(accessPointType)
		return CommErrorNew, details
	case commErrors > 0 && status.GetCode() != agentapi.Status_OK:
		stats.IncreaseErrorCount(accessPointType)
		return CommErrorContinued, details
	case commErrors == 0 && status.GetCode() == agentapi.Status_OK:
		return CommErrorNone, details
	case commErrors > 0 && status.GetCode() == agentapi.Status_OK:
		stats.ResetErrorCount(accessPointType)
		return CommErrorReset, details
	}
	return CommErrorNone, details
}

// Holds the communication states of the Kea daemons returned
// by the checkKeaCommState function.
type keaCommState struct {
	states map[dbmodel.DaemonName]CommErrorTransition
	// Contains an item for each command. If the command was successful, the
	// item is nil.
	errors map[dbmodel.DaemonName][]error
}

// Appends a new error.
func (s *keaCommState) appendError(daemon dbmodel.DaemonName, err error) {
	if s.errors == nil {
		s.errors = make(map[dbmodel.DaemonName][]error)
	}
	if s.errors[daemon] == nil {
		s.errors[daemon] = make([]error, 0)
	}
	s.errors[daemon] = append(s.errors[daemon], err)
}

// Returns number of errors recorded for a daemon.
func (s *keaCommState) getErrorCount(daemon dbmodel.DaemonName) int {
	if s.errors == nil {
		return 0
	}
	if s.errors[daemon] == nil {
		return 0
	}
	return len(s.errors[daemon])
}

// Returns errors recorded for a daemon.
func (s *keaCommState) getErrors(daemon dbmodel.DaemonName) []error {
	if s.errors == nil {
		return nil
	}
	return s.errors[daemon]
}

// Sets state for a daemon.
func (s *keaCommState) setState(daemon dbmodel.DaemonName, state CommErrorTransition) {
	if s.states == nil {
		s.states = make(map[dbmodel.DaemonName]CommErrorTransition)
	}
	s.states[daemon] = state
}

// Gets state for a daemon.
func (s *keaCommState) getState(daemon dbmodel.DaemonName) CommErrorTransition {
	if s.states == nil {
		return CommErrorNone
	}
	state, ok := s.states[daemon]
	if !ok {
		return CommErrorNone
	}
	return state
}

// It checks the communication state with the Kea daemons behind an agent. This
// function is called if there was no communication problem with an agent itself.
// If checks the status codes returned by the Kea Control Agent and returns the
// communication states for each of the daemons.
func (agents *connectedAgentsImpl) checkKeaCommState(stats *CommStatsKea, commands []keactrl.SerializableCommand, resp *agentapi.ForwardToKeaOverHTTPRsp) keaCommState {
	var state keaCommState
	uniqueDaemons := make(map[dbmodel.DaemonName]struct{})

	// Get all responses from the Kea server.
	for idx, daemonResp := range resp.GetKeaResponses() {
		command := commands[idx]
		daemons := command.GetDaemonsList()
		// It is expected that a single command is sent to a single daemon.
		// The multiple-daemon commands are supported only when the
		// communication is tunneled via Kea Control Agent. Stork sends the
		// commands to daemons directly for Kea daemons in 3 version, so it
		// always creates commands for a single daemon.
		daemon := daemons[0]
		uniqueDaemons[daemon] = struct{}{}

		if daemonResp.Status.Code != agentapi.Status_OK {
			message := "unknown error"
			if daemonResp.Status.Message != "" {
				message = daemonResp.Status.Message
			}

			err := errors.Errorf("received error while sending the command %s over GRPC: %s", command.GetCommand(), message)
			state.appendError(daemon, err)
			continue
		}

		var parsedResp []keactrl.ResponseHeader
		err := keactrl.UnmarshalResponseList(commands[idx], daemonResp.Response, &parsedResp)
		if err != nil {
			err := errors.WithMessage(err, "failed to parse Kea response")
			state.appendError(daemon, err)
			continue
		}

		for _, daemonResp := range parsedResp {
			if err := daemonResp.GetError(); err != nil {
				err := errors.Wrapf(err, "command %s failed", command.GetCommand())
				state.appendError(daemon, err)
			}
		}
	}

	for daemon := range uniqueDaemons {
		state.setState(daemon, stats.UpdateErrorCount(daemon, int64(state.getErrorCount(daemon))))
	}
	return state
}

// Check connectivity with a machine.
func (agents *connectedAgentsImpl) Ping(ctx context.Context, machine dbmodel.MachineTag) error {
	addrPort := net.JoinHostPort(machine.GetAddress(), strconv.FormatInt(machine.GetAgentPort(), 10))

	req := &agentapi.PingReq{}
	resp, err := agents.sendAndRecvViaQueue(addrPort, req)

	stats := agents.getConnectedAgentStats(machine.GetAddress(), machine.GetAgentPort())
	if stats == nil {
		return errors.Errorf("failed to get statistics for the non-existing agent %s", addrPort)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commState, details := agents.checkAgentCommState(stats, req, err)
	switch commState {
	case CommErrorNew:
		log.WithField("agent", addrPort).Warn("Failed to ping the agent")
		agents.eventCenter.AddErrorEvent("pinging Stork agent on {machine} failed", machine, dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("pinging Stork agent on {machine} succeeded", machine, dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithField("agent", addrPort).Warn("Failed to ping the Stork agent; the agent is still not responding")

	default:
		// Communication with the agent was ok and is still ok.
	}

	// If there was an error in communication with the agent, there is no need
	// to check the response because it is probably nil anyway. Return an error.
	if err != nil {
		return errors.Wrapf(err, "failed to ping the Stork agent %s", addrPort)
	}
	if _, ok := resp.(*agentapi.PingRsp); !ok {
		return errors.Wrapf(err, "wrong response for ping from the Stork agent %s", addrPort)
	}
	return nil
}

// Get machine statistics and version number.
func (agents *connectedAgentsImpl) GetState(ctx context.Context, machine dbmodel.MachineTag) (*State, error) {
	addrPort := net.JoinHostPort(machine.GetAddress(), strconv.FormatInt(machine.GetAgentPort(), 10))

	req := &agentapi.GetStateReq{}
	resp, err := agents.sendAndRecvViaQueue(addrPort, req)

	stats := agents.getConnectedAgentStats(machine.GetAddress(), machine.GetAgentPort())
	if stats == nil {
		return nil, errors.Errorf("failed to get statistics for the non-existing agent %s", addrPort)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commState, details := agents.checkAgentCommState(stats, req, err)
	switch commState {
	case CommErrorNew:
		log.WithField("agent", addrPort).Warn("Failed to get state from the Stork agent")
		agents.eventCenter.AddErrorEvent("communication with Stork agent on {machine} to get state failed", machine, dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication with Stork agent on {machine} to get state succeeded", machine, dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithField("agent", addrPort).Warn("Failed to get state from the Stork agent; the agent is still not responding")

	default:
		// Communication with the agent was ok and is still ok.
	}

	// If there was an error in communication with the agent, there is no need
	// to check the response because it is probably nil anyway. Return an derror.
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get state from agent %s", addrPort)
	}

	// Communication successful. Let's decode the response.
	grpcState, ok := resp.(*agentapi.GetStateRsp)
	if !ok || grpcState == nil {
		return nil, errors.Errorf("wrong response to get state from agent %s", addrPort)
	}

	var daemons []*Daemon
	for _, app := range grpcState.Apps {
		var accessPoints []AccessPoint

		for _, point := range app.AccessPoints {
			accessPoints = append(accessPoints, AccessPoint{
				Type:              point.Type,
				Address:           point.Address,
				Port:              point.Port,
				Key:               point.Key,
				UseSecureProtocol: point.UseSecureProtocol,
			})
		}

		daemons = append(daemons, &Daemon{
			Name:         DaemonName(app.Type),
			AccessPoints: accessPoints,
			Machine:      machine,
		})
	}

	state := State{
		Address:              machine.GetAddress(),
		AgentVersion:         grpcState.AgentVersion,
		Cpus:                 grpcState.Cpus,
		CpusLoad:             grpcState.CpusLoad,
		Memory:               grpcState.Memory,
		Hostname:             grpcState.Hostname,
		Uptime:               grpcState.Uptime,
		UsedMemory:           grpcState.UsedMemory,
		Os:                   grpcState.Os,
		Platform:             grpcState.Platform,
		PlatformFamily:       grpcState.PlatformFamily,
		PlatformVersion:      grpcState.PlatformVersion,
		KernelVersion:        grpcState.KernelVersion,
		KernelArch:           grpcState.KernelArch,
		VirtualizationSystem: grpcState.VirtualizationSystem,
		VirtualizationRole:   grpcState.VirtualizationRole,
		HostID:               grpcState.HostID,
		LastVisitedAt:        storkutil.UTCNow(),
		Error:                grpcState.Error,
		Daemons:              daemons,
	}

	return &state, nil
}

// The extracted output of the RNDC command.
type RndcOutput struct {
	Output string
}

// Forwards an RNDC command to named.
func (agents *connectedAgentsImpl) ForwardRndcCommand(ctx context.Context, daemon ControlledDaemon, command string) (*RndcOutput, error) {
	agentAddress := daemon.GetMachineTag().GetAddress()
	agentPort := daemon.GetMachineTag().GetAgentPort()

	// Get rndc control settings
	ctrlAddress, ctrlPort, _, _, err := daemon.GetControlAccessPoint()
	if err != nil {
		return nil, err
	}

	addrPort := net.JoinHostPort(agentAddress, strconv.FormatInt(agentPort, 10))

	// Prepare the on-wire representation of the commands.
	req := &agentapi.ForwardRndcCommandReq{
		Address: ctrlAddress,
		Port:    ctrlPort,
		RndcRequest: &agentapi.RndcRequest{
			Request: command,
		},
	}

	// Send the command to the Stork Agent.
	resp, err := agents.sendAndRecvViaQueue(addrPort, req)

	stats := agents.getConnectedAgentStats(agentAddress, agentPort)
	if stats == nil {
		return nil, errors.Errorf("failed to get statistics for the non-existing agent %s", addrPort)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commState, details := agents.checkAgentCommState(stats, req, err)
	switch commState {
	case CommErrorNew:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"rndc":  net.JoinHostPort(ctrlAddress, strconv.FormatInt(ctrlPort, 10)),
		}).Warnf("Failed to send the rndc command: %s", command)
		agents.eventCenter.AddErrorEvent("communication with Stork agent on {machine} to forward rndc command failed", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication with Stork agent on {machine} to forward rndc command succeeded", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"rndc":  net.JoinHostPort(ctrlAddress, strconv.FormatInt(ctrlPort, 10)),
		}).Warnf("Failed to send the rndc command to the Stork agent: %s; agent is still not responding", command)

	default:
		// Communication with the agent was ok and is still ok.
	}

	bind9Stats := stats.GetBind9Stats()

	// Check the result of the communication between the Stork agent and named by
	// examining the returned status code.
	commState, details = agents.checkBind9CommState(bind9Stats, dbmodel.AccessPointControl, resp)
	switch commState {
	case CommErrorNew:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"rndc":  net.JoinHostPort(ctrlAddress, strconv.FormatInt(ctrlPort, 10)),
		}).Warnf("Failed to send the rndc command: %s", command)
		agents.eventCenter.AddErrorEvent("communication between the Stork agent on {machine} and {daemon} to forward rndc command failed", daemon.GetMachineTag(), daemon, dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication between the Stork agent on {machine} and {daemon} to forward rndc command succeeded", daemon.GetMachineTag(), daemon, dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"rndc":  net.JoinHostPort(ctrlAddress, strconv.FormatInt(ctrlPort, 10)),
		}).Warnf("Failed to send the rndc command from Stork agent to named: %s; named is still returning an error", command)

	default:
		// Communication between the Stork agent and named was ok and is still ok.
	}

	// Stork agent returned an error.
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send the rndc command to Stork agent %s", addrPort)
	}

	// Communication with the Stork agent was ok, but named returned an error.
	if commState != CommErrorReset && commState != CommErrorNone {
		err = errors.Errorf("error communicating between Stork agent %s and named to send rndc command: %s", addrPort, details)
		return nil, err
	}

	response, ok := resp.(*agentapi.ForwardRndcCommandRsp)
	if !ok || response == nil {
		return nil, errors.Errorf("wrong response to the rndc command from the Stork agent %s", addrPort)
	}

	result := &RndcOutput{
		Output: "",
	}

	// named has responded but the response may also contain an error status.
	rndcResponse := response.GetRndcResponse()

	// If the status is ok, let's just return the result.
	if rndcResponse.Status.Code == agentapi.Status_OK {
		result.Output = rndcResponse.Response
		bind9Stats.ResetErrorCount(dbmodel.AccessPointControl)
		return result, nil
	}

	// Status code was not ok, so let's record an error message.
	err = errors.New(response.Status.Message)

	// Bump up error statistics. If this is a consecutive error let's
	// just return it and not log it again and again.
	if bind9Stats.IncreaseErrorCount(dbmodel.AccessPointControl) > 1 {
		err = errors.Errorf("failed to send rndc command via the agent %s; BIND 9 is still failing",
			agentAddress)
		return nil, err
	}
	// This is apparently the first error like this. Let's log it.
	log.WithFields(log.Fields{
		"agent": addrPort,
		"rndc":  net.JoinHostPort(ctrlAddress, strconv.FormatInt(ctrlPort, 10)),
	}).Warnf("named returned an error status to the RNDC command: %s", command)

	return result, err
}

// Forwards a statistics request via the Stork Agent to the named daemon and
// then parses the response. statsAddress, statsPort are used to construct
// base HTTP URL of the statistics channel. The requestType parameter is used
// to specify the path (or several paths in case of sequential requests) to
// the statistics-channel of the named daemon.
func (agents *connectedAgentsImpl) ForwardToNamedStats(ctx context.Context, daemon ControlledDaemon, requestType ForwardToNamedStatsRequestType, statsOutput any) error {
	addrPort := net.JoinHostPort(daemon.GetMachineTag().GetAddress(), strconv.FormatInt(daemon.GetMachineTag().GetAgentPort(), 10))
	statsAddress, statsPort, _, isSecure, err_ := daemon.GetStatisticsAccessPoint()
	if err_ != nil {
		return errors.WithMessage(err_, "failed to get statistics access point for daemon")
	}
	statsURL := storkutil.HostWithPortURL(statsAddress, statsPort, isSecure)

	// Prepare the on-wire representation of the commands.
	req := &agentapi.ForwardToNamedStatsReq{
		Url:          statsURL,
		StatsAddress: statsAddress,
		StatsPort:    statsPort,
		RequestType:  requestType,
	}
	req.NamedStatsRequest = &agentapi.NamedStatsRequest{
		Request: "",
	}

	// Send the commands to the Stork Agent.
	resp, err := agents.sendAndRecvViaQueue(addrPort, req)

	stats := agents.getConnectedAgentStats(daemon.GetMachineTag().GetAddress(), daemon.GetMachineTag().GetAgentPort())
	if stats == nil {
		return errors.Errorf("failed to get statistics for the non-existing agent %s", addrPort)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commState, details := agents.checkAgentCommState(stats, req, err)
	switch commState {
	case CommErrorNew:
		log.WithFields(log.Fields{
			"agent":     addrPort,
			"stats URL": statsURL,
		}).Warnf("Failed to send the named stats command: %s", req.NamedStatsRequest)
		agents.eventCenter.AddErrorEvent("communication with Stork agent on {machine} to query for named stats failed", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication with Stork agent on {machine} to query for named stats succeeded", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithFields(log.Fields{
			"agent":     addrPort,
			"stats URL": statsURL,
		}).Warnf("Failed to send the named stats command to the Stork agent: %s; agent is still not responding", req.NamedStatsRequest)

	default:
		// Communication with the agent was ok and is still ok.
	}

	response, ok := resp.(*agentapi.ForwardToNamedStatsRsp)
	if !ok || response == nil {
		return errors.Errorf("wrong response when querying stats from named via agent %s", addrPort)
	}

	bind9Stats := stats.GetBind9Stats()

	// Check the result of the communication between the Stork agent and named by
	// examining the returned status code.
	commState, details = agents.checkBind9CommState(bind9Stats, dbmodel.AccessPointStatistics, response.NamedStatsResponse)
	switch commState {
	case CommErrorNew:
		log.WithFields(log.Fields{
			"agent":     addrPort,
			"stats URL": statsURL,
		}).Warnf("Failed to forward the stats command %s from the Stork agent to named", req.NamedStatsRequest)
		agents.eventCenter.AddErrorEvent("communication between the Stork agent on {machine} and {daemon} to query named stats failed", daemon.GetMachineTag(), daemon, dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication between the Stork agent on {machine} and {daemon} to query named stats succeeded", daemon.GetMachineTag(), daemon, dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"rndc":  statsURL,
		}).Warnf("Failed to forward the stats command %s from the Stork agent to named; it is still returning an error", req.NamedStatsRequest)

	default:
		// Communication between the Stork agent and named was ok and is still ok.
	}

	// Stork agent returned an error.
	if err != nil {
		return errors.Wrapf(err, "failed to query stats from named via agent %s", addrPort)
	}

	// Communication with the Stork agent was ok, but named returned an error.
	if commState != CommErrorReset && commState != CommErrorNone {
		err = errors.Errorf("error communicating between Stork agent %s and named to query named stats: %s", addrPort, details)
		return err
	}

	statsResp := response.NamedStatsResponse

	// named responded but the response may contain an error status.
	if statsResp.Status.Code != agentapi.Status_OK {
		err = errors.New(statsResp.Status.Message)
	}

	// If status was ok, let's try to parse the response from the on-wire format.
	if err == nil {
		err = UnmarshalNamedStatsResponse(statsResp.Response, statsOutput)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse named statistics response from %s, response was: %s", statsURL, statsResp)
		} else {
			// No error parsing the response, so let's return.
			bind9Stats.ResetErrorCount(dbmodel.AccessPointStatistics)
			return nil
		}
	}

	// We've got some errors that have to be recorded in stats.
	if bind9Stats.IncreaseErrorCount(dbmodel.AccessPointStatistics) > 1 {
		err = errors.Errorf("failed to send named stats command via the agent %s, BIND 9 is still failing",
			daemon.GetMachineTag().GetAddress())
		return err
	}

	// This is apparently the first error like this. Let's log it.
	log.WithFields(log.Fields{
		"agent":     addrPort,
		"stats URL": statsURL,
	}).Warnf("named returned an error status to the stats query command: %s", req.NamedStatsRequest)

	return err
}

// Result of sending Kea commands to Kea.
type KeaCmdsResult struct {
	Error      error
	CmdsErrors []error
}

// Returns first error found in the KeaCmdsResult structure or nil if no
// error has been found.
func (result *KeaCmdsResult) GetFirstError() error {
	switch {
	case result == nil:
		return nil
	case result.Error != nil:
		return result.Error
	default:
		for _, err := range result.CmdsErrors {
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Forwards a Kea command via the Stork Agent and Kea Control Agent and then
// parses the response. It accepts a slice of commands that are aggregated
// in a single message to the Stork agent. The agent then sends them sequentially
// to the Kea servers via the control agent. This function tracks errors at several
// encapsulation level. First, it tracks the errors in sending the message to
// the Stork agent. Then, it tracks the errors reported by the Stork agent upon
// reception of this message. Next, it tracks the errors in communication between
// the Stork agent and Kea control agent. Finally, it tracks the errors reported
// by the daemons behind the control agent. Any new errors trigger appropriate
// events. If any of the existing errors go away in this communication, the
// warning events are triggered to indicate that the problem has gone away.
// The received responses are unmarshalled into the respective parameters at
// the end of the parameter list. The returned structure holds aggregated errors
// reported at different levels.
func (agents *connectedAgentsImpl) ForwardToKeaOverHTTP(ctx context.Context, daemon ControlledDaemon, commands []keactrl.SerializableCommand, cmdResponses ...any) (*KeaCmdsResult, error) {
	agentAddress := daemon.GetMachineTag().GetAddress()
	agentPort := daemon.GetMachineTag().GetAgentPort()
	agentURL := net.JoinHostPort(agentAddress, strconv.FormatInt(agentPort, 10))

	controlAddress, controlPort, _, controlUseSecureProtocol, err := daemon.GetControlAccessPoint()
	if err != nil {
		log.WithFields(log.Fields{
			"address": controlAddress,
			"port":    controlPort,
		}).Warnf("No Kea control access point found for daemon %s on machine %d", daemon.GetName(), daemon.GetMachineTag().GetID())
		return nil, err
	}
	controlAccessPointURL := storkutil.HostWithPortURL(controlAddress, controlPort, controlUseSecureProtocol)

	// Prepare the on-wire representation of the commands.
	req := &agentapi.ForwardToKeaOverHTTPReq{
		Url: controlAccessPointURL,
	}
	for _, cmd := range commands {
		// Verify the command is directed to the correct daemon.
		daemons := cmd.GetDaemonsList()
		if len(daemons) != 1 {
			return nil, errors.Errorf("expected a single daemon in the command %s, got %d", cmd.GetCommand(), len(daemons))
		} else if daemons[0] != daemon.GetName() {
			return nil, errors.Errorf("expected daemon %s in the command %s, got %s", daemon.GetName(), cmd.GetCommand(), daemons[0])
		}

		req.KeaRequests = append(req.KeaRequests, &agentapi.KeaRequest{
			Request: cmd.Marshal(),
		})
	}
	// Send the commands to the Stork Agent and get the response.
	resp, err := agents.sendAndRecvViaQueue(agentURL, req)

	// Check the communication issues with the Stork agent.
	stats := agents.getConnectedAgentStats(daemon.GetMachineTag().GetAddress(), daemon.GetMachineTag().GetAgentPort())
	if stats == nil {
		return nil, errors.Errorf("failed to get statistics for the non-existing agent %s", agentURL)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commState, details := agents.checkAgentCommState(stats, req, err)
	switch commState {
	case CommErrorNew:
		log.WithField("agent", agentURL).Warnf("Failed to send %d Kea command(s)", len(commands))
		agents.eventCenter.AddErrorEvent("communication with Stork agent on {machine} to forward Kea command failed", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication with Stork agent on {machine} to forward Kea command succeeded", daemon.GetMachineTag(), dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithField("agent", agentURL).Warnf("Failed to send %d Kea command(s) to the Stork agent; agent is still not responding", len(commands))

	default:
		// Communication with the agent was ok and is still ok.
	}

	// Stork agent returned an error.
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to send Kea commands via Stork agent %s", agentURL)
	}

	// Communication with the Stork agent was ok, but there was an error communicating
	// with a Kea agent. This is rather rare.
	if commState != CommErrorReset && commState != CommErrorNone {
		err = errors.Errorf("error communicating between Stork agent %s and Kea to send commands: %s", agentURL, details)
		return nil, err
	}

	response, ok := resp.(*agentapi.ForwardToKeaOverHTTPRsp)
	if !ok || response == nil {
		return nil, errors.Errorf("wrong response to a Kea command from agent %s", agentURL)
	}

	// We will aggregate the results from various communication levels in this structure.
	result := &KeaCmdsResult{}
	if response.Status.Code != agentapi.Status_OK {
		result.Error = errors.New(response.Status.Message)
	}

	// Check the communication issues with the Kea daemons. For each supported daemon we
	// get the current state of the communication with this daemon and optionally an
	// error message.
	keaCommState := agents.checkKeaCommState(stats.GetKeaStats(), commands, response)

	// Save Control Agent Errors.
	result.CmdsErrors = keaCommState.getErrors(daemon.GetName())
	state := keaCommState.getState(daemon.GetName())

	// Generate events for the Kea Control Agent.
	switch state {
	case CommErrorNew:
		// The connection was ok but now it is broken.
		log.WithFields(log.Fields{
			"agent":  agentURL,
			"daemon": daemon.GetName(),
		}).Warnf("Failed to forward Kea command to Kea daemon")
		agents.eventCenter.AddErrorEvent("forwarding Kea command to {daemon} on {machine} failed", daemon, daemon.GetMachineTag(), dbmodel.SSEConnectivity, keaCommState.getErrors(daemon.GetName()))
	case CommErrorReset:
		// The connection was broken but now is ok.
		agents.eventCenter.AddWarningEvent("forwarding Kea command to {daemon} on {machine} succeeded", daemon, daemon.GetMachineTag(), dbmodel.SSEConnectivity)
	case CommErrorContinued, CommErrorNone:
		// The connection was ok and is still ok.
		// No event is generated in this case.
	}

	// Get all responses from the Kea server.
	for idx, rsp := range response.GetKeaResponses() {
		cmdResp := cmdResponses[idx]
		// Try to parse the json response from the on-wire format.
		err = keactrl.UnmarshalResponseList(commands[idx], rsp.Response, cmdResp)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse Kea response from %s, response was: %s", controlAccessPointURL, rsp)
			// The sufficient number of elements should have been already allocated but
			// let's make sure.
			if len(result.CmdsErrors) > idx {
				result.CmdsErrors[idx] = err
			} else {
				result.CmdsErrors = append(result.CmdsErrors, err)
			}
			// Failure to parse the response.
			if state != CommErrorNew && state != CommErrorContinued {
				stats.GetKeaStats().IncreaseErrorCount(daemon.GetName())
			}
		}
	}

	// Everything was fine, so return no error.
	return result, nil
}

// Returns the PowerDNS server information.
func (agents *connectedAgentsImpl) GetPowerDNSServerInfo(ctx context.Context, daemon ControlledDaemon) (*pdnsdata.ServerInfo, error) {
	addrPort := net.JoinHostPort(daemon.GetMachineTag().GetAddress(), strconv.FormatInt(daemon.GetMachineTag().GetAgentPort(), 10))

	address, port, _, _, err := daemon.GetControlAccessPoint()
	if err != nil {
		return nil, err
	}
	req := &agentapi.GetPowerDNSServerInfoReq{
		WebserverAddress: address,
		WebserverPort:    port,
	}
	agentResponse, err := agents.sendAndRecvViaQueue(addrPort, req)
	if err != nil {
		return nil, err
	}
	response, ok := agentResponse.(*agentapi.GetPowerDNSServerInfoRsp)
	if !ok || response == nil {
		return nil, errors.Errorf("wrong response to getting PowerDNS server info from the Stork agent %s", addrPort)
	}
	serverInfo := &pdnsdata.ServerInfo{
		Type:             response.Type,
		ID:               response.Id,
		DaemonType:       response.DaemonType,
		Version:          response.Version,
		URL:              response.Url,
		ConfigURL:        response.ConfigURL,
		ZonesURL:         response.ZonesURL,
		AutoprimariesURL: response.AutoprimariesURL,
		Uptime:           response.Uptime,
	}
	return serverInfo, nil
}

// Get the tail of the remote text file.
func (agents *connectedAgentsImpl) TailTextFile(ctx context.Context, machine dbmodel.MachineTag, path string, offset int64) ([]string, error) {
	addrPort := net.JoinHostPort(machine.GetAddress(), strconv.FormatInt(machine.GetAgentPort(), 10))

	// Get the path to the file and the (seek) info indicating the location
	// from which the tail should be fetched.
	req := &agentapi.TailTextFileReq{
		Path:   path,
		Offset: offset,
	}

	// Send the request via queue.
	agentResponse, err := agents.sendAndRecvViaQueue(addrPort, req)

	stats := agents.getConnectedAgentStats(machine.GetAddress(), machine.GetAgentPort())
	if stats == nil {
		return nil, errors.Errorf("failed to get statistics for the non-existing agent %s", addrPort)
	}

	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	// Check connectivity with the Stork agent by examining the returned error.
	commIssue, details := agents.checkAgentCommState(stats, req, err)
	switch commIssue {
	case CommErrorNew:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"file":  path,
		}).Warn("Failed to tail the text file via the Stork agent", path)
		agents.eventCenter.AddErrorEvent("communication with Stork agent on {machine} to tail the text file failed", machine, dbmodel.SSEConnectivity, details)

	case CommErrorReset:
		agents.eventCenter.AddWarningEvent("communication with Stork agent on {machine} to tail the text file succeeded", machine, dbmodel.SSEConnectivity, details)

	case CommErrorContinued:
		log.WithFields(log.Fields{
			"agent": addrPort,
			"file":  path,
		}).Warn("Failed to tail the text file via the Stork agent; the agent is still not responding", path)
	default:
		// Communication with the agent was ok and is still ok.
	}

	// If there was an error in communication with the agent, there is no need
	// to check the response because it is probably nil anyway. Return an error.
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch text file contents: %s", path)
	}

	response, ok := agentResponse.(*agentapi.TailTextFileRsp)
	if !ok || response == nil {
		return nil, errors.Errorf("wrong response to tailing the text file from the Stork agent %s", addrPort)
	}

	// Check the status code.
	if response.Status.Code != agentapi.Status_OK {
		return nil, errors.New(response.Status.Message)
	}

	// All ok.
	return response.Lines, nil
}

// Receive DNS zones over the stream from a selected agent's zone inventory.
// It returns an iterator with a pointer to zone and error. The iterator ends
// when an error occurs. Receiving the zones is not cancellable at the moment.
func (agents *connectedAgentsImpl) ReceiveZones(ctx context.Context, daemon ControlledDaemon, filter *bind9stats.ZoneFilter) iter.Seq2[*bind9stats.ExtendedZone, error] {
	return func(yield func(*bind9stats.ExtendedZone, error) bool) {
		// Get control access point for the specified daemon. It will be sent
		// in the request to the agent, so the agent can identify correct
		// zone inventory.
		ctrlAddress, ctrlPort, _, _, err := daemon.GetControlAccessPoint()
		if err != nil {
			_ = yield(nil, err)
			return
		}
		// Get the agent's state. It holds the connection with the agent.
		agentAddressPort := net.JoinHostPort(daemon.GetMachineTag().GetAddress(), strconv.FormatInt(daemon.GetMachineTag().GetAgentPort(), 10))
		agent, err := agents.getConnectedAgent(agentAddressPort)
		if err != nil {
			_ = yield(nil, err)
			return
		}
		// Start creating the request.
		request := &agentapi.ReceiveZonesReq{
			ControlAddress: ctrlAddress,
			ControlPort:    ctrlPort,
		}
		// Set filtering rules, if required.
		if filter != nil {
			if filter.View != nil {
				request.ViewName = *filter.View
			}
			if filter.Limit != nil {
				request.Limit = int64(*filter.Limit)
			}
		}
		// This is the same pattern we're using in the manager.go. The connection is
		// cached so it is possible that it gets terminated or broken at some point.
		// By trying the actual operation and retrying on failure we should be able
		// to recover. There may be other ways to achieve recovery (e.g., getting
		// the connection state before attempting the call). However, it is hard to
		// say how reliable they are. This approach worked well for several years so
		// it should be fine to continue using it.
		var stream grpc.ServerStreamingClient[agentapi.Zone]
		if stream, err = agent.connector.createClient().ReceiveZones(ctx, request); err != nil {
			if err = agent.connector.connect(); err == nil {
				stream, err = agent.connector.createClient().ReceiveZones(ctx, request)
			}
		}
		if err != nil {
			// The zone inventory may signal errors indicating that it is
			// unable to return the zones because it is in a wrong state.
			// The server should interpret these errors and formulate hints
			// to the user that some administrative actions may be required.
			s := status.Convert(err)
			for _, d := range s.Details() {
				if info, ok := d.(*errdetails.ErrorInfo); ok {
					switch info.Reason {
					case "ZONE_INVENTORY_NOT_INITED":
						// Zone inventory hasn't been initialized.
						_ = yield(nil, NewZoneInventoryNotInitedError(agentAddressPort))
						return
					case "ZONE_INVENTORY_BUSY":
						// Zone inventory is busy. Retrying later may help.
						_ = yield(nil, NewZoneInventoryBusyError(agentAddressPort))
						return
					default:
						_ = yield(nil, err)
						return
					}
				}
			}
			// Other error.
			_ = yield(nil, err)
			return
		}

		for {
			// Start receiving zones.
			receivedZone, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					_ = yield(nil, err)
				}
				return
			}
			zone := &bind9stats.ExtendedZone{
				Zone: bind9stats.Zone{
					ZoneName: receivedZone.GetName(),
					Class:    receivedZone.GetClass(),
					Serial:   receivedZone.GetSerial(),
					Type:     receivedZone.GetType(),
					Loaded:   time.Unix(receivedZone.GetLoaded(), 0).UTC(),
				},
				RPZ:            receivedZone.GetRpz(),
				ViewName:       receivedZone.View,
				TotalZoneCount: receivedZone.TotalZoneCount,
			}
			if !yield(zone, nil) {
				// Stop if the caller no longer iterates over the zones.
				return
			}
		}
	}
}

// Makes a request to the agent to perform a zone transfer for a specified view
// and zone. It returns an iterator to the received RRs and error.
func (agents *connectedAgentsImpl) ReceiveZoneRRs(ctx context.Context, daemon ControlledDaemon, zoneName string, viewName string) iter.Seq2[[]*dnsconfig.RR, error] {
	return func(yield func([]*dnsconfig.RR, error) bool) {
		// Get control access point for the specified daemon. It will be sent
		// in the request to the agent, so the agent can identify correct
		// zone inventory.
		ctrlAddress, ctrlPort, _, _, err := daemon.GetControlAccessPoint()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		request := &agentapi.ReceiveZoneRRsReq{
			ControlAddress: ctrlAddress,
			ControlPort:    ctrlPort,
			ZoneName:       zoneName,
			ViewName:       viewName,
		}

		// Get the agent's state. It holds the connection with the agent.
		agentAddressPort := net.JoinHostPort(daemon.GetMachineTag().GetAddress(), strconv.FormatInt(daemon.GetMachineTag().GetAgentPort(), 10))
		agent, err := agents.getConnectedAgent(agentAddressPort)
		if err != nil {
			_ = yield(nil, err)
			return
		}

		// This is the same pattern we're using in the manager.go. The connection is
		// cached so it is possible that it gets terminated or broken at some point.
		// By trying the actual operation and retrying on failure we should be able
		// to recover. There may be other ways to achieve recovery (e.g., getting
		// the connection state before attempting the call). However, it is hard to
		// say how reliable they are. This approach worked well for several years so
		// it should be fine to continue using it.
		var stream grpc.ServerStreamingClient[agentapi.ReceiveZoneRRsRsp]
		if stream, err = agent.connector.createClient().ReceiveZoneRRs(ctx, request); err != nil {
			if err = agent.connector.connect(); err == nil {
				stream, err = agent.connector.createClient().ReceiveZoneRRs(ctx, request)
			}
		}
		if err != nil {
			// The zone inventory may signal errors indicating that it is
			// unable to return the RRs because it is in a wrong state.
			s := status.Convert(err)
			for _, d := range s.Details() {
				if info, ok := d.(*errdetails.ErrorInfo); ok {
					switch info.Reason {
					case "ZONE_INVENTORY_NOT_INITED":
						// Zone inventory hasn't been initialized.
						_ = yield(nil, NewZoneInventoryNotInitedError(agentAddressPort))
						return
					case "ZONE_INVENTORY_BUSY":
						// Zone inventory is busy. Retrying later may help.
						_ = yield(nil, NewZoneInventoryBusyError(agentAddressPort))
						return
					default:
						_ = yield(nil, err)
						return
					}
				}
			}
			// Other error.
			_ = yield(nil, err)
			return
		}
		for {
			// Receive the zone contents from the agent.
			receivedRRs, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					// Report the error excluding the EOF which is just the end of the stream.
					_ = yield(nil, err)
				}
				return
			}
			// Convert the received RRs to the format convenient for further processing
			// on the server side.
			rrs := make([]*dnsconfig.RR, len(receivedRRs.Rrs))
			for i, rr := range receivedRRs.Rrs {
				rrs[i], err = dnsconfig.NewRR(rr)
				if err != nil {
					// This is unlikely but we need to handle it.
					_ = yield(nil, err)
					return
				}
			}
			if !yield(rrs, nil) {
				// Stop if the caller no longer iterates over the RRs.
				return
			}
		}
	}
}
