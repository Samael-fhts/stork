package agent

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	_ processLister    = (*processListerImpl)(nil)
	_ supportedProcess = (*processWrapper)(nil)
)

// An interface to a process detected by the agent. Using the interface
// allows for mocking listing the processes in the unit tests.
type supportedProcess interface {
	getCmdline() (string, error)
	getCwd() (string, error)
	getName() (string, error)
	getPid() int32
	getParentPid() (int32, error)
}

// Wrapper for gopsutil process. It implements the supportedProcess interface.
type processWrapper struct {
	process *process.Process
}

// Returns the process command line.
func (p *processWrapper) getCmdline() (string, error) {
	cmdline, err := p.process.Cmdline()
	err = errors.Wrapf(err, "failed to get process command line for pid %d", p.getPid())
	return cmdline, err
}

// Returns the process current working directory.
func (p *processWrapper) getCwd() (string, error) {
	cwd, err := p.process.Cwd()
	err = errors.Wrapf(err, "failed to get process current working directory for pid %d", p.getPid())
	return cwd, err
}

// Returns the process pid.
func (p *processWrapper) getPid() int32 {
	return p.process.Pid
}

// Returns the parent pid of the parent process.
func (p *processWrapper) getParentPid() (int32, error) {
	ppid, err := p.process.Ppid()
	err = errors.Wrapf(err, "failed to get process parent pid for pid %d", p.getPid())
	return ppid, err
}

// Returns the process name.
func (p *processWrapper) getName() (string, error) {
	name, err := p.process.Name()
	err = errors.Wrapf(err, "failed to get process name for pid %d", p.getPid())
	return name, err
}

// An interface for listing the supported processes. It can be mocked in the
// unit tests.
type processLister interface {
	listProcesses() ([]supportedProcess, error)
}

// A default implementation of the processLister interface.
type processListerImpl struct{}

// Lists the supported processes using gopsutil library.
func (impl *processListerImpl) listProcesses() ([]supportedProcess, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list running processes")
	}
	var listedProcesses []supportedProcess
	for _, p := range processes {
		listedProcesses = append(listedProcesses, &processWrapper{process: p})
	}
	return listedProcesses, nil
}

// An instance listing the supported processes and filtering out their child
// processes. Some DNS servers (e.g., NSD) spawn many child processes. The
// agent must not treat child processes as distinct daemons. Therefore, the
// manager only selects top-level processes, removing the ones having parent
// PID matching a PID of another process.
type ProcessManager struct {
	lister processLister
}

// Lists processes and filters out their child processes.
func (pm *ProcessManager) ListProcesses() ([]supportedProcess, error) {
	processes, err := pm.lister.listProcesses()
	if err != nil {
		return nil, err
	}

	// Creates an index of processes by their PID.
	tree := make(map[int32]supportedProcess)
	for _, p := range processes {
		tree[p.getPid()] = p
	}

	// Finds top-level processes (i.e., processes without a parent in the list
	// or processes whose parent has a different name).
	var acceptedCandidates []supportedProcess

	for _, process := range processes {
		ppid, err := process.getParentPid()
		if err != nil {
			// There is only one process without a parent: the init process.
			// We are not interested in it.
			continue
		}
		parentProcesses, exists := tree[ppid]
		if !exists {
			// No parent process found, so this is a top-level process.
			acceptedCandidates = append(acceptedCandidates, process)
			continue
		}

		parentName, err := parentProcesses.getName()
		if err != nil {
			// No permission to get the parent process name.
			continue
		}

		processName, err := process.getName()
		if err != nil {
			// No permission to get the process name.
			continue
		}

		if parentName != processName {
			// Parent process has a different name, so this is a top-level process.
			acceptedCandidates = append(acceptedCandidates, process)
		}
	}

	return acceptedCandidates, nil
}

// Returns the ProcessManager instance.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		lister: &processListerImpl{},
	}
}
