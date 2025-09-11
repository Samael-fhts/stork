package daemons

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	agentcommtest "isc.org/stork/server/agentcomm/test"
	"isc.org/stork/server/config"
	"isc.org/stork/server/daemons/kea"
	appstest "isc.org/stork/server/daemons/test"
	dbmodel "isc.org/stork/server/database/model"
	dbtest "isc.org/stork/server/database/test"
)

// An error returned by the Commit function in fake Kea module.
type lackingStateError struct{}

// Error implementation.
func (lackingStateError) Error() string {
	return "context lacks state"
}

// Fake Kea module exposing a Commit function. It is used to
// test that the manager's Commit() function properly routes
// the calls to the Commit() function in the Kea module.
type fakeKeaModuleCommit struct {
	contexts []context.Context
	ops      []config.Operation
	err      error
}

// Creates new instance of the fake Kea module.
func newFakeKeaModuleCommit() *fakeKeaModuleCommit {
	return &fakeKeaModuleCommit{}
}

// Implementation of the fake Commit() function. It records
// the invoked commit operations and passed contexts.
func (fkm *fakeKeaModuleCommit) Commit(ctx context.Context) (context.Context, error) {
	state, ok := config.GetTransactionState[kea.ConfigRecipe](ctx)
	if !ok {
		return ctx, lackingStateError{}
	}
	fkm.contexts = append(fkm.contexts, ctx)
	for _, update := range state.Updates {
		fkm.ops = append(fkm.ops, update.Operation)
	}
	return ctx, fkm.err
}

// Test creating new config manager instance.
func TestNewManager(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	agents := &agentcommtest.FakeAgents{}
	lookup := dbmodel.NewDHCPOptionDefinitionLookup()
	daemonLocker := config.NewDaemonLocker()

	manager := NewManager(&appstest.ManagerAccessorsWrapper{
		DB:           db,
		Agents:       agents,
		DefLookup:    lookup,
		DaemonLocker: daemonLocker,
	})
	require.NotNil(t, manager)
	require.NotNil(t, manager.GetKeaModule())

	impl, ok := manager.(*configManagerImpl)
	require.True(t, ok)
	require.NotNil(t, impl)
	require.Equal(t, db, impl.GetDB())
	require.Equal(t, agents, impl.GetConnectedAgents())
	require.Equal(t, lookup, impl.GetDHCPOptionDefinitionLookup())
	require.Equal(t, daemonLocker, impl.GetDaemonLocker())
}

// Test creating new context with context ID and user ID.
func TestCreateContext(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	// Gather the generated context ids in the map to ensure
	// that each created context has a unique context ID.
	ids := make(map[int64]bool)
	for i := 0; i < 10; i++ {
		// Create new context with user ID between 0 and 9.
		ctx, err := manager.CreateContext(int64(i))
		require.NoError(t, err)
		require.NotNil(t, ctx)

		// Make sure that the context ID exists.
		ctxID, ok := config.GetValueAsInt64(ctx, config.ContextIDKey)
		require.True(t, ok)
		ids[ctxID] = true

		// Make sure that the user ID exists.
		userid, ok := config.GetValueAsInt64(ctx, config.UserContextKey)
		require.True(t, ok)
		require.EqualValues(t, i, userid)
	}
	// Ensure that each call to CreateContext generated new ID.
	require.Len(t, ids, 10)
}

// Test that a created context can be remembered and then recovered
// by context ID and user ID.
func TestRememberRecoverContext(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	// Create first context with user ID 123.
	ctx1, err := manager.CreateContext(int64(123))
	require.NoError(t, err)
	require.NotNil(t, ctx1)

	// Linters do not like when simple types are used for keys in the context.
	type testContextKeyType string

	// Add some additional data specific to this context.
	key := testContextKeyType("foo")
	ctx1 = context.WithValue(ctx1, key, "bar")

	// Retrieve the generated context ID. It will be later needed
	// to recover the context.
	id1, ok := config.GetValueAsInt64(ctx1, config.ContextIDKey)
	require.True(t, ok)

	// Store the context.
	err = manager.RememberContext(ctx1, time.Minute*10)
	require.NoError(t, err)
	defer manager.Done(ctx1)

	// Try to recover the context by retrieved ID and user ID.
	recovered1, cancel1 := manager.RecoverContext(id1, 123)
	require.NotNil(t, recovered1)
	require.NotNil(t, cancel1)

	// The context ID and user ID should be present in the recovered context.
	_, ok = config.GetValueAsInt64(recovered1, config.ContextIDKey)
	require.True(t, ok)
	user1, ok := config.GetValueAsInt64(recovered1, config.UserContextKey)
	require.True(t, ok)
	require.EqualValues(t, 123, user1)

	// Ensure that the context specific information is also present.
	foo, ok := recovered1.Value(key).(string)
	require.True(t, ok)
	require.Equal(t, "bar", foo)

	// Repeat the same test for the second context. context.
	ctx2, err := manager.CreateContext(int64(234))
	require.NoError(t, err)
	require.NotNil(t, ctx2)

	key = testContextKeyType("bar")
	ctx2 = context.WithValue(ctx2, key, "baz")

	id2, ok := config.GetValueAsInt64(ctx2, config.ContextIDKey)
	require.True(t, ok)

	err = manager.RememberContext(ctx2, time.Minute*10)
	require.NoError(t, err)
	defer manager.Done(ctx2)

	recovered2, cancel2 := manager.RecoverContext(id2, 234)
	require.NotNil(t, recovered2)
	require.NotNil(t, cancel2)

	_, ok = config.GetValueAsInt64(recovered2, config.ContextIDKey)
	require.True(t, ok)
	user2, ok := config.GetValueAsInt64(recovered2, config.UserContextKey)
	require.True(t, ok)
	require.EqualValues(t, 234, user2)

	bar, ok := recovered2.Value(key).(string)
	require.True(t, ok)
	require.Equal(t, "baz", bar)
}

