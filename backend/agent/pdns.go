package agent

import (
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	pdnsconfig "isc.org/stork/daemoncfg/pdns"
)

var (
	_ Daemon           = (*PDNSDaemon)(nil)
	_ DNSDaemon        = (*PDNSDaemon)(nil)
	_ pdnsConfigParser = (*pdnsconfig.Parser)(nil)

	// Pattern for detecting PowerDNS process.
	pdnsPattern = regexp.MustCompile(`(.*?)pdns_server(\s+.*)?`)
)

// An interface for parsing PowerDNS configuration files.
// It is mocked in the tests.
type pdnsConfigParser interface {
	ParseFile(path string) (*pdnsconfig.Config, error)
}

// PDNSDaemon implements the Daemon interface for PowerDNS.
type PDNSDaemon struct {
	daemon
	zoneInventory *zoneInventory
}

// Lifecycle method called once when the daemon is added to the monitor.
// It starts the zone inventory background tasks.
func (pa *PDNSDaemon) Bootstrap() error {
	return nil
}

// Lifecycle method called periodically by the monitor.
// It populates the zone inventory.
func (ba *PDNSDaemon) Evaluate(AgentManager) error {
	zoneInventory := ba.GetZoneInventory()
	if zoneInventory == nil || zoneInventory.getCurrentState().isReady() {
		return nil
	}
	var busyError *zoneInventoryBusyError
	if _, err := zoneInventory.populate(false); err != nil {
		switch {
		case errors.As(err, &busyError):
			// Inventory creation is in progress. This is not an error.
			return nil
		default:
			return errors.WithMessage(err, "Failed to populate DNS zones inventory")
		}
	}
	return nil
}

// Lifecycle method called once when the daemon is removed from the monitor.
// Waits for the zone inventory to complete background tasks.
func (pa *PDNSDaemon) Cleanup() error {
	if pa.zoneInventory != nil {
		pa.zoneInventory.stop()
	}
	return nil
}

// Returns the zone inventory.
func (pa *PDNSDaemon) GetZoneInventory() *zoneInventory {
	return pa.zoneInventory
}

// Detect the PowerDNS daemon by parsing the named process command line.
// If the path to the configuration file is relative and chroot directory is
// not specified, the path is resolved against the current working directory of
// the process. If the chroot directory is specified, the path is resolved
// against it.
//
// The function reads the configuration file and extracts webserver address,
// port, and API key (if configured).
//
// It returns the PowerDNS daemon instance or an error if the PowerDNS is not
// recognized or any error occurs.
func detectPowerDNSDaemon(p supportedProcess, parser pdnsConfigParser) (Daemon, error) {
	cmdline, err := p.getCmdline()
	if err != nil {
		return nil, err
	}
	cwd, err := p.getCwd()
	if err != nil {
		log.WithError(err).Warn("Failed to get PowerDNS process current working directory")
	}
	match := pdnsPattern.FindStringSubmatch(cmdline)
	if match == nil {
		return nil, errors.Errorf("failed to find pdns_server in cmdline: %s", cmdline)
	}

	configDir := ""
	configName := "pdns.conf"
	rootPrefix := ""
	if len(match) >= 3 {
		// The command line contains parameters. Check if they specify config
		// directory or config name.
		pdnsParams := match[2]
		paramsSlice := strings.Fields(pdnsParams)
		for _, param := range paramsSlice {
			key, value, found := strings.Cut(param, "=")
			if !found {
				continue
			}
			switch key {
			case "--chroot":
				rootPrefix = strings.TrimRight(value, "/")
			case "--config-dir":
				configDir = value
			case "--config-name":
				configName = value
			}
		}
	}
	if !path.IsAbs(configDir) {
		// PowerDNS configuration is typically stored in /etc/powerdns.
		configDir = path.Join("/etc/powerdns", configDir)
	}
	configPath := path.Join(configDir, configName)
	if rootPrefix != "" {
		configPath = path.Join(rootPrefix, configPath)
	}
	if !path.IsAbs(configPath) {
		// If path to config is not absolute then join it with current working directory.
		configPath = path.Join(cwd, configPath)
	}
	// Parse the configuration file.
	parsedConfig, err := parser.ParseFile(configPath)
	if err != nil {
		return nil, err
	}
	// Get the webserver address and port.
	webserverAddress, webserverPort, enabled := parsedConfig.GetWebserverConfig()
	if !enabled {
		return nil, errors.Errorf("API or webserver disabled in %s", configPath)
	}
	// Get the API key. It is mandatory.
	key := parsedConfig.GetString("api-key")
	if key == nil {
		return nil, errors.Errorf("api-key not found in %s", configPath)
	}
	// Create webserver client.
	client := newPDNSClient()
	// For larger deployments, it may take several minutes to retrieve the
	// zones from the DNS server.
	client.SetRequestTimeout(time.Minute * 3)

	// Create the zone inventory.
	inventory := newZoneInventory(newZoneInventoryStorageMemory(), parsedConfig, client, *webserverAddress, *webserverPort)

	// Create the PowerDNS app.
	daemon := &PDNSDaemon{
		daemon: daemon{
			Name: DaemonNamePDNS,
			AccessPoints: []AccessPoint{
				{
					Type:    AccessPointControl,
					Address: *webserverAddress,
					Port:    *webserverPort,
					Key:     *key,
				},
			},
		},
		zoneInventory: inventory,
	}
	return daemon, nil
}
