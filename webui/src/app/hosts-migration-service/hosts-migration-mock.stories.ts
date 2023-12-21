import { Observable, concatMap, delay, generate, ignoreElements, merge, of, throwError, timer } from 'rxjs'
import { HostsMigrationService, Migration } from './hosts-migration.service'

/**
 * A mock of the HostsMigrationService. It replaces the real service in the
 * component's stories.
 * We use the `Partial` type to keep consistency with the real service.
 * Note, the mock class isn't annotated with `@Injectable` decorator. It don't
 * know it is necessary; it works without it.
 */
export class MockHostsMigrationService implements Partial<HostsMigrationService> {
    private startCount = 0

    /**
     * There is no migration in progress.
     * Delayed by 5s to simulate a backend call.
     */
    getCurrentMigration(): Observable<Migration> {
        return of(null as Migration).pipe(delay(5000))
    }

    /**
     * Returns a stream of migration updates.
     * The updates are generated every 250ms.
     */
    getMigrationUpdates(): Observable<Migration> {
        const progress$ = generate({
            initialState: 0,
            condition: (i) => i <= 100,
            iterate: (i) => i + 1,
            resultSelector: (i) =>
                ({
                    progress: i / 100,
                    errors: Math.round(i / 10),
                    inProgress: i !== 100,
                    filter: `filter-${i}`,
                }) as Migration,
        })

        const bound$ = timer(250).pipe(ignoreElements())
        return progress$.pipe(concatMap((v) => merge(of(v), bound$)))
    }

    /**
     * Delayed by 2s to simulate a backend call.
     */
    removeMigration(): Observable<void> {
        return of(null).pipe(delay(2000))
    }

    /**
     * Returns an initial migration status after 5s.
     * Fails every 3rd call.
     */
    startMigration(filter: string): Observable<Migration> {
        this.startCount += 1
        if (this.startCount % 3 === 0) {
            return throwError(() => new Error('Could not start the migration')).pipe(delay(2000))
        }
        return of({
            id: this.startCount,
            errors: 0,
            inProgress: true,
            progress: 0,
            filter: filter,
        } as Migration).pipe(delay(5000))
    }
}
