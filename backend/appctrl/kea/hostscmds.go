package keactrl

import keaconfig "isc.org/stork/appcfg/kea"

func NewReservationAddCommand(service DHCPServiceName, reservation *keaconfig.HostCmdsReservation) *Command {
	arguments := map[string]any{
		"reservation": []*keaconfig.HostCmdsReservation{reservation},
	}
	return NewCommand("reservation-add", wrapDHCPService(service), arguments)
}

func NewReservationDelCommand(service DHCPServiceName, reservation *keaconfig.HostCmdsDeletedReservation) *Command {
	return NewCommand("reservation-del", wrapDHCPService(service), reservation)
}

type ReservationGetPageArguments struct {
	SubnetID    int64
	Limit       int64
	From        *int64
	SourceIndex *int64
}

func NewReservationGetPageCommand(service DHCPServiceName, arguments ReservationGetPageArguments) *Command {
	rawArguments := map[string]any{
		"subnet-id": arguments.SubnetID,
		"limit":     arguments.Limit,
	}

	if arguments.From != nil && *arguments.From > 0 {
		rawArguments["from"] = arguments.From
	}
	if arguments.SourceIndex != nil && *arguments.SourceIndex > 0 {
		rawArguments["source-index"] = arguments.SourceIndex
	}

	return NewCommand("reservation-get-page", wrapDHCPService(service), arguments)
}