// Test the case when a timeout occurs during config update.
func TestContextTimeout(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{
		DaemonLocker: config.NewDaemonLocker(),
	})
	require.NotNil(t, manager)

	ctx, err := manager.CreateContext(int64(123))
	require.NoError(t, err)
	require.NotNil(t, ctx)

	contextID, ok := config.GetValueAsInt64(ctx, config.ContextIDKey)
	require.True(t, ok)

	// Remember the context.
	err = manager.RememberContext(ctx, time.Second*10)
	require.NoError(t, err)

	// Use the context to lock the daemon 1.
	ctx, err = manager.Lock(ctx, 1)
	require.NoError(t, err)
	defer manager.Unlock(ctx)

	// Remember the context again. It specifies a very short timeout
	// overriding the previous timeout of 10s.
	err = manager.RememberContext(ctx, time.Microsecond)
	require.NoError(t, err)

	// Wait for a timeout. When the timeout elapses, an attempt to recover
	// the context should return nil because the context should be removed
	// after the timeout.
	require.Eventually(t, func() bool {
		ctx, _ := manager.RecoverContext(contextID, 123)
		return ctx == nil
	}, time.Second, time.Millisecond)

	// Try to lock the configuration on daemon 1. It should succeed because
	// the configuration should have been unlocked after the timeout.
	ctxLock, err := manager.CreateContext(int64(234))
	require.NoError(t, err)
	require.NotNil(t, ctxLock)
	require.Eventually(t, func() bool {
		ctxLock, err = manager.Lock(ctxLock, 1)
		defer manager.Unlock(ctxLock)
		return err == nil
	}, time.Second, time.Millisecond)
}

// Test that calling Done() function results in removing the context and
// unlocking the configuration.
func TestDone(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{
		DaemonLocker: config.NewDaemonLocker(),
	})
	require.NotNil(t, manager)

	ctx, err := manager.CreateContext(int64(123))
	require.NoError(t, err)
	require.NotNil(t, ctx)

	contextID, ok := config.GetValueAsInt64(ctx, config.ContextIDKey)
	require.True(t, ok)

	ctx, err = manager.Lock(ctx, 1)
	require.NoError(t, err)
	defer manager.Unlock(ctx)

	err = manager.RememberContext(ctx, time.Second*10)
	require.NoError(t, err)

	manager.Done(ctx)

	// An attempt to recover the context should return nil.
	ctx, cancel := manager.RecoverContext(contextID, 123)
	require.Nil(t, ctx)
	require.Nil(t, cancel)

	// An attempt to lock the daemon configuration should succeed
	// because the previous lock should have been removed as a result
	// of calling Done().
	ctxLock, err := manager.CreateContext(int64(234))
	require.NoError(t, err)
	require.NotNil(t, ctxLock)
	_, err = manager.Lock(ctxLock, 1)
	require.NoError(t, err)
	manager.Unlock(ctxLock)
}

// Test that that an error is returned upon an attempt to remember the context
// under the specific context ID when user ID doesn't match.
func TestRememberContextWithMismatchedUserID(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	// Create context with user ID 123.
	ctx, err := manager.CreateContext(int64(123))
	require.NoError(t, err)
	require.NotNil(t, ctx)

	// Remember the context.
	err = manager.RememberContext(ctx, time.Minute*10)
	require.NoError(t, err)

	// Retrieve the context ID. We are going to use this ID instead of the
	// user ID when trying to replace the remembered context. It should
	// cause the mismatch.
	id, ok := config.GetValueAsInt64(ctx, config.ContextIDKey)
	require.True(t, ok)

	// In unlikely event that both ids happen to be equal, modify the
	// ID to avoid the test failure.
	if id == 123 {
		id++
	}
	ctx = context.WithValue(ctx, config.UserContextKey, id)
	err = manager.RememberContext(ctx, time.Minute*10)
	require.Error(t, err)
}

