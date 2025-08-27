package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	keaconfig "isc.org/stork/appcfg/kea"
	keactrl "isc.org/stork/appctrl/kea"
	storkutil "isc.org/stork/util"
)

var _ Daemon = (*KeaDaemon)(nil)

// It holds common and Kea specific runtime information.
type KeaDaemon struct {
	daemon
	HTTPClient *httpClient // to communicate with Kea Control Agent
}

// Returns the control access point of the Kea daemon.
func (d *KeaDaemon) getControlAccessPoint() *AccessPoint {
	for _, ap := range d.AccessPoints {
		if ap.Type == AccessPointControl {
			return &ap
		}
	}
	return nil
}

// Sends a command to Kea and returns a response.
func (d *KeaDaemon) sendCommand(command *keactrl.Command, responses interface{}) error {
	// Get the textual representation of the command.
	request := command.Marshal()

	// Send the command to Kea CA.
	body, err := d.sendCommandRaw([]byte(request))
	if err != nil {
		return err
	}

	// Parse the response.
	err = keactrl.UnmarshalResponseList(command, body, responses)
	if err != nil {
		return errors.WithMessage(err, "failed to parse Kea response body received")
	}
	return nil
}

// Sends a serialized command to Kea and returns a serialized response.
func (d *KeaDaemon) sendCommandRaw(command []byte) ([]byte, error) {
	accessPoint := d.getControlAccessPoint()
	if accessPoint == nil {
		return nil, errors.New("no control access point found")
	}

	caURL := storkutil.HostWithPortURL(
		accessPoint.Address,
		accessPoint.Port,
		accessPoint.UseSecureProtocol,
	)

	// Send the command to the Kea server.
	response, err := d.HTTPClient.Call(caURL, bytes.NewBuffer(command))
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to send command to Kea: %s", caURL)
	}

	// Kea returned a non-success status code.
	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received non-success status code %d from Kea, with status text: %s; url: %s", response.StatusCode, response.Status, caURL)
	}

	// Read the response.
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to read Kea response body received from %s", caURL)
	}

	return body, nil
}

// Collect the list of log files which can be viewed by the Stork user
// from the UI. The response variable holds the pointer to the
// response to the config-get command returned by one of the Kea
// daemons. If this response contains loggers' configuration the log
// files are extracted from it and returned. This function is intended
// to be called by the functions which intercept config-get commands
// sent periodically by the server to the agents and by the
// DetectAllowedLogs when the agent is started.
func collectKeaAllowedLogs(response *keactrl.Response) ([]string, error) {
	if err := response.GetError(); err != nil {
		err = errors.WithMessage(err, "skipped refreshing viewable log files because config-get returned unsuccessful result")
		return nil, err
	}
	if response.Arguments == nil {
		err := errors.New("skipped refreshing viewable log files because config-get response has no arguments")
		return nil, err
	}
	cfg := keaconfig.NewConfigFromMap(response.Arguments)
	if cfg == nil {
		err := errors.New("skipped refreshing viewable log files because config-get response contains arguments which could not be parsed")
		return nil, err
	}

	loggers := cfg.GetLoggers()
	if len(loggers) == 0 {
		log.Info("No loggers found in the returned configuration while trying to refresh the viewable log files")
		return nil, nil
	}

	// Go over returned loggers and collect those found in the returned configuration.
	var paths []string
	for _, l := range loggers {
		for _, o := range l.GetAllOutputOptions() {
			// TODO: We could read the stdout and stderr too by reading
			// "/proc/<pid>/fd/1" and "/proc/<pid>/fd/2" symlinks.
			// It is also possible to read syslog.
			if o.Output != "stdout" && o.Output != "stderr" && !strings.HasPrefix(o.Output, "syslog") {
				paths = append(paths, o.Output)
			}
		}
	}
	return paths, nil
}

