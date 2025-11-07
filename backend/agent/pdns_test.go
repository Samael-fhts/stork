package agent

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	pdnsconfig "isc.org/stork/daemoncfg/pdns"
	"isc.org/stork/daemonctrl/constants/daemonname"
)

//go:generate mockgen -package=agent -destination=pdnsmock_test.go -mock_names=pdnsConfigParser=MockPDNSConfigParser isc.org/stork/agent pdnsConfigParser

// Test that the daemon structure can be accessed.
func TestPowerDNSDaemonGetBaseDaemon(t *testing.T) {
	daemon := &PDNSDaemon{
		dnsDaemonImpl: dnsDaemonImpl{
			daemon: daemon{
				Name: daemonname.PDNS,
			},
		},
	}
	require.Equal(t, daemonname.PDNS, daemon.GetName())
}

// Test that the refreshing state of the PowerDNS daemon doesn't return any errors.
func TestPowerDNSDaemonRefreshState(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentManager := NewMockAgentManager(ctrl)

	zoneInventory := NewMockZoneInventory(ctrl)
	zoneInventory.EXPECT().populate(gomock.Any()).Return(nil, nil)
	zoneInventory.EXPECT().getCurrentState().Return(&zoneInventoryState{})

	daemon := &PDNSDaemon{dnsDaemonImpl: dnsDaemonImpl{zoneInventory: zoneInventory}}

	// Act
	err := daemon.RefreshState(t.Context(), agentManager)

	// Assert
	require.NoError(t, err)
}

// Test that cleanup doesn't panic when zone inventory is nil.
func TestPowerDNSDaemonCleanupNilZoneInventory(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	zoneInventory := NewMockZoneInventory(ctrl)
	zoneInventory.EXPECT().stop()

	daemon := &PDNSDaemon{dnsDaemonImpl: dnsDaemonImpl{zoneInventory: zoneInventory}}

	// Act & Assert
	require.NotPanics(t, func() {
		err := daemon.Cleanup()
		require.NoError(t, err)
	})
}

// Test that the zone inventory can be accessed.
func TestPowerDNSDaemonGetZoneInventory(t *testing.T) {
	daemon := &PDNSDaemon{dnsDaemonImpl: dnsDaemonImpl{
		zoneInventory: &zoneInventoryImpl{},
	}}
	require.Equal(t, daemon.zoneInventory, daemon.getZoneInventory())
}

// Test successfully detecting PowerDNS daemon.
func TestDetectPowerDNSDaemon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc", nil)
	process.EXPECT().getCwd().Return("/etc", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(defaultPDNSConfig))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)

	pdnsDaemon := daemon.(*PDNSDaemon)
	require.NotNil(t, pdnsDaemon.getZoneInventory())
}

// Test that the PowerDNS is correctly detected when no parameters are
// specified. It should use the default config directory.
func TestDetectPowerDNSDaemonNoConfigDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server", nil)
	process.EXPECT().getCwd().Return("/etc", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/powerdns/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(defaultPDNSConfig))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)

	pdnsDaemon := daemon.(*PDNSDaemon)
	require.NotNil(t, pdnsDaemon.getZoneInventory())
}

// Test that an error is returned when getting a process command line fails.
func TestDetectPowerDNSDaemonCmdLineError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("", errors.New("test error"))

	daemon, err := detectPowerDNSDaemon(process, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "test error")
	require.Nil(t, daemon)
}

// Test that an error is returned when getting a process current working directory fails.
func TestDetectPowerDNSDaemonCwdError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-name=pdns.conf", nil)
	process.EXPECT().getCwd().Return("", errors.New("test error"))

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/powerdns/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(defaultPDNSConfig))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)
	require.NotNil(t, daemon.(*PDNSDaemon).getZoneInventory())
}

// Test that the daemon can be detected when the chroot directory is used.
func TestDetectPowerDNSDaemonChroot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --chroot=/chroot --config-dir=/etc --config-name=pdns.conf", nil)
	process.EXPECT().getCwd().Return("/chroot", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/chroot/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(defaultPDNSConfig))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)
	require.NotNil(t, daemon.(*PDNSDaemon).getZoneInventory())
}

// Test that custom config directory and name can be specified while detecting
// PowerDNS daemon.
func TestDetectPowerDNSDaemonConfigDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/opt/etc --config-name=server.conf", nil)
	process.EXPECT().getCwd().Return("/chroot", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/opt/etc/server.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(defaultPDNSConfig))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)
	require.NotNil(t, daemon.(*PDNSDaemon).getZoneInventory())
}

// Test that an error is returned when parsing the configuration file fails.
func TestDetectPowerDNSDaemonParseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc --config-name=pdns.conf", nil)
	process.EXPECT().getCwd().Return("/etc", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").Return(nil, errors.New("test error"))

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.Error(t, err)
	require.ErrorContains(t, err, "test error")
	require.Nil(t, daemon)
}

// Test that default webserver address and port are used when not specified
// in the configuration file.
func TestDetectPowerDNSDaemonDefaultWebserver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc", nil)
	process.EXPECT().getCwd().Return("", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(`
			api=yes
			webserver=yes
			api-key=stork
		`))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.NoError(t, err)
	require.NotNil(t, daemon)

	require.IsType(t, &PDNSDaemon{}, daemon)
	require.Equal(t, daemonname.PDNS, daemon.GetName())
	require.Len(t, daemon.GetAccessPoints(), 1)
	require.Equal(t, AccessPointControl, daemon.GetAccessPoints()[0].Type)
	require.EqualValues(t, 8081, daemon.GetAccessPoints()[0].Port)
	require.Equal(t, "127.0.0.1", daemon.GetAccessPoints()[0].Address)
	require.Equal(t, "stork", daemon.GetAccessPoints()[0].Key)
	require.NotNil(t, daemon.(*PDNSDaemon).getZoneInventory())
}

// Test that an error is returned when the API key is not specified in the
// configuration file.
func TestDetectPowerDNSDaemonNoAPIKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc", nil)
	process.EXPECT().getCwd().Return("", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(`
			api
			webserver=yes
		`))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.Error(t, err)
	require.ErrorContains(t, err, "api-key not found in /etc/pdns.conf")
	require.Nil(t, daemon)
}

// Test that an error is returned when the webserver is disabled in the
// configuration file.
func TestDetectPowerDNSDaemonNoWebserver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc", nil)
	process.EXPECT().getCwd().Return("", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(`
			api
			webserver=no
		`))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.Error(t, err)
	require.ErrorContains(t, err, "webserver disabled in /etc/pdns.conf")
	require.Nil(t, daemon)
}

// Test that an error is returned when the API is disabled in the
// configuration file.
func TestDetectPowerDNSDaemonNoAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	process := NewMockSupportedProcess(ctrl)
	process.EXPECT().getCmdline().Return("/dir/pdns_server --config-dir=/etc", nil)
	process.EXPECT().getCwd().Return("", nil)

	parser := NewMockPDNSConfigParser(ctrl)
	parser.EXPECT().ParseFile("/etc/pdns.conf").DoAndReturn(func(path string) (*pdnsconfig.Config, error) {
		return pdnsconfig.NewParser().Parse(strings.NewReader(`
			webserver=yes
		`))
	})

	daemon, err := detectPowerDNSDaemon(process, parser)
	require.Error(t, err)
	require.ErrorContains(t, err, "API or webserver disabled in /etc/pdns.conf")
	require.Nil(t, daemon)
}
