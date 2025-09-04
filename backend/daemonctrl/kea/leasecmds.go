package keactrl

import "isc.org/stork/daemonctrl/constant"

// Lease type specified in the commands.
type LeaseType string

const (
	LeaseTypeNA LeaseType = "IA_NA"
	LeaseTypePD LeaseType = "IA_PD"
)

const (
	Lease4Get            CommandName = "lease4-get"
	Lease6Get            CommandName = "lease6-get"
	Lease4GetByClientID  CommandName = "lease4-get-by-client-id"
	Lease6GetByDUID      CommandName = "lease6-get-by-duid"
	Lease4GetByHostname  CommandName = "lease4-get-by-hostname"
	Lease6GetByHostname  CommandName = "lease6-get-by-hostname"
	Lease4GetByHWAddress CommandName = "lease4-get-by-hw-address"
)

// Creates lease4-get command.
func NewCommandLease4Get(ipAddress string, daemon constant.KeaDaemonName) *Command {
	return NewCommandBase(Lease4Get, daemon).WithArgument("ip-address", ipAddress)
}

// Creates lease6-get command.
func NewCommandLease6Get(leaseType LeaseType, ipAddress string, daemon constant.KeaDaemonName) *Command {
	return NewCommandBase(Lease6Get, daemon).
		WithArgument("type", leaseType).
		WithArgument("ip-address", ipAddress)
}