// Sends config-get command to the running Kea daemon to fetch logging
// configuration. The log files locations are stored in the logTailer instance
// of the agent as allowed for viewing. This function should be called when the
// agent has been started and the running Kea daemons have been detected.
func (d *KeaDaemon) detectAllowedLogs() ([]string, error) {
	// Prepare config-get command to be sent to Kea Control Agent.
	command := keactrl.NewCommandBase(keactrl.ConfigGet, keactrl.DaemonName(d.GetName()))
	// Send the command to Kea.
	responses := keactrl.ResponseList{}
	err := d.sendCommand(command, &responses)
	if err != nil {
		return nil, err
	}

	// There should be exactly one response received because we sent the command
	// to only one daemon.
	if len(responses) != 1 {
		return nil, errors.Errorf("invalid response received from Kea CA to config-get command sent to %s", d)
	}
	response := responses[0]

	// It does not make sense to proceed if the CA returned non-success status
	// because this response neither contains logging configuration nor
	// sockets configurations.
	if err := response.GetError(); err != nil {
		return nil, errors.WithMessagef(
			err, "unsuccessful response received from Kea CA to config-get command sent to %s", d,
		)
	}

	// Allow the log files used by the CA.
	paths, err := collectKeaAllowedLogs(&response)
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// Reads the Kea configuration file, resolves the includes, and parses the content.
func readKeaConfig(path string) (*keaconfig.Config, error) {
	text, err := storkutil.ReadFileWithIncludes(path)
	if err != nil {
		err = errors.WithMessage(err, "Cannot read Kea config file")
		return nil, err
	}

	config, err := keaconfig.NewConfig(text)
	if err != nil {
		err = errors.WithMessage(err, "Cannot parse Kea Control Agent config file")
		return nil, err
	}

	return config, err
}

// Detect the Kea daemon(s).
//
// The communication model with Kea changed significantly with the release of
// Kea 3.0. The Kea Control Agent is no longer required to establish connection
// with the Kea daemons (DHCP, DDNS, etc.). Instead, the daemons provide its
// own control channels. The Kea CA still exists and can be used to manage the
// daemons but it is deprecated and may be removed in future releases.
// The Kea daemons support two modes of control channel: HTTP-based (same as
// the Kea CA) and socket-based. In both cases, the expected data format is
// JSON.
//
// This function supports all Kea versions (prior and post 3.0) and all modes
// of control channel (HTTP- and socket-based).
//
// For Kea prior to 3.0, the function detects multiple daemons if CA daemon is
// passed, and no daemons if any other daemon is passed. It is because only Kea
// CA can contact other daemons in this Kea version. All daemons detected this
// way have the same control access point because they are connected via the
// CA.
// For Kea 3.0 and later, the function detects only the passed daemon because
// it expects the connection will be established directly with the daemon. Each
// daemon has its own control channel.
//
// The access points of the daemons are detected by reading the daemon
// configuration file. The function parses command line of the specified
// process. It looks for the configuration file path in the command line. If
// the path is relative, it is resolved against the current working directory
// of the process.
//
// It reads the configuration file and extracts its HTTP host, port,
// TLS configuration, basic authentication credentials. For Kea prior to 3.0,
// the function also reads the list of configured daemons and then sends the
// version-get command to each daemon to check if it is running.
//
// The version of the Kea daemon is recognized by calling its executable with
// the --version flag.
//
// The specified httpClientConfig is used to create a new HTTP client instance
// for the detected Kea app. The client inherits the the general HTTP client
// configuration from the Stork agent configuration and additionally sets the
// basic authentication credentials if they are provided in the Kea CA
// configuration. It picks the first credentials with the user name "stork" or
// starting with "stork." If there are no such credentials, it picks the first
// one. See @readClientCredentials for details.
//
// It returns the Kea app instance or an error if the Kea is not recognized or
// any error occurs.
func detectKeaDaemons(p supportedProcess, httpClientConfig HTTPClientConfig, commander storkutil.CommandExecutor) ([]Daemon, error) {
	// Extract the daemon name from the process.
	processName, err := p.getName()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get process name")
	}

	daemonName := convertProcessNameToDaemonName(processName)
	if daemonName == "" {
		return nil, errors.Errorf("unsupported Kea process: %s", processName)
	}

	// Extract the config path and the executable path from the command line.
	cmdline, err := p.getCmdline()
	if err != nil {
		return nil, err
	}
	cwd, err := p.getCwd()
	if err != nil {
		log.WithError(err).Warn("Cannot get Kea process current working directory")
	}

	pattern := regexp.MustCompile(fmt.Sprintf(`(.*?)%s\s+.*-c\s+(\S+)`, processName))

	match := pattern.FindStringSubmatch(cmdline)
	if match == nil {
		return nil, errors.Errorf("problem parsing Kea command line: %s", cmdline)
	}

	if len(match) < 3 {
		return nil, errors.Errorf("problem parsing Kea command line: %s", match[0])
	}

	// Check the version of the Kea binary. We need to differentiate between
	// Kea prior to 3.0 and Kea post 3.0.
	executablePath := match[1] + processName
	versionRaw, err := commander.Output(executablePath, "--version")
	if err != nil {
		return nil, errors.WithMessagef(err, "cannot get Kea version by executing %s --version", executablePath)
	}
	version, err := storkutil.ParseSemanticVersion(string(versionRaw))
	if err != nil {
		return nil, errors.WithMessagef(err, "cannot parse Kea version: %s", string(versionRaw))
	}
	shouldTunnelViaCA := version.LessThan(storkutil.SemanticVersion{Major: 3, Minor: 0, Patch: 0})
	if shouldTunnelViaCA && daemonName != DaemonNameCA {
		// For Kea prior to 3.0, only the CA daemon can connect to other daemons.
		// If the process is not CA, we cannot detect any daemons.
		return nil, nil
	}

	// Read the configuration file.
	configPath := match[2]

	if !strings.HasPrefix(configPath, "/") {
		// If path to config is not absolute then join it with CWD of Kea.
		configPath = path.Join(cwd, configPath)
	}

	config, err := readKeaConfig(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "invalid Kea Control Agent config")
	}

	controlSockets := config.GetListeningControlSockets()
	if len(controlSockets) == 0 {
		return nil, errors.New("no listening control sockets configured in Kea config")
	}
	controlSocket := controlSockets[0]

	// Credentials
	// Key is a user name that Stork uses to authenticate with Kea.
	var key string
	if controlSocket.Authentication != nil {
		allCredentials, err := readClientCredentials(controlSocket.Authentication)
		if err != nil {
			return nil, errors.WithMessage(err, "cannot read client credentials")
		}

		if len(allCredentials) > 0 {
			// Fall back to the first set of credentials.
			credentials := allCredentials[0]

			// Look for the credentials prefixed with "stork".
			for _, c := range allCredentials {
				if strings.HasPrefix(c.User, "stork") {
					credentials = c
					break
				}
			}

			httpClientConfig.BasicAuth = basicAuthCredentials(credentials)
			key = credentials.User
		}
	}

	accessPoints := []AccessPoint{
		{
			Type:              AccessPointControl,
			Address:           controlSocket.GetAddress(),
			Port:              controlSocket.GetPort(),
			UseSecureProtocol: controlSocket.UseSecureProtocol(),
			Key:               key,
		},
	}
	thisDaemon := &KeaDaemon{
		daemon: daemon{
			Name:         daemonName,
			Pid:          p.getPid(),
			AccessPoints: accessPoints,
		},
		HTTPClient: NewHTTPClient(httpClientConfig),
	}

	if shouldTunnelViaCA {
		// For Kea prior to 3.0, get the list of configured daemons.
		managementControlSockets := config.GetManagementControlSockets()
		managedDaemons := managementControlSockets.GetManagedDaemonNames()
	}

	return thisDaemon, nil
}

