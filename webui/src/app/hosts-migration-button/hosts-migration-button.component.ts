import { Component, OnDestroy, OnInit } from '@angular/core'
import { MenuItem } from 'primeng/api'

interface Migration {
    id: number
    progress: number
    errors: number
    inProgress: boolean
}

type State = 'initializing' | 'ready' | 'migrating' | 'done' | 'error'

@Component({
    selector: 'app-hosts-migration-button',
    templateUrl: './hosts-migration-button.component.html',
    styleUrls: ['./hosts-migration-button.component.sass'],
})
export class HostsMigrationButtonComponent implements OnInit, OnDestroy {
    // Component states.
    state: State
    migration: Migration = null

    private setState(state: State, migration: Migration = null) {
        this.state = state
        this.migration = migration
        this.deregisterFromUpdates()
    }

    private transitionToInitializingState() {
        this.setState('initializing')

        // Check the current migration status.
        this.fetchCurrentMigration()
            .then((migration) => {
                if (migration) {
                    if (migration.inProgress) {
                        this.transitionToMigratingState(migration)
                    } else {
                        this.transitionToDoneState(migration)
                    }
                } else {
                    this.transitionToReadyState()
                }
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    private transitionToReadyState() {
        this.setState('ready')
    }

    private transitionToMigrationRequestedState() {
        this.setState('migrating')

        // Start a new migration.
        this.startMigration()
            .then((migration) => {
                this.transitionToMigratingState(migration)
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    private transitionToMigratingState(migration: Migration) {
        this.setState('migrating', migration)

        // Register for updates
        this.registerForUpdates(this.migration.id)
            .then(() => {
                // Wait for updates.
                // TODO
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    private transitionToDoneState(migration: Migration) {
        this.setState('done', migration)

        // Unregister for updates
        this.deregisterFromUpdates()
    }

    private transitionToErrorState(err: Error) {
        this.setState('error')

        // Generate an error message.
        // TODO
    }

    // Component lifecycle hooks.

    ngOnInit(): void {
        this.transitionToInitializingState()
    }
    ngOnDestroy(): void {
        this.deregisterFromUpdates()
    }

    // UI event handlers.
    onStartMigrationClick() {
        this.transitionToMigrationRequestedState()
    }

    onRedirectToMigrationDetailsClick() {
        this.redirectToMigrationDetails(this.migration.id)
    }

    onRetryOnErrorClick() {
        this.transitionToInitializingState()
    }

    onShowErroredHostsClick() {
        this.emitFilterList('filter', true)
    }

    onShowAffectedHostsClick() {
        this.emitFilterList('filter', false)
    }

    onCancelMigrationClick() {
        this.cancelMigration(this.migration.id)
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    onMarkAsReadClick() {
        this.removeMigration(this.migration.id)
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    // HTTP calls.
    // All below function should be excluded to a dedicated service.
    private async fetchCurrentMigration(): Promise<Migration> {
        // TODO
        return null
    }

    private async startMigration(): Promise<Migration> {
        // TODO
        return null
    }

    private async registerForUpdates(migrationId: number): Promise<void> {
        // TODO
    }

    private deregisterFromUpdates() {
        // TODO
    }

    private async cancelMigration(migrationId: number): Promise<void> {
        // TODO
    }

    private async removeMigration(migrationId: number): Promise<void> {}

    // Event emitters.
    private emitFilterList(filter: unknown, errorsOnly: boolean) {
        // ToDo
    }

    // Helpers.
    private redirectToMigrationDetails(migrationId: number) {
        // ToDo
    }
}
