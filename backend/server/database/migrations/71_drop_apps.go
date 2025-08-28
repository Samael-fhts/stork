package dbmigs

import (
	"github.com/go-pg/migrations/v8"
	"github.com/pkg/errors"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		_, err := db.Exec(`
			-- Add a column in the access point table to store a new foreign key
			-- to the daemon table.
			ALTER TABLE access_point ADD COLUMN daemon_id bigint;
			-- Fill a reference to the daemon table. Use the daemon ID of the
			-- control daemon to preserve the existing capabilities to connect
			-- to the Kea CA daemon. This value is temporary and will be
			-- updated as soon as the Stork agent re-detects the Kea processes.
			UPDATE access_point
			SET daemon_id = daemon.id
			FROM app, daemon
			WHERE access_point.app_id = app.id AND app.id = daemon.app_id;
			-- Set constraints for the new column.
			ALTER TABLE access_point ALTER COLUMN daemon_id SET NOT NULL;
			ALTER TABLE access_point ADD CONSTRAINT access_point_daemon_id_fkey
				FOREIGN KEY (daemon_id)
				REFERENCES daemon(id)
				ON DELETE CASCADE;
			-- Drop the unnecessary reference to the machine table.
			ALTER TABLE access_point DROP COLUMN machine_id;
			-- Change the primary key.
			ALTER TABLE access_point DROP CONSTRAINT access_point_pkey;
			ALTER TABLE access_point ADD CONSTRAINT access_point_pkey
				PRIMARY KEY (daemon_id, type);
			-- Drop the unnecessary reference to the app table.
			ALTER TABLE access_point DROP COLUMN app_id;

			-- Add a reference to the machine table in the daemon table.
			ALTER TABLE daemon ADD COLUMN machine_id bigint;
			-- Fill the reference to the machine table.
			UPDATE daemon
			SET machine_id = machine.id
			FROM app, machine
			WHERE daemon.app_id = app.id AND app.machine_id = machine.id;
			-- Set constraints for the new column.
			ALTER TABLE daemon ALTER COLUMN machine_id SET NOT NULL;
			ALTER TABLE daemon ADD CONSTRAINT daemon_machine_id_fkey
				FOREIGN KEY (machine_id)
				REFERENCES machine(id)
				ON DELETE CASCADE;
			-- Drop the unnecessary reference to the app table.
			ALTER TABLE daemon DROP COLUMN app_id;

			-- Update name of the state puller puller in settings.
			UPDATE setting
			SET name = 'state_puller_interval'
			WHERE name = 'apps_state_puller_interval';
		`)
		return err
	}, func(db migrations.DB) error {
		return errors.New("not implemented")
	})
}