// Detects the active Kea daemons by sending the version-get command to each daemon.
// The non-nil list of active daemons is returned.
// Returns an error if the Kea CA is down but it doesn't throw an error if Kea
// daemons are down. In the latter case, the error is logged but only if the
// daemon was not already detected as inactive.
func detectKeaActiveDaemons(keaApp *KeaApp, previousActiveDaemons []string) (daemons []string, err error) {
	// Detect active daemons.
	// Send the version-get command to each daemon to check if it is running.
	command := keactrl.NewCommandBase(keactrl.VersionGet, keaApp.ConfiguredDaemons...)
	responses := keactrl.ResponseList{}
	err = keaApp.sendCommand(command, &responses)
	if err != nil {
		// The Kea CA seems to be down, so we cannot detect the active daemons.
		return nil, errors.WithMessage(err, "failed to send command to Kea Control Agent")
	}

	// Return non-nil list of active daemons to indicate that the detection was performed.
	daemons = []string{}
	for _, r := range responses {
		if err := r.GetError(); err != nil {
			// If it is a first detection, the daemon is newly inactive.
			// Otherwise, it depends on the previous state.
			isNewlyInactive := previousActiveDaemons == nil
			for _, ad := range previousActiveDaemons {
				if ad == r.GetDaemon() {
					// Daemon was previously active.
					isNewlyInactive = true
					break
				}
			}

			if isNewlyInactive {
				log.WithError(err).
					WithField("daemon", r.GetDaemon()).
					Errorf("Failed to communicate with Kea daemon")
			}
		} else {
			daemons = append(daemons, r.GetDaemon())
		}
	}

	return daemons, nil
}

