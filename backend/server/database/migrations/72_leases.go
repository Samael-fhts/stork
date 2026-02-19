package dbmigs

import (
	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		_, err := db.Exec(
			`-- Leases table.
             CREATE TABLE IF NOT EXISTS public.leases (
                 id              BIGSERIAL NOT NULL,
                 ip_version      SMALLINT NOT NULL,
                 hw_address      MACADDR,
                 duid            BYTEA,
                 ip_address      INET NOT NULL,
                 cltt            TIMESTAMP WITHOUT TIME ZONE NOT NULL,
                 state           SMALLINT NOT NULL,
                 valid_lifetime  INT NOT NULL,
                 prefix_length   SMALLINT,
                 subnet_id       BIGINT,
                 daemon_id       BIGINT,
                 CONSTRAINT leases_pkey PRIMARY KEY (id),
                 CONSTRAINT leases_subnet_fkey FOREIGN KEY (subnet_id)
	                 REFERENCES subnet (id) MATCH SIMPLE
                         ON UPDATE CASCADE
                         ON DELETE SET NULL,
                 CONSTRAINT leases_daemons_fkey FOREIGN KEY (daemon_id)
                     REFERENCES daemon (id) MATCH SIMPLE
                         ON UPDATE CASCADE
                         ON DELETE SET NULL
             );

             -- Lease updates table.
             CREATE TABLE IF NOT EXISTS public.lease_updates (
                 id              BIGSERIAL NOT NULL,
                 ip_version      SMALLINT NOT NULL,
                 hw_address      MACADDR,
                 duid            BYTEA,
                 ip_address      INET NOT NULL,
                 cltt            TIMESTAMP WITHOUT TIME ZONE NOT NULL,
                 state           SMALLINT NOT NULL,
                 valid_lifetime  INT NOT NULL,
                 prefix_length   SMALLINT,
                 subnet_id       BIGINT,
                 daemon_id       BIGINT,
                 CONSTRAINT lease_updates_pkey PRIMARY KEY (id),
                 CONSTRAINT lease_updates_subnet_fkey FOREIGN KEY (subnet_id)
	                 REFERENCES subnet (id) MATCH SIMPLE
                         ON UPDATE CASCADE
                         ON DELETE SET NULL,
                 CONSTRAINT lease_updates_daemons_fkey FOREIGN KEY (daemon_id)
                     REFERENCES daemon (id) MATCH SIMPLE
                         ON UPDATE CASCADE
                         ON DELETE SET NULL
             );
           `)
		return err
	}, func(db migrations.DB) error {
		_, err := db.Exec(
			`-- Remove table with leases.
             DROP TABLE IF EXISTS public.leases;
             -- Remove table with lease updates.
             DROP TABLE IF EXISTS public.lease_updates;
        `)
		return err
	})
}
