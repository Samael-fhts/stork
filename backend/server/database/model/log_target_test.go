package dbmodel

import (
	"testing"

	"github.com/stretchr/testify/require"
	dbtest "isc.org/stork/server/database/test"
)

// Test that the log target can be fetched from the database by ID.
func TestGetLogTargetByID(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	m := &Machine{
		ID:        0,
		Address:   "localhost",
		AgentPort: 8080,
	}
	err := AddMachine(db, m)
	require.NoError(t, err)
	require.NotZero(t, m.ID)

	daemon := NewDaemon(m, "kea-dhcp4", true, []*AccessPoint{
		{
			Type:    AccessPointControl,
			Address: "",
			Port:    8000,
			Key:     "",
		},
	})
	daemon.Version = "1.7.5"
	daemon.LogTargets = []*LogTarget{
		{
			Output: "stdout",
		},
		{
			Output: "/tmp/filename.log",
		},
	}
	err = AddDaemon(db, daemon)
	require.NoError(t, err)
	require.NotZero(t, daemon.ID)

	require.Len(t, daemon.LogTargets, 2)

	// Make sure that the log targets have been assigned IDs.
	require.NotZero(t, daemon.LogTargets[0].ID)
	require.NotZero(t, daemon.LogTargets[1].ID)

	// Get the first log target from the database by id.
	logTarget, err := GetLogTargetByID(db, daemon.LogTargets[0].ID)
	require.NoError(t, err)
	require.NotNil(t, logTarget)
	require.Equal(t, "stdout", logTarget.Output)
	require.NotNil(t, logTarget.Daemon)
	require.NotNil(t, logTarget.Daemon.Machine)

	// Get the second log target by id.
	logTarget, err = GetLogTargetByID(db, daemon.LogTargets[1].ID)
	require.NoError(t, err)
	require.NotNil(t, logTarget)
	require.Equal(t, "/tmp/filename.log", logTarget.Output)
	require.NotNil(t, logTarget.Daemon)
	require.NotNil(t, logTarget.Daemon.Machine)

	// Use the non existing id. This should return nil.
	logTarget, err = GetLogTargetByID(db, daemon.LogTargets[1].ID+1000)
	require.NoError(t, err)
	require.Nil(t, logTarget)
}