type ClientCredentials struct {
	User     string
	Password string
}

// Reads the client credentials.
// Kea supports multiple ways of providing client credentials.
//
// 1. Username and password can be provided directly in the configuration file.
// 2. Username and password can be provided in separate files.
// 3. Username and password can be provided in a separate file delimited by a colon.
// 4. Username can be provided directly in the configuration file and the password in a separate file.
// 5. Username can be provided in the separate file and the password directly in the configuration file.
func readClientCredentials(authentication *keaconfig.Authentication) ([]ClientCredentials, error) {
	allCredentials := []ClientCredentials{}

	directory := "/"
	if authentication.Directory != nil {
		directory = *authentication.Directory
	}

	for _, client := range authentication.Clients {
		var credentials ClientCredentials

		// Read the user.
		switch {
		case client.User != nil:
			// The user provided as a string.
			credentials.User = *client.User
		case client.UserFile != nil:
			// The user is provided in a file.
			userPath := path.Join(directory, *client.UserFile)
			userRaw, err := os.ReadFile(userPath)
			if err != nil {
				return nil, errors.WithMessagef(err,
					"could not read the user file '%s'",
					userPath,
				)
			}
			credentials.User = strings.TrimSpace(string(userRaw))
		case client.PasswordFile != nil:
			// The user and password are provided in a single file.
			passwordPath := path.Join(directory, *client.PasswordFile)
			passwordRaw, err := os.ReadFile(passwordPath)
			if err != nil {
				return nil, errors.WithMessagef(err,
					"could not read the password file '%s'",
					passwordPath,
				)
			}
			parts := strings.Split(strings.TrimSpace(string(passwordRaw)), ":")
			if len(parts) != 2 {
				return nil, errors.Errorf(
					"invalid format of the password file '%s'",
					passwordPath,
				)
			}
			credentials.User = parts[0]
			credentials.Password = parts[1]
		default:
			// Missing user.
			return nil, errors.New(
				"invalid client credentials: neither user nor user-file provided",
			)
		}

		// Read the password.
		switch {
		case credentials.Password != "":
			// The password has been provided together with the user in
			// the password file.
		case client.Password != nil:
			// The password provided as a string.
			credentials.Password = *client.Password
		case client.PasswordFile != nil:
			// The password is provided in a file.
			passwordPath := path.Join(directory, *client.PasswordFile)
			passwordRaw, err := os.ReadFile(passwordPath)
			if err != nil {
				return nil, errors.WithMessagef(err,
					"could not read the password file '%s'",
					passwordPath,
				)
			}
			credentials.Password = strings.TrimSpace(string(passwordRaw))
		default:
			// Missing password.
			return nil, errors.New(
				"invalid client credentials - password or password-file is not provided",
			)
		}

		allCredentials = append(allCredentials, credentials)
	}
	return allCredentials, nil
}
