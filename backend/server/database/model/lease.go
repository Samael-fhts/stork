package dbmodel

import (
	"context"

	"github.com/go-pg/pg/v10"
	//"github.com/go-pg/pg/v10/orm"

	keadata "isc.org/stork/appdata/kea"

	pkgerrors "github.com/pkg/errors"
)

// Extends basic Lease information with database specific information.
type Lease struct {
	ID int64

	keadata.Lease

	AppID int64
	App   *App
}

type LeaseUpdate struct {
	ID            int64
	Address       string
	HWAddress     []byte
	ClientID      []byte
	ValidLifetime uint32
	/*
		Cltt          uint64
		SubnetID      uint32
		FqdnFwd       bool
		FqdnRev       bool
		Hostname      string
		State         int
		UserContext   string
	*/
	AppID uint64
}

func addLeaseInternal(tx *pg.Tx, lease *LeaseUpdate) error {
	_, err := tx.Model(lease).Insert()
	if err != nil {
		err = pkgerrors.WithMessagef(err, "problem with adding lease %s", lease.Address)
		return err
	}
	return err
}

func AddLease(dbIface interface{}, lease *LeaseUpdate) error {
	if db, ok := dbIface.(*pg.DB); ok {
		return db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
			return addLeaseInternal(tx, lease)
		})
	}
	return addLeaseInternal(dbIface.(*pg.Tx), lease)
}
