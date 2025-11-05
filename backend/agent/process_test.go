package agent

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// Test listing processes with eliminating child processes.
func TestListProcesses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Process tree:
	//  0 (root)
	//  |
	//  5 (supervisord)       ? (unknown)
	// /         \            |
	// 1 (CA)     4 (CA)      7 (unknown
	// |                      |
	// 2 (CA)                 6 (CA)
	// |
	// 3 (CA)

	proc1 := NewMockSupportedProcess(ctrl)
	proc1.EXPECT().getPid().AnyTimes().Return(int32(1))
	proc1.EXPECT().getName().AnyTimes().Return("kea-ctrl-agent", nil)
	proc1.EXPECT().getParentPid().AnyTimes().Return(int32(5), nil)

	proc2 := NewMockSupportedProcess(ctrl)
	proc2.EXPECT().getPid().AnyTimes().Return(int32(2))
	proc2.EXPECT().getName().AnyTimes().Return("kea-ctrl-agent", nil)
	proc2.EXPECT().getParentPid().AnyTimes().Return(int32(1), nil)

	proc3 := NewMockSupportedProcess(ctrl)
	proc3.EXPECT().getPid().AnyTimes().Return(int32(3))
	proc3.EXPECT().getName().AnyTimes().Return("kea-ctrl-agent", nil)
	proc3.EXPECT().getParentPid().AnyTimes().Return(int32(2), nil)

	proc4 := NewMockSupportedProcess(ctrl)
	proc4.EXPECT().getPid().AnyTimes().Return(int32(4))
	proc4.EXPECT().getName().AnyTimes().Return("kea-ctrl-agent", nil)
	proc4.EXPECT().getParentPid().AnyTimes().Return(int32(5), nil)

	proc5 := NewMockSupportedProcess(ctrl)
	proc5.EXPECT().getPid().AnyTimes().Return(int32(5))
	proc5.EXPECT().getName().AnyTimes().Return("supervisord", nil)
	proc5.EXPECT().getParentPid().AnyTimes().Return(int32(0), nil)

	proc6 := NewMockSupportedProcess(ctrl)
	proc6.EXPECT().getPid().AnyTimes().Return(int32(6))
	proc6.EXPECT().getName().AnyTimes().Return("supervisord", nil)
	proc6.EXPECT().getParentPid().AnyTimes().Return(int32(7), nil)

	proc0 := NewMockSupportedProcess(ctrl)
	proc0.EXPECT().getPid().AnyTimes().Return(int32(0))
	proc0.EXPECT().getName().AnyTimes().Return("init", nil)
	proc0.EXPECT().getParentPid().AnyTimes().Return(int32(0), errors.New("no parent"))

	lister := NewMockProcessLister(ctrl)
	lister.EXPECT().listProcesses().Return([]supportedProcess{proc0, proc1, proc2, proc3, proc4, proc5, proc6}, nil)

	pm := NewProcessManager()
	pm.lister = lister
	processes, err := pm.ListProcesses()
	require.NoError(t, err)
	require.Len(t, processes, 4)

	// The highest-level Kea Ctrl Agent process. First branch.
	require.EqualValues(t, 1, processes[0].getPid())
	// The highest-level Kea Ctrl Agent process. Second branch.
	require.EqualValues(t, 4, processes[1].getPid())
	// The highest-level Supervisord process.
	require.EqualValues(t, 5, processes[2].getPid())
	// The highest-level known Kea Ctrl Agent process (orphaned).
	require.EqualValues(t, 6, processes[3].getPid())
}
