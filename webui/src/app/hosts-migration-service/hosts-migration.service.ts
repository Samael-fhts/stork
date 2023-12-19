import { Injectable } from '@angular/core'
import { Observable, throwError } from 'rxjs'

/**
 * The migration structure.
 */
export interface Migration {
    // The progress in migration, from 0 to 1.
    progress: number
    // Count of errors encountered during migration.
    errors: number
    // Indicates if the migration is in progress.
    inProgress: boolean
    // The host reservation filter related to the migration.
    filter: string
}

/**
 * The migration service to interact with the backend.
 */
@Injectable({
    providedIn: 'root',
})
export class HostsMigrationService {
    /**
     * Checks the status of the current migration. It's empty if there is no
     * migration in progress.
     * @returns An observable emitting the current migration status.
     */
    getCurrentMigration(): Observable<Migration> {
        return throwError(() => new Error('Not implemented'))
    }

    /**
     * Starts a new migration.
     * @param filter The host reservation filter to apply during migration.
     * @returns An observable emitting the migration status.
     */
    startMigration(filter: string): Observable<Migration> {
        return throwError(() => new Error('Not implemented'))
    }

    /**
     * Stops and removes the current migration.
     * @returns An observable emitting nothing.
     */
    removeMigration(): Observable<void> {
        return throwError(() => new Error('Not implemented'))
    }

    /**
     * Returns a stream of migration updates.
     * @returns An observable emitting migration updates.
     */
    getMigrationUpdates(): Observable<Migration> {
        return throwError(() => new Error('Not implemented'))
    }
}
