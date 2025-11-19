package dbmodel

import (
	"fmt"
	"hash/adler32"
	"testing"

	"github.com/stretchr/testify/require"
	"isc.org/stork/datamodel/daemonname"
	dbtest "isc.org/stork/server/database/test"
)

// Test that the virtual app can be derived from a daemon.
func TestGetVirtualApp(t *testing.T) {
	// Arrange
	machine := &Machine{Address: "foo"}
	daemon := NewDaemon(machine, daemonname.DHCPv4, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "bar",
			Port:    8000,
			Key:     "baz",
		},
	})

	// Act
	virtualApp := daemon.GetVirtualApp()

	// Assert
	expectedAppID := int64(adler32.Checksum([]byte("bar:8000")))
	expectedAppName := fmt.Sprintf("%s@%s%%%d", VirtualAppTypeKea, machine.Address, expectedAppID)

	require.Equal(t, expectedAppID, virtualApp.ID)
	require.Equal(t, expectedAppName, virtualApp.Name)
	require.Equal(t, VirtualAppTypeKea, virtualApp.Type)
}

// Test that daemons with the same access point have the same virtual app ID.
func TestGetVirtualAppForVariousDaemons(t *testing.T) {
	// Arrange
	machine := &Machine{Address: "foo"}
	daemon1 := NewDaemon(machine, daemonname.DHCPv4, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "bar",
			Port:    8000,
			Key:     "baz",
		},
	})
	daemon2 := NewDaemon(machine, daemonname.DHCPv6, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "bar",
			Port:    8000,
			Key:     "baz",
		},
	})

	// Act
	virtualApp1 := daemon1.GetVirtualApp()
	virtualApp2 := daemon2.GetVirtualApp()

	// Assert
	require.Equal(t, virtualApp1.ID, virtualApp2.ID)
}

// Test that a daemon without control access point has virtual app ID zero.
func TestGetVirtualAppNoControlAccessPoint(t *testing.T) {
	// Arrange
	machine := &Machine{Address: "foo"}
	daemon := NewDaemon(machine, daemonname.DHCPv4, true, []*AccessPoint{
		{
			Type:    AccessPointStatistics,
			Address: "bar",
			Port:    8000,
			Key:     "baz",
		},
	})

	// Act
	virtualApp := daemon.GetVirtualApp()

	// Assert
	require.Equal(t, int64(0), virtualApp.ID)
}

// Test that daemons can be retrieved by virtual app ID.
func TestGetDaemonsByVirtualAppID(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	m1 := &Machine{
		Address:   "machine1",
		AgentPort: 8080,
	}
	_ = AddMachine(db, m1)

	daemon1 := NewDaemon(m1, daemonname.DHCPv4, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "ap1",
			Port:    8000,
			Key:     "key1",
		},
	})
	_ = AddDaemon(db, daemon1)

	daemon2 := NewDaemon(m1, daemonname.DHCPv6, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "ap1",
			Port:    8000,
			Key:     "key1",
		},
	})
	_ = AddDaemon(db, daemon2)

	daemon3 := NewDaemon(m1, daemonname.CA, true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "ap2",
			Port:    9000,
			Key:     "key2",
		},
	})
	_ = AddDaemon(db, daemon3)

	// Act
	virtualApp1 := daemon1.GetVirtualApp()
	daemons, err := GetDaemonsByVirtualAppID(db, virtualApp1.ID)

	// Assert
	require.NoError(t, err)
	require.Len(t, daemons, 2)
	var daemonIDs []int64
	for _, d := range daemons {
		daemonIDs = append(daemonIDs, d.ID)
	}
	require.Contains(t, daemonIDs, daemon1.ID)
	require.Contains(t, daemonIDs, daemon2.ID)
}

// Test that the machine ID can be retrieved by virtual app ID.
func TestGetMachineIDByVirtualAppID(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	m := &Machine{
		Address:   "machine1",
		AgentPort: 8080,
	}
	_ = AddMachine(db, m)

	daemon := NewDaemon(m, daemonname.DHCPv4, true, []*AccessPoint{{
		Type:    AccessPointControl,
		Address: "ap1",
		Port:    8000,
		Key:     "key1",
	}})
	_ = AddDaemon(db, daemon)

	// Act
	virtualApp := daemon.GetVirtualApp()
	machineID, err := GetMachineIDByVirtualAppID(db, virtualApp.ID)

	// Assert
	require.NoError(t, err)
	require.Equal(t, m.ID, machineID)
	require.NotZero(t, machineID)
}

// Test that zero machine ID is returned for non-existing virtual app ID.
func TestGetMachineIDByVirtualAppIDNotFound(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Act
	machineID, err := GetMachineIDByVirtualAppID(db, 42)

	// Assert
	require.NoError(t, err)
	require.Zero(t, machineID)
}
