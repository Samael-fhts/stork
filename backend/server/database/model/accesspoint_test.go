package dbmodel

import (
	"testing"

	"github.com/stretchr/testify/require"
	"isc.org/stork/daemonctrl/daemonname"
	dbtest "isc.org/stork/server/database/test"
)

// Test that the no output and no error are returned if the entry is not found.
func TestGetAccessPointForMissingEntry(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Act
	accessPoint, err := GetAccessPoint(db, 42, AccessPointControl)

	// Assert
	require.NoError(t, err)
	require.Nil(t, accessPoint)
}

// Test that the error is returned if any database problem occurs.
func TestGetAccessPointForInvalidDatabase(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)

	// Act
	teardown()
	accessPoint, err := GetAccessPoint(db, 42, AccessPointControl)

	// Assert
	require.Error(t, err)
	require.Nil(t, accessPoint)
}

// Test that the access point is properly returned.
func TestGetAccessPoint(t *testing.T) {
	// Arrange
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	machine := &Machine{Address: "localhost", AgentPort: 8080}
	_ = AddMachine(db, machine)
	daemon := &Daemon{
		MachineID: machine.ID,
		Name:      daemonname.Bind9,
		AccessPoints: []*AccessPoint{{
			Type:     AccessPointControl,
			Address:  "127.0.0.1",
			Port:     8080,
			Key:      "secret",
			Protocol: "https",
		}},
	}
	_ = AddDaemon(db, daemon)

	// Act
	accessPoint, err := GetAccessPoint(db, daemon.ID, AccessPointControl)

	// Assert
	require.NoError(t, err)
	require.EqualValues(t, "127.0.0.1", accessPoint.Address)
	require.EqualValues(t, daemon.ID, accessPoint.DaemonID)
	require.EqualValues(t, "secret", accessPoint.Key)
	require.EqualValues(t, 8080, accessPoint.Port)
	require.EqualValues(t, AccessPointControl, accessPoint.Type)
	require.EqualValues(t, "https", accessPoint.Protocol)
}
