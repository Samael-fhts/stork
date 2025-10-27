package dbmodel

import (
	"fmt"
	"hash/adler32"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"isc.org/stork/daemonctrl/daemonname"
)

// The virtual app is a solution to preserve backward compatibility with the
// legacy REST API. It mimics a legacy App instance for a given daemon.
type VirtualApp struct {
	ID   int64
	Name string
	Type VirtualAppType
}

type VirtualAppType string

const (
	VirtualAppTypeKea   VirtualAppType = "kea"
	VirtualAppTypeBind9 VirtualAppType = "bind9"
	VirtualAppTypePDNS  VirtualAppType = "pdns"
)

// Returns a virtual app for a given daemon.
func (d Daemon) GetVirtualApp() *VirtualApp {
	var appID int64
	if accessPoint, err := d.GetAccessPoint(AccessPointControl); err == nil {
		// The app ID must be deterministic, be the same for the daemons having
		// the same control access point and unique in other cases.
		checksum := adler32.Checksum([]byte(fmt.Sprintf("%s:%d", accessPoint.Address, accessPoint.Port)))
		appID = int64(checksum)
	}

	var appType VirtualAppType
	switch {
	case d.Name.IsKea():
		appType = VirtualAppTypeKea
	case d.Name == daemonname.Bind9:
		appType = VirtualAppTypeBind9
	case d.Name == daemonname.PDNS:
		appType = VirtualAppTypePDNS
	}

	appName := fmt.Sprintf("%s@%s%%%d", appType, d.Machine.Address, appID)

	return &VirtualApp{
		ID:   appID,
		Name: appName,
		Type: appType,
	}
}

// Returns daemons generating a given (virtual) app ID. The app ID is derived
// from the control access point of the daemon. If multiple daemons share the
// same control access point, they will have the same app ID and all of them
// will be returned.
func GetDaemonsByVirtualAppID(dbi pg.DBI, appID int64) (daemons []*Daemon, err error) {
	var accessPoints []AccessPoint
	err = dbi.Model(&accessPoints).
		Where("type = ?", AccessPointControl).
		Select()
	if err != nil {
		return nil, errors.Wrapf(err, "problem selecting control access points")
	}

	var matchingDaemonIDs []int64
	for _, ap := range accessPoints {
		checksum := adler32.Checksum([]byte(fmt.Sprintf("%s:%d", ap.Address, ap.Port)))
		if int64(checksum) == appID {
			matchingDaemonIDs = append(matchingDaemonIDs, ap.DaemonID)
		}
	}

	if len(matchingDaemonIDs) == 0 {
		// No matching access points, return empty result.
		return []*Daemon{}, nil
	}

	err = dbi.Model(&daemons).
		Relation(DaemonRelationAccessPoints).
		Relation(DaemonRelationMachine).
		Relation(DaemonRelationKeaDHCPDaemon).
		Relation(DaemonRelationBind9Daemon).
		Relation(DaemonRelationPDNSDaemon).
		Where("daemon.id IN (?)", pg.In(matchingDaemonIDs)).
		OrderExpr("daemon.id ASC").
		Select()

	if errors.Is(err, pg.ErrNoRows) {
		return daemons, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "problem selecting daemons for virtual app ID %d", appID)
	}

	return daemons, nil
}
