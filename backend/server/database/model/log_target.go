package dbmodel

import (
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	pkgerrors "github.com/pkg/errors"
	dbops "isc.org/stork/server/database"
)

// A structure reflecting information about a logger used by a daemon.
type LogTarget struct {
	ID        int64 // Logger ID
	Name      string
	Severity  string
	Output    string
	CreatedAt time.Time

	DaemonID int64
	Daemon   *Daemon `pg:"rel:has-one"`
}

// Retrieves log target from the database by id.
func GetLogTargetByID(db dbops.DBI, id int64) (*LogTarget, error) {
	logTarget := LogTarget{}
	err := db.Model(&logTarget).
		Relation("Daemon.App.Machine").
		Where("log_target.id = ?", id).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, pkgerrors.Wrapf(err, "problem getting log target with ID %d", id)
	}
	return &logTarget, nil
}

// Deletes a log targets by IDs. Keeps the log targets with IDs in keepIDs slice.
func DeleteLogTargetByDaemonID(db dbops.DBI, daemonID int64, keepIDs []int64) error {
	q := db.Model(&LogTarget{}).
		Where("log_target.daemon_id = ?", daemonID)
	if len(keepIDs) > 0 {
		q = q.Where("log_target.id NOT IN (?)", pg.In(keepIDs))
	}
	_, err := q.Delete()
	return pkgerrors.Wrapf(err, "problem deleting log targets for daemon ID %d, keeping IDs: %v", daemonID, keepIDs)
}

// Adds a log target to the database.
func AddLogTarget(db dbops.DBI, logTarget *LogTarget) error {
	_, err := db.Model(logTarget).Insert()
	return pkgerrors.Wrapf(err, "problem adding log target %+v", logTarget)
}

// Updates a log target in the database.
func UpdateLogTarget(db dbops.DBI, logTarget *LogTarget) error {
	result, err := db.Model(logTarget).WherePK().Update()
	if err != nil {
		err = pkgerrors.Wrapf(err, "problem updating log target %+v", logTarget)
	} else if result.RowsAffected() <= 0 {
		err = pkgerrors.Wrapf(ErrNotExists, "log target with ID %d does not exist", logTarget.ID)
	}
	return err
}
