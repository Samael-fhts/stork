import { Injectable } from '@angular/core'
import { Observable, concatAll, interval, map, takeWhile } from 'rxjs'
import { DHCPService, HostMigration, HostMigrationFilter } from '../backend'

/**
 * The migration service to interact with the backend.
 */
@Injectable({
    providedIn: 'root',
})
export class HostsMigrationService {
    constructor(private dhcpApi: DHCPService) {}

    /**
     * Checks the status of the current migration. It's empty if there is no
     * migration in progress.
     * @returns An observable emitting the current migration status.
     */
    getCurrentMigration(): Observable<HostMigration> {
        return this.dhcpApi.getHostMigration()
    }

    /**
     * Starts a new migration.
     * @param filter The host reservation filter to apply during migration.
     * @returns An observable emitting the migration status.
     */
    startMigration(filter: HostMigrationFilter): Observable<HostMigration> {
        return this.dhcpApi.createHostMigration(filter)
    }

    /**
     * Stops and removes the current migration.
     * @returns An observable emitting nothing.
     */
    removeMigration(): Observable<void> {
        return this.dhcpApi.deleteHostMigration()
    }

    /**
     * Returns a stream of migration updates.
     * @returns An observable emitting migration updates.
     */
    getMigrationUpdates(): Observable<HostMigration> {
        // Creates an observable that emits in a regular interval. Each
        // emission fetches the current migration status. If the migration is
        // finished (inProgress = false), the observable completes.
        return interval(1000).pipe(
            map(() => this.dhcpApi.getHostMigration()),
            concatAll(),
            takeWhile((m) => m.inProgress, true)
        )
    }
}
