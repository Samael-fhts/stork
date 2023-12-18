import { Component, OnDestroy, OnInit } from '@angular/core'
import { MessageService } from 'primeng/api'
import { HostsMigrationService } from '../hosts-migration-service/hosts-migration.service'
import { Subscription, lastValueFrom } from 'rxjs'
import { getErrorMessage } from '../utils'

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
    updateSubscription: Subscription = null
    fetchingAPI: boolean = false

    constructor(private messageService: MessageService, private migrationService: HostsMigrationService) {
        // The MenuItem commands must be bound to the component instance.
        this.onStartMigrationClick = this.onStartMigrationClick.bind(this)
        this.onShowErroredHostsClick = this.onShowErroredHostsClick.bind(this)
        this.onShowAffectedHostsClick = this.onShowAffectedHostsClick.bind(this)
        this.onCancelMigrationClick = this.onCancelMigrationClick.bind(this)
        this.onMarkAsReadClick = this.onMarkAsReadClick.bind(this)
    }

    private setState(state: State, migration: Migration = null) {
        this.state = state
        this.migration = migration
        this.deregisterFromUpdates()
    }

    private transitionToInitializingState() {
        this.setState('initializing')
        this.fetchingAPI = true

        // Check the current migration status.
        lastValueFrom(this.migrationService.getCurrentMigration())
            .finally(() => {
                this.fetchingAPI = false
            })
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
        this.setState('migrating', {
            errors: 0,
            id: null,
            inProgress: true,
            progress: 0,
        })
        this.fetchingAPI = true

        // Start a new migration.
        lastValueFrom(this.migrationService.startMigration())
            .finally(() => {
                this.fetchingAPI = false
            })
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
        this.updateSubscription = this.migrationService.getMigrationUpdates(migration.id).subscribe(m => {
            this.migration = m
            if (!m.inProgress) {
                this.transitionToDoneState(m)
            }
        })
    }

    private transitionToDoneState(migration: Migration) {
        this.setState('done', migration)
    }

    private transitionToErrorState(err: Error) {
        this.setState('error')

        // Generate an error message.
        const errorMessage = getErrorMessage(err)
        this.messageService.add({
            severity: 'error',
            summary: 'Migration error',
            detail: errorMessage,
        })
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
        this.deregisterFromUpdates()
        this.fetchingAPI = true
        lastValueFrom(this.migrationService.cancelMigration(this.migration.id))
            .finally(() => { this.fetchingAPI = false })
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    onMarkAsReadClick() {
        this.fetchingAPI = true
        lastValueFrom(this.migrationService.removeMigration(this.migration.id))
            .finally(() => { this.fetchingAPI = false })
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    // Event emitters.
    private emitFilterList(filter: unknown, errorsOnly: boolean) {
        // ToDo
    }

    // Helpers.
    private redirectToMigrationDetails(migrationId: number) {
        // ToDo
    }

    private deregisterFromUpdates() {
        if (this.updateSubscription) {
            this.updateSubscription.unsubscribe()
            this.updateSubscription = null
        }
    }
}
