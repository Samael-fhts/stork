package configmigrator

import (
	"github.com/go-pg/pg/v10"
	dbmodel "isc.org/stork/server/database/model"
)

type hostMigrator struct {
	db     *pg.DB
	filter dbmodel.HostsByPageFilters
	items  []dbmodel.Host
	limit  int64
}

var _ Migrator = &hostMigrator{}

func (m *hostMigrator) CountTotal() (int64, error) {
	_, count, err := dbmodel.GetHostsByPage(m.db, 0, 0, m.filter, "", dbmodel.SortDirAny)
	return count, err
}

func (m *hostMigrator) LoadItems(offset int64) (int64, error) {
	items, count, err := dbmodel.GetHostsByPage(m.db, offset, m.limit, m.filter, "id", dbmodel.SortDirAsc)
	if err != nil {
		// Returns the number of items tried to load.
		return m.limit, err
	}
	m.items = items
	return count, nil
}

// Adds the hosts to the database. Sends the delete command to Kea.
func (m *hostMigrator) Migrate() map[int64]error {
	errs := make(map[int64]error)

	for _, host := range m.items {

	}

	return errs
}

// Sends the config write command to Kea.
func (m *hostMigrator) Finish() error {

}