// Test that nil context is returned when user ID or context ID doesn't
// match the remembered values.
func TestRecoverContextMismatch(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	// Create first context with user ID 123.
	ctx1, err := manager.CreateContext(int64(123))
	require.NoError(t, err)
	require.NotNil(t, ctx1)
	id1, ok := config.GetValueAsInt64(ctx1, config.ContextIDKey)
	require.True(t, ok)
	err = manager.RememberContext(ctx1, time.Minute*10)
	require.NoError(t, err)

	// Create second context with user ID 234.
	ctx2, err := manager.CreateContext(int64(234))
	require.NoError(t, err)
	require.NotNil(t, ctx2)
	id2, ok := config.GetValueAsInt64(ctx2, config.ContextIDKey)
	require.True(t, ok)
	err = manager.RememberContext(ctx2, time.Minute*10)
	require.NoError(t, err)

	// When a user ID or context ID doesn't match the nil context
	// should be returned.
	recovered, cancel := manager.RecoverContext(id1, 234)
	require.Nil(t, recovered)
	require.Nil(t, cancel)
	recovered, cancel = manager.RecoverContext(id2, 123)
	require.Nil(t, recovered)
	require.Nil(t, cancel)
	recovered, cancel = manager.RecoverContext(111, 111)
	require.Nil(t, recovered)
	require.Nil(t, cancel)
}

// Test that daemon configurations can be locked for updates and then
// unlocked allowing for locking again.
func TestLockUnlock(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{
		DaemonLocker: config.NewDaemonLocker(),
	})
	require.NotNil(t, manager)

	// Create context and lock daemons 1, 2, 3.
	ctx1, err := manager.CreateContext(123)
	require.NoError(t, err)
	ctx1, err = manager.Lock(ctx1, 1, 2, 3)
	require.NoError(t, err)

	// An attempt to lock one of these daemons should fail.
	_, err = manager.Lock(ctx1, 4, 1)
	require.Error(t, err)

	// Create another context and try to lock unlocked daemon by different user.
	ctx2, err := manager.CreateContext(234)
	require.NoError(t, err)
	ctx2, err = manager.Lock(ctx2, 4)
	require.NoError(t, err)

	// Locking already locked daemon should fail.
	_, err = manager.Lock(ctx2, 1)
	require.Error(t, err)

	// Unlock the daemons locked by the first user.
	manager.Unlock(ctx1)

	// An attempt to lock the daemon should this time pass.
	_, err = manager.Lock(ctx2, 1)
	require.NoError(t, err)
}

// Test that the context passed to the Unlock method must contain the lock key.
// Otherwise, no daemons are unlocked.
func TestUnlockForMissingKey(t *testing.T) {
	// Arrange
	locker := config.NewDaemonLocker()
	manager := NewManager(&appstest.ManagerAccessorsWrapper{
		DaemonLocker: locker,
	})
	require.NotNil(t, manager)

	lockCtx, _ := manager.CreateContext(123)
	lockCtx, _ = manager.Lock(lockCtx, 1, 2, 3)

	// Context with no lock key.
	unlockCtx := context.WithValue(
		context.Background(),
		config.DaemonsContextKey,
		lockCtx.Value(config.DaemonsContextKey),
	)

	// Act
	manager.Unlock(unlockCtx)

	// Assert
	for _, daemonID := range []int64{1, 2, 3} {
		require.True(t, locker.IsLocked(daemonID))
	}
}

// Test that the commit call is routed to the Kea module when the
// transaction target is dbmodel.AppTypeKea.
func TestCommitKeaModule(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	// Replace the interface for committing changes in the Kea
	// configuration module for the fake one.
	impl := manager.(*configManagerImpl)
	require.NotNil(t, impl)
	fkm := newFakeKeaModuleCommit()
	impl.keaCommit = fkm

	ctx, err := impl.CreateContext(123)
	require.NoError(t, err)

	// Create a new transaction with Kea.
	state := config.TransactionState[kea.ConfigRecipe]{
		Updates: []*config.Update[kea.ConfigRecipe]{
			config.NewUpdate[kea.ConfigRecipe](config.OperationKeaHostAdd),
		},
	}
	ctx = context.WithValue(ctx, config.StateContextKey, state)

	// Commit the changes. They should result in a call to the Kea
	// module.
	_, err = manager.Commit(ctx)
	require.NoError(t, err)
	require.Len(t, fkm.ops, 1)
	require.Equal(t, config.OperationKeaHostAdd, fkm.ops[0])
}

// Test that an error is returned when unknown tool is specified in the
// Kea context.
func TestCommitUnknownTarget(t *testing.T) {
	manager := NewManager(&appstest.ManagerAccessorsWrapper{})
	require.NotNil(t, manager)

	ctx, err := manager.CreateContext(123)
	require.NoError(t, err)

	// Create a new transaction with unknown target.
	state := config.TransactionState[any]{
		Updates: []*config.Update[any]{
			config.NewUpdate[any](config.OperationKeaHostAdd),
		},
	}
	ctx = context.WithValue(ctx, config.StateContextKey, state)

	// Commit the changes and expect an error.
	_, err = manager.Commit(ctx)
	require.Error(t, err)
}
