package keactrl

import (
	keaconfig "isc.org/stork/daemoncfg/kea"
	"isc.org/stork/datamodel/constants/daemonname"
)

const (
	ReservationAdd     CommandName = "reservation-add"
	ReservationDel     CommandName = "reservation-del"
	ReservationGetPage CommandName = "reservation-get-page"
)

// Creates reservation-add command.
func NewCommandReservationAdd(reservation *keaconfig.HostCmdsReservation, daemonName daemonname.Name) *Command {
	return newCommand(ReservationAdd, daemonName, map[string]any{"reservation": reservation})
}

// Creates reservation-del command.
func NewCommandReservationDel(reservation *keaconfig.HostCmdsDeletedReservation, daemonName daemonname.Name) *Command {
	return newCommand(ReservationDel, daemonName, reservation)
}

// Creates reservation-get-page command. The arguments from and source-index
// are only included in the command when they are greater than 0.
func NewCommandReservationGetPage(localSubnetID, sourceIndex, from, limit int64, daemonName daemonname.Name) *Command {
	arguments := map[string]any{
		"subnet-id": localSubnetID,
		"limit":     limit,
	}

	if from > 0 {
		arguments["from"] = from
	}

	if sourceIndex > 0 {
		arguments["source-index"] = sourceIndex
	}

	return newCommand(ReservationGetPage, daemonName, arguments)
}
